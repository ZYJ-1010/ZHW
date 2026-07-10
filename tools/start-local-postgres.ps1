param(
  [string]$DataDir = ".local-pg\data",
  [int]$Port = 55432,
  [string]$Database = "zhw_mini",
  [string]$User = "zhw"
)

$ErrorActionPreference = "Stop"
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$root = Split-Path -Parent $PSScriptRoot
$fullDataDir = Join-Path $root $DataDir
$binDir = Split-Path -Parent (Get-Command psql.exe).Source
$initdb = Join-Path $binDir "initdb.exe"
$pgCtl = Join-Path $binDir "pg_ctl.exe"
$postgres = Join-Path $binDir "postgres.exe"
$createdb = Join-Path $binDir "createdb.exe"
$psql = Join-Path $binDir "psql.exe"
$logDir = Join-Path $root "logs"
$logFile = Join-Path $logDir "local-postgres.log"
$errFile = Join-Path $logDir "local-postgres.err.log"

New-Item -ItemType Directory -Force -Path $logDir | Out-Null

if (-not (Test-Path $fullDataDir)) {
  New-Item -ItemType Directory -Force -Path (Split-Path -Parent $fullDataDir) | Out-Null
  & $initdb -D $fullDataDir -U $User -A trust -E UTF8 --locale=C
  if ($LASTEXITCODE -ne 0) {
    throw "initdb failed"
  }
}

$running = $false
& $pgCtl -D $fullDataDir status | Out-Null
if ($LASTEXITCODE -eq 0) {
  $running = $true
}

if (-not $running) {
  $args = @("-D", $fullDataDir, "-p", "$Port")
  Start-Process -FilePath $postgres -ArgumentList $args -WindowStyle Hidden -UseNewEnvironment -RedirectStandardOutput $logFile -RedirectStandardError $errFile
}

for ($i = 0; $i -lt 30; $i++) {
  & $psql "postgres://$User@127.0.0.1:$Port/postgres?sslmode=disable" -c "select 1" | Out-Null
  if ($LASTEXITCODE -eq 0) {
    break
  }
  Start-Sleep -Seconds 1
}

& $psql "postgres://$User@127.0.0.1:$Port/postgres?sslmode=disable" -tAc "select 1 from pg_database where datname = '$Database'" | Out-Null
if ($LASTEXITCODE -ne 0) {
  throw "failed to inspect local postgres databases"
}

$exists = (& $psql "postgres://$User@127.0.0.1:$Port/postgres?sslmode=disable" -tAc "select 1 from pg_database where datname = '$Database'").Trim()
if ($exists -ne "1") {
  & $createdb -h 127.0.0.1 -p $Port -U $User $Database
  if ($LASTEXITCODE -ne 0) {
    throw "createdb failed"
  }
}

$databaseUrl = "postgres://$User@127.0.0.1:$Port/$Database`?sslmode=disable"
Write-Host $databaseUrl
