param(
  [int]$Port = 18080,
  [string]$DatabaseUrl = "postgres://zhw:zhw@127.0.0.1:55432/zhw_mini?sslmode=disable",
  [string]$InviteCode = "LOCAL-EXPERT-GUIDE",
  [string]$LoginCode = "local-expert-guide",
  [switch]$KeepServer
)

$ErrorActionPreference = "Stop"
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$Root = Split-Path -Parent $PSScriptRoot
$BaseUrl = "http://127.0.0.1:$Port"
$Results = New-Object System.Collections.Generic.List[object]
$AppToken = $null
$AdminToken = $null
$CreatedGameId = $null
$ServerJob = $null

function Resolve-GoExe {
  $items = Get-Command go -All -ErrorAction SilentlyContinue
  foreach ($item in $items) {
    if ($item.Source -and (Test-Path $item.Source)) {
      $file = Get-Item $item.Source
      if ($file.Length -gt 0) {
        return $file.FullName
      }
    }
  }
  throw "valid go executable not found"
}

function Add-Result {
  param(
    [string]$Module,
    [string]$Status,
    [string]$Detail
  )

  $Results.Add([pscustomobject]@{
    module = $Module
    status = $Status
    detail = $Detail
  }) | Out-Null
}

function Read-ErrorBody {
  param($ErrorRecord)

  $response = $ErrorRecord.Exception.Response
  if ($null -eq $response) {
    return $ErrorRecord.Exception.Message
  }

  try {
    $stream = $response.GetResponseStream()
    if ($null -eq $stream) {
      return $ErrorRecord.Exception.Message
    }
    $reader = New-Object System.IO.StreamReader($stream, [System.Text.Encoding]::UTF8)
    return $reader.ReadToEnd()
  } catch {
    return $ErrorRecord.Exception.Message
  }
}

function Invoke-Api {
  param(
    [string]$Method,
    [string]$Path,
    $Body = $null,
    [string]$Token = "",
    [switch]$Admin
  )

  $headers = @{}
  if ($Token) {
    $headers.Authorization = "Bearer $Token"
  }

  $uri = "$BaseUrl$Path"
  $args = @{
    Method = $Method
    Uri = $uri
    UseBasicParsing = $true
    Headers = $headers
    ErrorAction = "Stop"
  }

  if ($null -ne $Body) {
    $args.ContentType = "application/json; charset=utf-8"
    $args.Body = ($Body | ConvertTo-Json -Depth 20 -Compress)
  }

  try {
    $response = Invoke-WebRequest @args
    $json = $null
    if ($response.Content) {
      $json = $response.Content | ConvertFrom-Json
    }
    return [pscustomobject]@{
      ok = $true
      statusCode = [int]$response.StatusCode
      json = $json
      raw = $response.Content
    }
  } catch {
    return [pscustomobject]@{
      ok = $false
      statusCode = if ($_.Exception.Response) { [int]$_.Exception.Response.StatusCode } else { 0 }
      json = $null
      raw = (Read-ErrorBody $_)
    }
  }
}

function Assert-Ok {
  param(
    $Response,
    [string]$Label
  )

  if (-not $Response.ok) {
    throw "$Label failed: HTTP $($Response.statusCode) $($Response.raw)"
  }
  if ($Response.json -and ($null -ne $Response.json.code) -and ([int]$Response.json.code -ne 0)) {
    throw "$Label failed: code=$($Response.json.code) message=$($Response.json.message)"
  }
  return $Response
}

function Check-Module {
  param(
    [string]$Name,
    [scriptblock]$Body
  )

  try {
    & $Body
    Add-Result $Name "PASS" "front-api-backend flow ok"
  } catch {
    Add-Result $Name "FAIL" $_.Exception.Message
  }
}

function Test-Database {
  $psql = (Get-Command psql.exe -ErrorAction SilentlyContinue).Source
  if (-not $psql) {
    throw "psql.exe not found; start PostgreSQL first or add PostgreSQL bin to PATH."
  }
  & $psql $DatabaseUrl -tAc "select 1" | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "local database is not reachable: $DatabaseUrl"
  }
}

function Invoke-Sql {
  param([string]$Sql)
  $psql = (Get-Command psql.exe -ErrorAction SilentlyContinue).Source
  if (-not $psql) {
    throw "psql.exe not found; start PostgreSQL first or add PostgreSQL bin to PATH."
  }
  & $psql $DatabaseUrl -tAc $Sql | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "sql failed: $Sql"
  }
}

function Apply-LocalSeed {
  $psql = (Get-Command psql.exe -ErrorAction SilentlyContinue).Source
  if (-not $psql) {
    throw "psql.exe not found; start PostgreSQL first or add PostgreSQL bin to PATH."
  }
  & $psql $DatabaseUrl -v ON_ERROR_STOP=1 -f (Join-Path $Root "db\seeds\local_flow_accounts.sql") | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "local flow seed failed"
  }
  Invoke-Sql "update games set created_at = now() - interval '2 days' where creator_user_id = 10001 and (id = 10001 or title like 'local-%')"
}

