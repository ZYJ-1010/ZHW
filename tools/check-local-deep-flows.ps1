param(
  [int]$Port = 18081,
  [string]$DatabaseUrl = "postgres://zhw:zhw@127.0.0.1:55432/zhw_mini?sslmode=disable",
  [switch]$KeepServer
)

$ErrorActionPreference = "Stop"
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$Root = Split-Path -Parent $PSScriptRoot
$BaseUrl = "http://127.0.0.1:$Port"
$Results = New-Object System.Collections.Generic.List[object]
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
  param([string]$Module, [string]$Status, [string]$Detail)
  $Results.Add([pscustomobject]@{ module = $Module; status = $Status; detail = $Detail }) | Out-Null
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
    [string]$Token = ""
  )

  $headers = @{}
  if ($Token) {
    $headers.Authorization = "Bearer $Token"
  }
  $args = @{
    Method = $Method
    Uri = "$BaseUrl$Path"
    UseBasicParsing = $true
    Headers = $headers
    ErrorAction = "Stop"
  }
  if ($null -ne $Body) {
    $args.ContentType = "application/json; charset=utf-8"
    $args.Body = ($Body | ConvertTo-Json -Depth 30 -Compress)
  }

  try {
    $response = Invoke-WebRequest @args
    $json = $null
    if ($response.Content) {
      $json = $response.Content | ConvertFrom-Json
    }
    return [pscustomobject]@{ ok = $true; statusCode = [int]$response.StatusCode; json = $json; raw = $response.Content }
  } catch {
    return [pscustomobject]@{ ok = $false; statusCode = if ($_.Exception.Response) { [int]$_.Exception.Response.StatusCode } else { 0 }; json = $null; raw = (Read-ErrorBody $_) }
  }
}

function Assert-Ok {
  param($Response, [string]$Label)
  if (-not $Response.ok) {
    throw "$Label failed: HTTP $($Response.statusCode) $($Response.raw)"
  }
  if ($Response.json -and ($null -ne $Response.json.code) -and ([int]$Response.json.code -ne 0)) {
    throw "$Label failed: code=$($Response.json.code) message=$($Response.json.message)"
  }
  return $Response
}

function Assert-Fails {
  param($Response, [string]$Label)
  if ($Response.ok -and ($null -eq $Response.json.code -or [int]$Response.json.code -eq 0)) {
    throw "$Label should fail but succeeded"
  }
  return $Response
}

function Check-Module {
  param([string]$Name, [scriptblock]$Body)
  try {
    & $Body
    Add-Result $Name "PASS" "frontend-api to backend to database flow ok"
  } catch {
    Add-Result $Name "FAIL" $_.Exception.Message
  }
}

function Invoke-Sql {
  param([string]$Sql)
  $psql = (Get-Command psql.exe -ErrorAction SilentlyContinue).Source
  if (-not $psql) {
    throw "psql.exe not found"
  }
  & $psql $DatabaseUrl -tAc $Sql
  if ($LASTEXITCODE -ne 0) {
    throw "sql failed: $Sql"
  }
}

