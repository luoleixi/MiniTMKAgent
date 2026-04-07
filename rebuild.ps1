# 重新编译脚本 - 生成 Windows 原生版本
# 用法: 在 PowerShell 中运行: .\rebuild.ps1

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   MiniTMK Agent - 重新编译" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 Go 是否安装
try {
    $goVersion = go version
    Write-Host "Go 版本: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "[错误] 未找到 Go，请先安装 Go" -ForegroundColor Red
    Write-Host "下载地址: https://go.dev/dl/"
    pause
    exit 1
}

Write-Host ""
Write-Host "[1/3] 清理旧文件..."
if (Test-Path "mini-tmk-agent.exe") {
    Remove-Item "mini-tmk-agent.exe" -Force
    Write-Host "  已删除旧文件"
}

Write-Host ""
Write-Host "[2/3] 编译 Windows 原生版本..."
go build -ldflags="-s -w" -o mini-tmk-agent.exe .

if (-not (Test-Path "mini-tmk-agent.exe")) {
    Write-Host "[错误] 编译失败" -ForegroundColor Red
    pause
    exit 1
}

$size = (Get-Item "mini-tmk-agent.exe").Length / 1MB
Write-Host "  编译成功!" -ForegroundColor Green
Write-Host "  文件大小: $([math]::Round($size, 2)) MB"

Write-Host ""
Write-Host "[3/3] 复制到 release 目录..."
if (-not (Test-Path "release")) {
    New-Item -ItemType Directory -Path "release" | Out-Null
}
Copy-Item "mini-tmk-agent.exe" "release\mini-tmk-agent-windows-amd64.exe" -Force

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "   编译完成!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "在终端中运行: .\mini-tmk-agent.exe --help"
Write-Host ""
pause
