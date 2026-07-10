@echo off
REM std-agent 安装脚本入口 (Windows cmd)
REM 转发到 install.ps1，因为 GitHub API JSON 解析与 SHA256 校验在纯 cmd 中不可靠
setlocal

where powershell >nul 2>nul
if errorlevel 1 (
  echo [install] PowerShell 不可用，请改用 install.ps1 手动安装。
  exit /b 1
)

set "SCRIPT_DIR=%~dp0"
set "PS1=%SCRIPT_DIR%install.ps1"

if exist "%PS1%" (
  echo [install] 使用本地 install.ps1
  powershell -NoProfile -ExecutionPolicy Bypass -File "%PS1%" %*
) else (
  if "%STD_AGENT_OWNER%"=="" (set "OWNER=StringKe") else (set "OWNER=%STD_AGENT_OWNER%")
  if "%STD_AGENT_REPO%"=="" (set "REPO=std-agent") else (set "REPO=%STD_AGENT_REPO%")
  set "PS_URL=https://raw.githubusercontent.com/%OWNER%/%REPO%/main/install.ps1"
  echo [install] 远程拉取 %PS_URL%
  powershell -NoProfile -ExecutionPolicy Bypass -Command ^
    "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; iex (irm '%PS_URL%')"
)

set "RC=%ERRORLEVEL%"
endlocal & exit /b %RC%
