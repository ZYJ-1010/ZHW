param(
  [string]$BaseUrl = "http://127.0.0.1:8080",
  [string]$Code = "local-expert-guide",
  [string]$InviteCode = "LOCAL-EXPERT-GUIDE"
)

$ErrorActionPreference = "Stop"
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$body = @{
  code = $Code
  inviteCode = $InviteCode
  entryType = "link"
} | ConvertTo-Json -Compress

$resp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/app/auth/wechat-login" -ContentType "application/json; charset=utf-8" -Body $body

if ($resp.data.token) {
  Write-Host $resp.data.token
  return
}

if (-not $resp.data.preAuthToken) {
  throw "Login did not return token or preAuthToken."
}

$issueResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/app/auth/issue-token-after-identity" -Headers @{
  Authorization = "Bearer $($resp.data.preAuthToken)"
} -ContentType "application/json; charset=utf-8" -Body "{}"

Write-Host $issueResp.data.token
