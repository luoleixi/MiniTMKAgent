@echo off
chcp 65001 >nul
title MiniTMK Agent
cd /d "%~dp0"

echo =========================================
echo    MiniTMK Agent
echo =========================================
echo.

if not exist "mini-tmk-agent.exe" (
    echo ERROR: mini-tmk-agent.exe not found
    pause
    exit /b 1
)

echo Starting MiniTMK Agent...
echo.
mini-tmk-agent.exe interactive
echo.
echo =========================================
echo Program exited
echo.
pause