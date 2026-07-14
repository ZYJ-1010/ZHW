param(
  [string]$DatabaseUrl = "postgres://zhw:zhw@127.0.0.1:55432/zhw_mini?sslmode=disable",
  [switch]$SkipDocker
)

$ErrorActionPreference = "Stop"
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not $SkipDocker) {
  docker compose -f deploy/docker-compose.local.yml up -d postgres
  if ($LASTEXITCODE -ne 0) {
    throw "Docker PostgreSQL did not start. Start Docker Desktop, or rerun with -SkipDocker and pass a working -DatabaseUrl."
  }
}

Write-Host "Waiting for PostgreSQL..."
for ($i = 0; $i -lt 30; $i++) {
  & psql $DatabaseUrl -c "select 1" | Out-Null
  if ($LASTEXITCODE -eq 0) {
    break
  }
  Start-Sleep -Seconds 1
}

if ($LASTEXITCODE -ne 0) {
  throw "PostgreSQL is not ready or credentials are wrong: $DatabaseUrl"
}

Get-ChildItem -Path (Join-Path $root "db\migrations") -Filter "*.sql" |
  Sort-Object Name |
  ForEach-Object {
    Write-Host "Applying migration $($_.Name)"
    & psql $DatabaseUrl -v ON_ERROR_STOP=1 -f $_.FullName
    if ($LASTEXITCODE -ne 0) {
      throw "Migration failed: $($_.FullName)"
    }
  }

@(
  "admin_roles_permissions.sql",
  "system_configs.sql",
  "invite_codes.sql",
  "demo_game_orders.sql",
  "user_system_management_configs.sql",
  "local_test_account.sql",
  "local_flow_accounts.sql"
) | ForEach-Object {
  $seedPath = Join-Path $root "db\seeds\$_"
  if (Test-Path $seedPath) {
    Write-Host "Applying seed $_"
    & psql $DatabaseUrl -v ON_ERROR_STOP=1 -f $seedPath
    if ($LASTEXITCODE -ne 0) {
      throw "Seed failed: $seedPath"
    }
  }
}

Write-Host "Local database is ready."
Write-Host "Test login code: local-expert-guide"
Write-Host "Test invite code: LOCAL-EXPERT-GUIDE"
