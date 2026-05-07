$ErrorActionPreference = 'Stop'

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
Set-Location $ProjectRoot

if (-not (Get-Command air -ErrorAction SilentlyContinue)) {
    Write-Host 'air 未安装，请先执行: go install github.com/air-verse/air@latest'
    exit 1
}

air -c .air.toml
