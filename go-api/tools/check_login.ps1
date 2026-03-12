# check_login.ps1
# Windows PowerShell diagnostic script for intermittent login errors
# Usage: .\tools\check_login.ps1

Write-Host "=========================================="
Write-Host "Login Diagnostic Tool (PowerShell)"
Write-Host "=========================================="
Write-Host ""

$GreenColor = "Green"
$RedColor = "Red"
$YellowColor = "Yellow"

function Ok($msg) { Write-Host "[OK] $msg" -ForegroundColor $GreenColor }
function Fail($msg) { Write-Host "[ERR] $msg" -ForegroundColor $RedColor }
function Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor $YellowColor }

function Get-YamlSectionValue([string[]]$lines, [string]$section, [string]$key) {
  $inSection = $false
  foreach ($line in $lines) {
    if ($line -match ('^\s*' + [regex]::Escape($section) + ':\s*$')) {
      $inSection = $true
      continue
    }
    if ($inSection -and $line -match '^\S') {
      break
    }
    if ($inSection -and $line -match ('^\s+' + [regex]::Escape($key) + ':\s*"?([^"#\r\n]+)"?')) {
      return $Matches[1].Trim()
    }
  }
  return ""
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$configPath = Join-Path $projectRoot 'config\config.yaml'

Write-Host "1) Check config file..."
if (-not (Test-Path $configPath)) {
  Fail "Missing config file: $configPath"
  Write-Host "Please create it from config\config.example.yaml"
  exit 1
}
Ok "Found config: $configPath"

$cfgLines = Get-Content $configPath
$cfgText = $cfgLines -join "`n"
if ($cfgText -match "database:") { Ok "database section exists" } else { Fail "database section missing" }
if ($cfgText -match "jwt:") { Ok "jwt section exists" } else { Fail "jwt section missing" }

$dbHost = Get-YamlSectionValue $cfgLines "database" "host"
$dbPort = Get-YamlSectionValue $cfgLines "database" "port"
$dbUser = Get-YamlSectionValue $cfgLines "database" "user"
$dbPass = Get-YamlSectionValue $cfgLines "database" "password"
$dbName = Get-YamlSectionValue $cfgLines "database" "dbname"
$port = Get-YamlSectionValue $cfgLines "server" "port"
if (-not $port) { $port = "8080" }

Write-Host ""

# 2) mysql connectivity
Write-Host "2) Check MySQL connectivity..."
$mysqlCmd = Get-Command mysql -ErrorAction SilentlyContinue
if (-not $mysqlCmd) {
  Warn "mysql client not found, skip MySQL tests"
} else {
  if ($dbHost -and $dbPort -and $dbUser -and $dbName) {
    Write-Host "DB target: $dbHost`:$dbPort / $dbName"
    & mysql "-h$dbHost" "-P$dbPort" "-u$dbUser" "-p$dbPass" -e "USE $dbName; SELECT 1;" 1>$null 2>$null
    if ($LASTEXITCODE -eq 0) { Ok "MySQL connection success" } else { Fail "MySQL connection failed" }
  } else {
    Warn "DB params incomplete from config.yaml"
  }
}

Write-Host ""

# 3) table/schema checks
Write-Host "3) Check tables and key fields..."
if ($mysqlCmd -and $dbHost -and $dbPort -and $dbUser -and $dbName) {
  & mysql "-h$dbHost" "-P$dbPort" "-u$dbUser" "-p$dbPass" "-D$dbName" -e "DESC qingka_wangke_user;" 1>$null 2>$null
  if ($LASTEXITCODE -eq 0) { Ok "table qingka_wangke_user exists" } else { Fail "table qingka_wangke_user missing or no permission" }

  $pass2 = & mysql "-h$dbHost" "-P$dbPort" "-u$dbUser" "-p$dbPass" "-D$dbName" -Nse "SHOW COLUMNS FROM qingka_wangke_user LIKE 'pass2';" 2>$null
  if ($pass2) { Ok "field pass2 exists" } else { Warn "field pass2 missing (compat mode may still work)" }

  & mysql "-h$dbHost" "-P$dbPort" "-u$dbUser" "-p$dbPass" "-D$dbName" -e "DESC qingka_wangke_config;" 1>$null 2>$null
  if ($LASTEXITCODE -eq 0) { Ok "table qingka_wangke_config exists" } else { Fail "table qingka_wangke_config missing or no permission" }

  $pass2kg = & mysql "-h$dbHost" "-P$dbPort" "-u$dbUser" "-p$dbPass" "-D$dbName" -Nse "SELECT k FROM qingka_wangke_config WHERE v='pass2_kg' LIMIT 1;" 2>$null
  if ($pass2kg) { Ok "config pass2_kg exists: $pass2kg" } else { Warn "config pass2_kg missing" }
} else {
  Warn "Skip schema checks (mysql/db params unavailable)"
}

Write-Host ""

Write-Host "4) Check process and listen port..."
$proc = Get-Process -ErrorAction SilentlyContinue | Where-Object { $_.ProcessName -match "go-api|server|go-api-linux" } | Select-Object -First 1
if ($proc) { Ok "process found: $($proc.ProcessName) pid=$($proc.Id)" } else { Warn "no go-api/server process found" }

try {
  $tcp = Get-NetTCPConnection -State Listen -ErrorAction Stop | Where-Object { $_.LocalPort -eq [int]$port } | Select-Object -First 1
  if ($tcp) { Ok "port $port is listening" } else { Warn "port $port is not listening" }
} catch {
  Warn "cannot query listen ports: $($_.Exception.Message)"
}

Write-Host ""

Write-Host "5) Check /api/v1/site/config endpoint..."
try {
  $resp = Invoke-WebRequest -Uri ("http://127.0.0.1:{0}/api/v1/site/config" -f $port) -UseBasicParsing -TimeoutSec 5
  if ($resp.StatusCode -eq 200) { Ok "/api/v1/site/config returned 200" } else { Warn "endpoint status: $($resp.StatusCode)" }
} catch {
  Fail "Cannot access public endpoint: $($_.Exception.Message)"
}

Write-Host ""
Write-Host "=========================================="
Write-Host "Diagnostic finished"
Write-Host "=========================================="
Write-Host "Next: send me the full output, I will give exact SQL/config fix."