function Apply-LocalSeed {
  $psql = (Get-Command psql.exe -ErrorAction SilentlyContinue).Source
  if (-not $psql) {
    throw "psql.exe not found"
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
    throw "port $Port is already in use"
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
    $probe = Invoke-Api -Method "POST" -Path "/api/app/invites/precheck" -Body @{ inviteCode = "LOCAL-EXPERT-GUIDE"; entryType = "link" }
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
  throw "local server did not become ready: $jobOutput"
}

function Login-App {
  param([string]$InviteCode, [string]$LoginCode)
  Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/invites/precheck" -Body @{ inviteCode = $InviteCode; entryType = "link" }) "invite precheck $InviteCode" | Out-Null
  $login = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/auth/wechat-login" -Body @{ code = $LoginCode; inviteCode = $InviteCode; entryType = "link" }) "wechat login $LoginCode"
  if ($login.json.data.token) {
    return $login.json.data.token
  }
  $issued = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/auth/issue-token-after-identity" -Token $login.json.data.preAuthToken -Body @{}) "issue token $LoginCode"
  return $issued.json.data.token
}

function Login-Admin {
  $login = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/admin/auth/login" -Body @{ username = "admin"; password = "admin123" }) "admin login"
  return $login.json.data.token
}

function Create-FreeGame {
  param([string]$Token, [string]$Title)
  $created = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games" -Token $Token -Body @{
    title = $Title
    gameType = "free"
    type = "free"
    description = "local deep flow game"
    minPlayers = 5
    maxPlayers = 8
    cityCode = "330100"
    cityName = "Hangzhou"
    address = "local deep flow address"
    longitude = 120.1551
    latitude = 30.2741
    price = 0
    primaryCategory = "social"
    primaryCategoryText = "social"
    secondaryCategory = "coffee"
    secondaryCategoryText = "coffee"
    tags = @("local-deep-flow")
    completionRules = @("service_confirm")
  }) "create free game $Title"
  return [int64]$created.json.data.id
}

function Approve-Game {
  param([string]$AdminToken, [int64]$GameID)
  Assert-Ok (Invoke-Api -Method "POST" -Path "/api/admin/games/$GameID/audit" -Token $AdminToken -Body @{ approve = $true; remark = "local deep flow approve" }) "admin approve game $GameID" | Out-Null
}

function Apply-And-Approve {
  param([string]$PlayerToken, [string]$OwnerToken, [int64]$GameID, [string]$Reason)
  $apply = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/$GameID/applications" -Token $PlayerToken -Body @{ reason = $Reason; fileIds = @() }) "apply game $GameID"
  $appID = [int64]$apply.json.data.id
  Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/game-applications/$appID/audit" -Token $OwnerToken -Body @{ approve = $true }) "approve app $appID" | Out-Null
  return $appID
}

function Submit-AllPairReviews {
  param([int64]$GameID, [object[]]$Members)
  foreach ($reviewer in $Members) {
    foreach ($target in $Members) {
      if ($reviewer.userId -eq $target.userId) {
        continue
      }
      $score = 5
      $again = "yes"
      if ($reviewer.userId -eq 10002 -and $target.userId -eq 10001) {
        $score = 1
        $again = "no"
      }
      Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/reviews" -Token $reviewer.token -Body @{
        gameId = $GameID
        targetUserId = $target.userId
        targetRole = "member"
        score = $score
        content = "local deep flow review"
        tags = @("local")
        againIntent = $again
      }) "submit review $($reviewer.userId) to $($target.userId)" | Out-Null
    }
  }
}