function Start-LocalServer {
  $existing = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue
  if ($existing) {
    throw "port $Port is already in use; choose another -Port or stop the old server."
  }

  $goExe = Resolve-GoExe
  $ServerJob = Start-Job -ArgumentList $Root, $Port, $DatabaseUrl, $goExe -ScriptBlock {
    param($RootPath, $ApiPort, $DbUrl, $GoExe)
    $ErrorActionPreference = "Stop"
    $env:APP_ENV = "local"
    $env:GO_API_ADDR = ":$ApiPort"
    $env:DATABASE_URL = $DbUrl
    $env:DATABASE_DRIVER = "postgres"
    $env:JWT_SECRET = "local-dev-jwt-secret"
    $env:STORAGE_PROVIDER = "local"
    $env:STORAGE_UPLOAD_BASE_URL = "http://127.0.0.1:$ApiPort/mock-upload"
    $env:STORAGE_DOWNLOAD_BASE_URL = "http://127.0.0.1:$ApiPort/mock-files"
    $env:GOTELEMETRY = "off"
    $env:GOCACHE = Join-Path $RootPath ".cache\go-build"
    $env:GOTELEMETRYDIR = Join-Path $RootPath ".cache\go-telemetry"
    New-Item -ItemType Directory -Force -Path $env:GOCACHE, $env:GOTELEMETRYDIR | Out-Null
    Set-Location (Join-Path $RootPath "services\go-api")
    & $GoExe run -buildvcs=false ./cmd/server
  }

  for ($i = 0; $i -lt 60; $i++) {
    $probe = Invoke-Api -Method "POST" -Path "/api/app/invites/precheck" -Body @{ inviteCode = $InviteCode; entryType = "link" }
    if ($probe.ok) {
      return
    }
    Start-Sleep -Seconds 1
    if ($ServerJob.State -eq "Failed" -or $ServerJob.State -eq "Completed") {
      $output = Receive-Job $ServerJob -Keep | Out-String
      throw "local server stopped before ready: $output"
    }
  }

  $jobOutput = Receive-Job $ServerJob -Keep | Out-String
  throw "local server did not become ready on $BaseUrl. $jobOutput"
}

function Login-App {
  param(
    [string]$Code = $LoginCode,
    [string]$Invite = $InviteCode
  )

  $precheck = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/invites/precheck" -Body @{ inviteCode = $Invite; entryType = "link" }) "invite precheck"
  if (-not $precheck.json.data.valid) {
    throw "invite precheck returned invalid invite"
  }

  $login = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/auth/wechat-login" -Body @{ code = $Code; inviteCode = $Invite; entryType = "link" }) "wechat login"
  if ($login.json.data.token) {
    return $login.json.data.token
  }
  if (-not $login.json.data.preAuthToken) {
    throw "wechat login did not return token or preAuthToken"
  }

  $issued = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/auth/issue-token-after-identity" -Token $login.json.data.preAuthToken -Body @{}) "issue token after identity"
  return $issued.json.data.token
}

function Login-Admin {
  $login = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/admin/auth/login" -Body @{ username = "admin"; password = "admin123" }) "admin login"
  if (-not $login.json.data.token) {
    throw "admin login did not return token"
  }
  return $login.json.data.token
}

