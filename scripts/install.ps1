# MiniTMK Agent - One-Click Installer for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.ps1 | iex

# Set UTF-8 encoding
$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::InputEncoding = [System.Text.Encoding]::UTF8

$ErrorActionPreference = "Stop"
$ProgressPreference = 'SilentlyContinue'

# Configuration
$Version = if ($env:MINI_TMK_VERSION) { $env:MINI_TMK_VERSION } else { "latest" }
$RepoOwner = "luoleixi"
$RepoName = "MiniTMKAgent"
$BinaryName = "mini-tmk-agent.exe"
$InstallDir = "$env:LOCALAPPDATA\mini-tmk-agent"

# Output functions
function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Success($msg) {
    try {
        Write-Host "[OK] $msg" -ForegroundColor Green
    } catch {
        Write-Output "[OK] $msg"
    }
}
function Write-Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err($msg) { Write-Host "[ERROR] $msg" -ForegroundColor Red }

# Get system architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
    switch ($arch) {
        "amd64" { return "amd64" }
        "x86" { return "386" }
        "arm64" { return "arm64" }
        default {
            Write-Err "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Get latest version
function Get-LatestVersion {
    Write-Info "Fetching latest version..."
    try {
        $apiUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
        $release = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing -TimeoutSec 30
        return $release.tag_name
    } catch {
        Write-Err "Failed to get latest version: $_"
        exit 1
    }
}

# Download binary
function Download-Binary($version, $arch) {
    $assetName = "mini-tmk-agent.exe"
    $downloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$version/$assetName"
    $outputPath = "$env:TEMP\$BinaryName"

    Write-Info "Downloading $assetName ..."

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $outputPath -UseBasicParsing -TimeoutSec 120

        if (-not (Test-Path $outputPath) -or (Get-Item $outputPath).Length -lt 1000) {
            Write-Err "Downloaded file is invalid"
            exit 1
        }

        $size = [math]::Round((Get-Item $outputPath).Length / 1MB, 2)
        Write-Success "Download complete (${size} MB)"
        return $outputPath
    } catch {
        Write-Err "Download failed: $_"
        exit 1
    }
}

# Install binary
function Install-Binary($sourcePath) {
    $targetPath = Join-Path $InstallDir $BinaryName

    # Create installation directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Write-Info "Created directory: $InstallDir"
    }

    # Remove existing file if running
    if (Test-Path $targetPath) {
        try {
            Remove-Item $targetPath -Force -ErrorAction Stop
        } catch {
            Write-Warn "Cannot replace running program, attempting to overwrite..."
        }
    }

    # Copy binary file
    Copy-Item $sourcePath $targetPath -Force
    Remove-Item $sourcePath -Force
    Write-Success "Installed to: $targetPath"

    return $targetPath
}

# Add to user PATH
function Add-ToUserPath {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if ($currentPath -like "*$InstallDir*") {
        Write-Info "Already in PATH"
        return
    }

    Write-Info "Adding to user PATH..."

    $newPath = if ($currentPath.EndsWith(";")) { "$currentPath$InstallDir" } else { "$currentPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")

    Write-Success "Added to PATH"
}

# Main program
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MiniTMK Agent One-Click Installer" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$arch = Get-Architecture
Write-Info "System architecture: $arch"

if ($Version -eq "latest") {
    $Version = Get-LatestVersion
}
Write-Info "Installing version: $Version"

$tempFile = Download-Binary $Version $arch
$installedPath = Install-Binary $tempFile
Add-ToUserPath

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Installation Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Update current session PATH
$env:Path += ";$InstallDir"

Write-Host "Start using:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  mini-tmk-agent" -ForegroundColor White
Write-Host ""
Write-Host "Or:" -ForegroundColor Gray
Write-Host "  $InstallDir\$BinaryName"
Write-Host ""
Write-Host "Get API Key: https://dashscope.console.aliyun.com/"
Write-Host ""
