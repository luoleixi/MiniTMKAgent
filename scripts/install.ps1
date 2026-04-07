# MiniTMK Agent - One-Click Installer for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.ps1 | iex

# 设置 UTF-8 编码
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$PSDefaultParameterValues['*:Encoding'] = 'utf8'

param(
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"
$ProgressPreference = 'SilentlyContinue'

# 配置
$RepoOwner = "luoleixi"
$RepoName = "MiniTMKAgent"
$BinaryName = "mini-tmk-agent.exe"
$InstallDir = "$env:LOCALAPPDATA\mini-tmk-agent"

# 输出函数
function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Success($msg) { Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err($msg) { Write-Host "[ERROR] $msg" -ForegroundColor Red }

# 获取系统架构
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
    switch ($arch) {
        "amd64" { return "amd64" }
        "x86" { return "386" }
        "arm64" { return "arm64" }
        default {
            Write-Err "不支持的架构: $arch"
            exit 1
        }
    }
}

# 获取最新版本
function Get-LatestVersion {
    Write-Info "获取最新版本..."
    try {
        $apiUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
        $release = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing -TimeoutSec 30
        return $release.tag_name
    } catch {
        Write-Err "无法获取最新版本: $_"
        exit 1
    }
}

# 下载二进制文件
function Download-Binary($version, $arch) {
    $assetName = "mini-tmk-agent.exe"
    $downloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$version/$assetName"
    $outputPath = "$env:TEMP\$BinaryName"

    Write-Info "下载 $assetName ..."

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $outputPath -UseBasicParsing -TimeoutSec 120

        if (-not (Test-Path $outputPath) -or (Get-Item $outputPath).Length -lt 1000) {
            Write-Err "下载文件无效"
            exit 1
        }

        $size = [math]::Round((Get-Item $outputPath).Length / 1MB, 2)
        Write-Success "下载完成 (${size} MB)"
        return $outputPath
    } catch {
        Write-Err "下载失败: $_"
        exit 1
    }
}

# 安装二进制文件
function Install-Binary($sourcePath) {
    $targetPath = Join-Path $InstallDir $BinaryName

    # 创建安装目录
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Write-Info "创建目录: $InstallDir"
    }

    # 如果文件正在运行，先尝试删除
    if (Test-Path $targetPath) {
        try {
            Remove-Item $targetPath -Force -ErrorAction Stop
        } catch {
            Write-Warn "无法替换正在运行的程序，尝试覆盖..."
        }
    }

    # 复制二进制文件
    Copy-Item $sourcePath $targetPath -Force
    Remove-Item $sourcePath -Force
    Write-Success "安装到: $targetPath"

    return $targetPath
}

# 添加到用户 PATH
function Add-ToUserPath {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if ($currentPath -like "*$InstallDir*") {
        Write-Info "PATH 中已存在"
        return
    }

    Write-Info "添加到用户 PATH..."

    $newPath = if ($currentPath.EndsWith(";")) { "$currentPath$InstallDir" } else { "$currentPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")

    # 同时更新当前会话的 PATH，使其立即生效
    $env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + $newPath

    Write-Success "已添加到 PATH"
}

# 验证安装
function Verify-Installation($binaryPath) {
    Write-Info "验证安装..."

    try {
        $result = & $binaryPath --version 2>&1
        Write-Success "运行正常"
        return $true
    } catch {
        Write-Warn "无法验证版本，但安装已完成"
        return $false
    }
}

# 主程序
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MiniTMK Agent 一键安装" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$arch = Get-Architecture
Write-Info "系统架构: $arch"

if ($Version -eq "latest") {
    $Version = Get-LatestVersion
}
Write-Info "安装版本: $Version"

$tempFile = Download-Binary $Version $arch
$installedPath = Install-Binary $tempFile
Verify-Installation $installedPath
Add-ToUserPath

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  安装完成!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "立即使用:" -ForegroundColor Yellow
Write-Host "  mini-tmk-agent quickstart"
Write-Host ""
Write-Host "或者完整路径运行:" -ForegroundColor Gray
Write-Host "  $InstallDir\$BinaryName quickstart"
Write-Host ""
Write-Host "获取 API Key: https://dashscope.console.aliyun.com/"
Write-Host ""