try {
  Apply-LocalSeed
  Start-LocalServer

  $ownerToken = Login-App -InviteCode "LOCAL-EXPERT-GUIDE" -LoginCode "local-expert-guide"
  $adminToken = Login-Admin
  $players = @()
  for ($i = 1; $i -le 8; $i++) {
    $players += [pscustomobject]@{
      userId = 10001 + $i
      token = (Login-App -InviteCode "LOCAL-PLAYER-$i" -LoginCode "local-player-$i")
    }
  }

  Check-Module "game-audit-and-join-to-five" {
    $gameID = Create-FreeGame -Token $ownerToken -Title ("local-deep-five-" + (Get-Date -Format "HHmmss"))
    Approve-Game -AdminToken $adminToken -GameID $gameID
    Assert-Fails (Invoke-Api -Method "POST" -Path "/api/app/games/$gameID/manual-start" -Token $ownerToken -Body @{}) "manual start before 5" | Out-Null

    $firstApply = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/$gameID/applications" -Token $players[0].token -Body @{ reason = "first apply"; fileIds = @() }) "first apply"
    Assert-Fails (Invoke-Api -Method "POST" -Path "/api/app/games/$gameID/applications" -Token $players[0].token -Body @{ reason = "duplicate apply"; fileIds = @() }) "duplicate apply" | Out-Null
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/game-applications/$($firstApply.json.data.id)/audit" -Token $ownerToken -Body @{ approve = $true }) "approve first apply" | Out-Null
    for ($i = 1; $i -le 3; $i++) {
      Apply-And-Approve -PlayerToken $players[$i].token -OwnerToken $ownerToken -GameID $gameID -Reason "join to five" | Out-Null
    }
    $members = Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/$gameID/members" -Token $ownerToken) "members after five"
    if (@($members.json.data.items).Count -ne 5) {
      throw "expected 5 members, got $(@($members.json.data.items).Count)"
    }
    $started = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/$gameID/manual-start" -Token $ownerToken -Body @{}) "manual start at five"
    if ($started.json.data.status -ne "in_progress") {
      throw "expected in_progress after manual start, got $($started.json.data.status)"
    }
    $script:FiveGameID = $gameID
  }

  Check-Module "game-eight-lock-rule" {
    $gameID = Create-FreeGame -Token $ownerToken -Title ("local-deep-eight-" + (Get-Date -Format "HHmmss"))
    Approve-Game -AdminToken $adminToken -GameID $gameID
    for ($i = 0; $i -le 6; $i++) {
      Apply-And-Approve -PlayerToken $players[$i].token -OwnerToken $ownerToken -GameID $gameID -Reason "join to eight" | Out-Null
    }
    $detail = Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/$gameID" -Token $ownerToken) "detail after full"
    if ([int]$detail.json.data.currentPlayers -ne 8 -or $detail.json.data.status -ne "full") {
      throw "expected 8 full, got players=$($detail.json.data.currentPlayers) status=$($detail.json.data.status)"
    }
    Assert-Fails (Invoke-Api -Method "POST" -Path "/api/app/games/$gameID/applications" -Token $players[7].token -Body @{ reason = "ninth member"; fileIds = @() }) "apply when full" | Out-Null
  }

  Check-Module "im-service-review-revenue-report" {
    $gameID = [int64]$script:FiveGameID
    $members = @(
      [pscustomobject]@{ userId = 10001; token = $ownerToken },
      [pscustomobject]@{ userId = 10002; token = $players[0].token },
      [pscustomobject]@{ userId = 10003; token = $players[1].token },
      [pscustomobject]@{ userId = 10004; token = $players[2].token },
      [pscustomobject]@{ userId = 10005; token = $players[3].token }
    )

    $room = Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/games/$gameID/chat-room" -Token $ownerToken) "owner chat room"
    $roomID = [int64]$room.json.data.id
    Assert-Fails (Invoke-Api -Method "GET" -Path "/api/app/games/$gameID/chat-room" -Token $players[7].token) "non-member chat room" | Out-Null
    $textMsg = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/$gameID/chat/messages" -Token $players[0].token -Body @{ messageType = "text"; content = "local deep flow text" }) "send text"
    $upload = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/files/upload-token" -Token $players[0].token -Body @{ bizType = "chat_file"; objectId = $gameID; fileName = "local-flow.txt"; mimeType = "text/plain"; size = 12 }) "chat file token"
    $fileID = [int64]$upload.json.data.upload.fileId
    if (-not $fileID) {
      $fileID = [int64]$upload.json.data.file.fileId
    }
    if (-not $fileID) {
      throw "upload token did not return file id"
    }
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/$gameID/chat/messages" -Token $players[0].token -Body @{ messageType = "file"; content = "local file"; fileId = $fileID }) "send file message" | Out-Null
    $adminRoom = Assert-Ok (Invoke-Api -Method "GET" -Path "/api/admin/im/rooms/$roomID" -Token $adminToken) "admin room detail"
    if ([int]$adminRoom.json.data.messageCount -lt 2 -or [int]$adminRoom.json.data.fileMessageCount -lt 1) {
      throw "admin im room missing message or file count"
    }

    foreach ($member in $members) {
      Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/games/$gameID/service-confirm" -Token $member.token -Body @{ note = "confirmed"; fileIds = @() }) "service confirm $($member.userId)" | Out-Null
    }
    $available = Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/reviews/available" -Token $players[0].token) "reviews available"
    if (@($available.json.data.items).Count -eq 0) {
      throw "expected review todos"
    }
    Submit-AllPairReviews -GameID $gameID -Members $members

    $credit = Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/profile/credit-center" -Token $ownerToken) "credit after low review"
    $creditLogs = @($credit.json.data.records)
    if (-not ($creditLogs | Where-Object { $_.reason -eq "low_review" })) {
      throw "low score review did not create credit log"
    }
    $creditLogID = [int64]($creditLogs | Where-Object { $_.reason -eq "low_review" } | Select-Object -First 1).creditLogId
    $appeal = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/profile/credit-appeals" -Token $ownerToken -Body @{ creditLogId = $creditLogID; content = "local appeal"; fileIds = @() }) "credit appeal"
    $appealID = [int64]$appeal.json.data.id
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/admin/reports/batch-handle" -Token $adminToken -Body @{ reportIds = @($appealID); action = "handle"; adminId = 1; result = "approved"; outcome = "appeal_approved"; creditDeduct = 5 }) "handle credit appeal" | Out-Null

    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/admin/revenue/templates" -Token $adminToken) "admin revenue templates" | Out-Null
    $revenue = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/admin/revenue/records/generate" -Token $adminToken -Body @{ gameId = $gameID; amountCent = 10000; templateId = 10001 }) "generate revenue"
    if (-not $revenue.json.data.id) {
      throw "revenue record not generated"
    }
    $income = Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/income/summary" -Token $ownerToken) "income summary"
    if ([int64]$income.json.data.totalCent -le 0) {
      throw "income summary did not increase"
    }

    $report = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/reports" -Token $players[0].token -Body @{ gameId = $gameID; targetUserId = 10001; reportType = "service_dispute"; content = "local report"; chatMessageId = $textMsg.json.data.id }) "create report"
    Assert-Ok (Invoke-Api -Method "POST" -Path "/api/admin/reports/batch-handle" -Token $adminToken -Body @{ reportIds = @([int64]$report.json.data.id); action = "handle"; adminId = 1; result = "confirmed"; outcome = "confirmed"; rewardPoints = 10; creditDeduct = 10 }) "handle report" | Out-Null
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/notifications" -Token $players[0].token) "report notification" | Out-Null
  }

  Check-Module "asset-order-detail-cancel" {
    $items = Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/redemption/items" -Token $ownerToken) "redemption items"
    $item = @($items.json.data.items) | Where-Object { $_.id -eq 10001 } | Select-Object -First 1
    if (-not $item) {
      throw "local redemption item not found"
    }
    $order = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/redemption/orders" -Token $ownerToken -Body @{ itemId = 10001 }) "create redemption order"
    $orderID = [int64]$order.json.data.id
    Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/redemption/orders/$orderID" -Token $ownerToken) "order detail" | Out-Null
    $cancel = Assert-Ok (Invoke-Api -Method "POST" -Path "/api/app/redemption/orders/$orderID/cancel" -Token $ownerToken -Body @{ reason = "local cancel" }) "cancel order"
    if ($cancel.json.data.order.status -ne "canceled") {
      throw "expected canceled order, got $($cancel.json.data.order.status)"
    }
    $orders = Assert-Ok (Invoke-Api -Method "GET" -Path "/api/app/redemption/orders/my" -Token $ownerToken) "orders after cancel"
    if (-not (@($orders.json.data.items) | Where-Object { $_.id -eq $orderID -and $_.status -eq "canceled" })) {
      throw "canceled order not found in order list"
    }
  }
} finally {
  if ($ServerJob -and -not $KeepServer) {
    Stop-Job $ServerJob -ErrorAction SilentlyContinue
    Remove-Job $ServerJob -Force -ErrorAction SilentlyContinue
  }
}

$Results | Format-Table -AutoSize
$failed = @($Results | Where-Object { $_.status -ne "PASS" })
Write-Host "deep_feature_flows=$($Results.Count)"
Write-Host "failed_deep_feature_flows=$($failed.Count)"
if ($failed.Count -gt 0) {
  exit 1
}
