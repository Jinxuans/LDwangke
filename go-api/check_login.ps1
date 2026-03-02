# check_login.ps1
# Windows PowerShell diagnostic script for intermittent login 500 errors

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

# Safe yaml value parser (simple key:value, first match)
function Get-YamlVal([string]$text, [string]$key) {
  $escapedKey = [regex]::Escape($key)
  $pattern = '(?m)^\s*' + $escapedKey + '\s*:\s*"?([^"#\r\n]+)"?'
  $m = [regex]::Match($text, $pattern)
  if ($m.Success) { return $m.Groups[1].Value.Trim() }
  return ""
}

# 1) config check
Write-Host "1) Check config file..."
$configPath = "config/config.yaml"
if (-not (Test-Path $configPath)) {
  Fail "Missing config file: $configPath"
  Write-Host "Please create it from config/config.example.yaml"
  exit 1
}
Ok "Found config: $configPath"

$cfgText = Get-Content $configPath -Raw
if ($cfgText -match "database:") { Ok "database section exists" } else { Fail "database section missing" }
if ($cfgText -match "jwt:") { Ok "jwt section exists" } else { Fail "jwt section missing" }

$dbHost = Get-YamlVal $cfgText "host"
$dbPort = Get-YamlVal $cfgText "port"
$dbUser = Get-YamlVal $cfgText "username"
$dbPass = Get-YamlVal $cfgText "password"
$dbName = Get-YamlVal $cfgText "database"
$port = Get-YamlVal $cfgText "port"
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

# 4) process/port
Write-Host "4) Check process and listen port..."
$proc = Get-Process -ErrorAction SilentlyContinue | Where-Object { $_.ProcessName -match "go-api|server" } | Select-Object -First 1
if ($proc) { Ok "process found: $($proc.ProcessName) pid=$($proc.Id)" } else { Warn "no go-api/server process found" }

try {
  $tcp = Get-NetTCPConnection -State Listen -ErrorAction Stop | Where-Object { $_.LocalPort -eq [int]$port } | Select-Object -First 1
  if ($tcp) { Ok "port $port is listening" } else { Warn "port $port is not listening" }
} catch {
  Warn "cannot query listen ports: $($_.Exception.Message)"
}

Write-Host ""

# 5) health endpoint
Write-Host "5) Check /health endpoint..."
try {
  $resp = Invoke-WebRequest -Uri ("http://127.0.0.1:{0}/health" -f $port) -UseBasicParsing -TimeoutSec 5
  if ($resp.StatusCode -eq 200) { Ok "/health returned 200" } else { Warn "/health status: $($resp.StatusCode)" }
} catch {
  Fail "Cannot access /health: $($_.Exception.Message)"
}

Write-Host ""
Write-Host "=========================================="
Write-Host "Diagnostic finished"
Write-Host "=========================================="
Write-Host "Next: send me the full output, I will give exact SQL/config fix."