try {
  Test-Database
  Apply-LocalSeed
  Start-LocalServer

  Check-Module "invite-registration-login" {
    $script:AppToken = Login-App
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/users/me" -Token $script:AppToken) "users me" | Out-Null
  }

  Check-Module "identity-realname" {
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/identity/status" -Token $script:AppToken) "identity status" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/users/me" -Token $script:AppToken) "users me identity" | Out-Null
  }

  Check-Module "role-application" {
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/role-applications/expert/config" -Token $script:AppToken) "expert config" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/role-applications/guide/config" -Token $script:AppToken) "guide config" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/role-applications/my" -Token $script:AppToken) "my role applications" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/roles/my" -Token $script:AppToken) "my roles" | Out-Null
  }

  Check-Module "membership" {
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/membership/my" -Token $script:AppToken) "membership my" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/membership/plans" -Token $script:AppToken) "membership plans" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/membership/radar-config" -Token $script:AppToken) "membership radar config" | Out-Null
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/membership/radar/actions" -Token $script:AppToken -Body @{ action = "start"; targetId = "local-flow" }) "membership radar action" | Out-Null
  }

  Check-Module "create-game" {
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/category-config" -Token $script:AppToken) "game category config" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/condition-rule-config" -Token $script:AppToken) "condition rule config" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/profit-templates" -Token $script:AppToken) "profit templates" | Out-Null
    $title = "local-flow-free-game-" + (Get-Date -Format "HHmmss")
    $created = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games" -Token $script:AppToken -Body @{
      title = $title
      gameType = "free"
      type = "free"
      description = "local frontend-backend flow free game"
      minPlayers = 5
      maxPlayers = 8
      cityCode = "330100"
      cityName = "Hangzhou"
      address = "local test address"
      longitude = 120.1551
      latitude = 30.2741
      price = 0
      primaryCategory = "social"
      primaryCategoryText = "social"
      secondaryCategory = "coffee"
      secondaryCategoryText = "coffee"
      tags = @("local-flow")
      completionRules = @("checkin")
    }) "create free game"
    $script:CreatedGameId = $created.json.data.id
    if (-not $script:CreatedGameId) {
      throw "create free game did not return id"
    }
    if (-not $script:AdminToken) {
      $script:AdminToken = Login-Admin
    }
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/admin/games/$script:CreatedGameId/audit" -Token $script:AdminToken -Body @{ approve = $true; remark = "local feature flow approve" }) "admin approve created game" | Out-Null
  }

  Check-Module "home-browse-game-detail" {
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/home" -Token $script:AppToken) "home" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games" -Token $script:AppToken) "games list" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/10001" -Token $script:AppToken) "seed game detail" | Out-Null
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/locations/current" -Token $script:AppToken -Body @{ longitude = 120.1551; latitude = 30.2741; accuracyMeter = 80; cityCode = "330100"; cityName = "Hangzhou" }) "save current location" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/nearby?longitude=120.1551&latitude=30.2741" -Token $script:AppToken) "nearby games" | Out-Null
  }

  Check-Module "im-and-files" {
    $gameId = if ($script:CreatedGameId) { $script:CreatedGameId } else { 10001 }
    for ($i = 1; $i -le 4; $i++) {
      $playerToken = Login-App -Code "local-player-$i" -Invite "LOCAL-PLAYER-$i"
      $apply = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/$gameId/applications" -Token $playerToken -Body @{ reason = "local feature flow join"; fileIds = @() }) "player $i apply game"
      Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/game-applications/$($apply.json.data.id)/audit" -Token $script:AppToken -Body @{ approve = $true }) "owner approve player $i" | Out-Null
    }
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/$gameId/manual-start" -Token $script:AppToken -Body @{}) "manual start game" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/$gameId/chat-room" -Token $script:AppToken) "chat room" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/$gameId/chat/messages" -Token $script:AppToken) "chat messages" | Out-Null
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/$gameId/chat/messages" -Token $script:AppToken -Body @{ messageType = "text"; content = "local frontend backend flow message" }) "send text message" | Out-Null
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/files/upload-token" -Token $script:AppToken -Body @{ bizType = "chat_file"; objectId = $gameId; fileName = "flow.png"; mimeType = "image/png"; size = 128 }) "file upload token" | Out-Null
  }

  Check-Module "delivery-service-review-entry" {
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/10001/service-confirm" -Token $script:AppToken -Body @{}) "service confirm" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/reviews/available" -Token $script:AppToken) "available reviews" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/reviews/complete-config" -Token $script:AppToken) "review complete config" | Out-Null
  }

  Check-Module "credit-report-appeal" {
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/profile/credit-center" -Token $script:AppToken) "credit center" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/reports/config" -Token $script:AppToken) "reports config" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/reports/my" -Token $script:AppToken) "my reports" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/reports/appeals/my" -Token $script:AppToken) "my appeals" | Out-Null
  }

  Check-Module "message-notification-actions" {
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/notifications" -Token $script:AppToken) "notifications" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/messages/trade-warning" -Token $script:AppToken) "trade warning" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/messages/system-notification" -Token $script:AppToken) "system notification" | Out-Null
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/messages/system-notification/feedback" -Token $script:AppToken -Body @{ messageId = "local-flow"; value = "useful" }) "system notification feedback" | Out-Null
  }

  Check-Module "profile-assets-orders" {
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/profile/home" -Token $script:AppToken) "profile home" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/profile/assets" -Token $script:AppToken) "profile assets" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/points/summary" -Token $script:AppToken) "points summary" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/redemption/items" -Token $script:AppToken) "redemption items" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/redemption/orders/my" -Token $script:AppToken) "redemption orders" | Out-Null
  }

  Check-Module "map-basic-play-pages" {
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/map/index-config" -Token $script:AppToken) "map index config" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/map/my-city" -Token $script:AppToken) "map my city" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/map/play-pages" -Token $script:AppToken) "map play pages" | Out-Null
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/map/blind-routes" -Token $script:AppToken -Body @{ cardId = "local-card"; title = "local blind route" }) "create blind route" | Out-Null
  }

  Check-Module "admin-management-config-logs" {
    $script:AdminToken = Login-Admin
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/admin/auth/permissions" -Token $script:AdminToken) "admin permissions" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/admin/games" -Token $script:AdminToken) "admin games" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/admin/games/category-config" -Token $script:AdminToken) "admin game category config" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/admin/games/condition-rule-config" -Token $script:AdminToken) "admin condition rule config" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/admin/operation-logs" -Token $script:AdminToken) "admin operation logs" | Out-Null
  }
} finally {
  if ($ServerJob -and -not $KeepServer) {
    Stop-Job $ServerJob -ErrorAction SilentlyContinue
    Remove-Job $ServerJob -Force -ErrorAction SilentlyContinue
  }
}

$Results | Format-Table -AutoSize
$failed = @($Results | Where-Object { $_.status -ne "PASS" })
Write-Host "feature_http_flows=$($Results.Count)"
Write-Host "failed_feature_http_flows=$($failed.Count)"
if ($failed.Count -gt 0) {
  exit 1
}
