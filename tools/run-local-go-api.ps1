param(
  [string]$DatabaseUrl = "postgres://zhw:zhw@127.0.0.1:55432/zhw_mini?sslmode=disable"
)

$ErrorActionPreference = "Stop"
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$root = Split-Path -Parent $PSScriptRoot
$env:APP_ENV = "local"
$env:GO_API_ADDR = ":8080"
$env:DATABASE_URL = $DatabaseUrl
$env:DATABASE_DRIVER = "postgres"
$env:JWT_SECRET = "local-dev-jwt-secret"
$env:STORAGE_PROVIDER = "local"
$env:STORAGE_UPLOAD_BASE_URL = "http://127.0.0.1:8080/mock-upload"
$env:STORAGE_DOWNLOAD_BASE_URL = "http://127.0.0.1:8080/mock-files"

Set-Location (Join-Path $root "services\go-api")
go run ./cmd/server
