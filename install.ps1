<#
.SYNOPSIS
  std-ai 安装脚本 (Windows PowerShell 5.1+ / PowerShell 7+)

.DESCRIPTION
  用法:
    irm https://raw.githubusercontent.com/StringKe/std-ai/main/install.ps1 | iex

  可选环境变量:
    STD_AI_OWNER       GitHub owner (默认 StringKe)
    STD_AI_REPO        仓库名 (默认 std-ai)
    STD_AI_VERSION     版本 tag (如 v0.1.0；默认 latest)
    STD_AI_INSTALL_DIR 安装目录 (默认 $LOCALAPPDATA\Programs\std-ai)
#>

[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

function Write-Info($msg) { Write-Host "[install] $msg" -ForegroundColor Cyan }
function Write-Warn($msg) { Write-Host "[install] $msg" -ForegroundColor Yellow }
function Write-Err($msg)  { Write-Host "[install] $msg" -ForegroundColor Red }

$Owner       = if ($env:STD_AI_OWNER)       { $env:STD_AI_OWNER }       else { 'StringKe' }
$Repo        = if ($env:STD_AI_REPO)        { $env:STD_AI_REPO }        else { 'std-ai' }
$Version     = if ($env:STD_AI_VERSION)     { $env:STD_AI_VERSION }     else { 'latest' }
$InstallDir  = if ($env:STD_AI_INSTALL_DIR) { $env:STD_AI_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\std-ai' }
$BinName     = 'stdagent.exe'

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { throw "不支持的架构: $($env:PROCESSOR_ARCHITECTURE)" }
}

if ($Version -eq 'latest') {
    Write-Info '解析最新版本'
    $apiUrl = "https://api.github.com/repos/$Owner/$Repo/releases/latest"
    $headers = @{ 'User-Agent' = 'std-ai-install' }
    if ($env:GITHUB_TOKEN) { $headers['Authorization'] = "Bearer $($env:GITHUB_TOKEN)" }
    $Version = (Invoke-RestMethod -Uri $apiUrl -Headers $headers).tag_name
    if (-not $Version) { throw '无法解析最新版本，请显式指定 STD_AI_VERSION' }
}

$verNoV       = $Version.TrimStart('v')
$archive      = "${Repo}_${verNoV}_windows_${arch}.zip"
$downloadBase = "https://github.com/$Owner/$Repo/releases/download/$Version"
$archiveUrl   = "$downloadBase/$archive"
$checksumUrl  = "$downloadBase/checksums.txt"

$tmpRoot = Join-Path $env:TEMP ("std-ai-install-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tmpRoot -Force | Out-Null
try {
    Write-Info "下载 $archiveUrl"
    $archivePath = Join-Path $tmpRoot $archive
    Invoke-WebRequest -Uri $archiveUrl -OutFile $archivePath -UseBasicParsing

    Write-Info '下载并校验 checksum'
    $checksumPath = Join-Path $tmpRoot 'checksums.txt'
    Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath -UseBasicParsing

    $expectedLine = (Get-Content $checksumPath | Where-Object {
        $_ -match ('\s' + [regex]::Escape($archive) + '\s*$')
    }) | Select-Object -First 1
    if (-not $expectedLine) { throw "checksums.txt 中找不到 $archive" }
    $expected = ($expectedLine -split '\s+')[0].ToLower()
    $actual   = (Get-FileHash -Algorithm SHA256 -Path $archivePath).Hash.ToLower()
    if ($expected -ne $actual) {
        throw "checksum 不匹配 expected=$expected actual=$actual"
    }

    Write-Info '解包'
    Expand-Archive -Path $archivePath -DestinationPath $tmpRoot -Force
    $binPath = Join-Path $tmpRoot $BinName
    if (-not (Test-Path $binPath)) { throw "归档中未找到二进制 $BinName" }

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    $target = Join-Path $InstallDir $BinName
    Move-Item -Path $binPath -Destination $target -Force

    Write-Info "已安装 $target ($Version)"

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $segments = if ($userPath) { $userPath.Split(';') } else { @() }
    if ($segments -notcontains $InstallDir) {
        Write-Warn "$InstallDir 未在用户 PATH 中"
        Write-Warn "可执行: [Environment]::SetEnvironmentVariable('Path', '$InstallDir;' + [Environment]::GetEnvironmentVariable('Path','User'), 'User')"
    }
}
finally {
    Remove-Item -Recurse -Force -Path $tmpRoot -ErrorAction SilentlyContinue
}
