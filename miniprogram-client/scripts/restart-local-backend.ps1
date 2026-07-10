param(
    [string]$ProjectRoot = "",
    [string]$ApiDir = "",
    [string]$GoExe = "",
    [string]$RuntimeRoot = "E:\zhw-local-runtime",
    [string]$LogDir = "",
    [string]$EnvFile = "",
    [string]$Addr = "",
    [int]$TimeoutSec = 45,
    [switch]$SkipStop,
    [switch]$NoWait
)

$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

function Resolve-FullPath {
    param([string]$Path)
    if ([string]::IsNullOrWhiteSpace($Path)) {
        return ""
    }
    return [System.IO.Path]::GetFullPath($Path)
}

function Load-EnvFile {
    param([string]$Path)
    if ([string]::IsNullOrWhiteSpace($Path) -or !(Test-Path -LiteralPath $Path)) {
        return
    }
    Get-Content -Encoding UTF8 -LiteralPath $Path | ForEach-Object {
        $line = $_.Trim()
        if ($line -eq "" -or $line.StartsWith("#")) {
            return
        }
        $index = $line.IndexOf("=")
        if ($index -le 0) {
            return
        }
        $key = $line.Substring(0, $index).Trim()
        $value = $line.Substring($index + 1).Trim()
        if (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'"))) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        if ($key -ne "") {
            Set-Item -Path "Env:$key" -Value $value
        }
    }
    Write-Host "Loaded env file: $Path"
}

function Get-ChildProcesses {
    param([int[]]$ParentIds)
    $all = @(Get-CimInstance Win32_Process)
    $seen = @{}
    $queue = New-Object System.Collections.Queue
    foreach ($id in $ParentIds) {
        $queue.Enqueue($id)
    }
    while ($queue.Count -gt 0) {
        $parentID = [int]$queue.Dequeue()
        foreach ($child in $all | Where-Object { $_.ParentProcessId -eq $parentID }) {
            if (!$seen.ContainsKey($child.ProcessId)) {
                $seen[$child.ProcessId] = $true
                $queue.Enqueue([int]$child.ProcessId)
                $child
            }
        }
    }
}

function Stop-LocalBackend {
    $parents = @(Get-CimInstance Win32_Process | Where-Object {
        ($_.Name -ieq "go.exe" -or $_.Name -ieq "go") -and
        ($_.CommandLine -match "run\s+\.\/cmd\/server" -or $_.CommandLine -match "run\s+\./cmd/server")
    })
    if ($parents.Count -eq 0) {
        Write-Host "No existing go-api go run process found."
        return
    }

    $children = @(Get-ChildProcesses -ParentIds @($parents | ForEach-Object { [int]$_.ProcessId }))
    $targets = @($children + $parents | Sort-Object ProcessId -Unique)
    foreach ($proc in $targets | Sort-Object ParentProcessId -Descending) {
        try {
            Write-Host "Stopping PID $($proc.ProcessId) $($proc.Name)"
            Stop-Process -Id $proc.ProcessId -Force -ErrorAction Stop
        } catch {
            Write-Warning "Failed to stop PID $($proc.ProcessId): $($_.Exception.Message)"
        }
    }
}

function Resolve-ListenPort {
    param([string]$ListenAddr)
    $value = $ListenAddr.Trim()
    if ($value -eq "") {
        return 8080
    }
    if ($value -match ":(\d+)$") {
        return [int]$Matches[1]
    }
    if ($value -match "^(\d+)$") {
        return [int]$Matches[1]
    }
    return 0
}

function Stop-BackendPortOwner {
    param([string]$ListenAddr, [string]$RuntimeRoot)
    $port = Resolve-ListenPort $ListenAddr
    if ($port -le 0) {
        return
    }

    $owners = @(Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue |
        Select-Object -ExpandProperty OwningProcess -Unique)
    if ($owners.Count -eq 0) {
        Write-Host "No process is listening on port $port."
        return
    }

    foreach ($ownerId in $owners) {
        $proc = Get-Process -Id $ownerId -ErrorAction SilentlyContinue
        $cim = Get-CimInstance Win32_Process -Filter "ProcessId = $ownerId" -ErrorAction SilentlyContinue
        $path = if ($proc) { [string]$proc.Path } else { "" }
        $commandLine = if ($cim) { [string]$cim.CommandLine } else { "" }
        $isBackend =
            ($proc -and ($proc.ProcessName -ieq "server" -or $proc.ProcessName -ieq "go")) -or
            ($path -ne "" -and $path.StartsWith($RuntimeRoot, [System.StringComparison]::OrdinalIgnoreCase)) -or
            ($commandLine -match "cmd/server" -or $commandLine -match "cmd\\server")

        if ($isBackend) {
            Write-Host "Stopping port $port owner PID $ownerId $($proc.ProcessName)"
            Stop-Process -Id $ownerId -Force -ErrorAction Stop
        } else {
            Write-Warning "Port $port is owned by PID $ownerId $($proc.ProcessName), not stopping because it does not look like go-api."
        }
    }
}

function Resolve-GoExe {
    param([string]$ExplicitGoExe, [string]$RuntimeRoot)
    if (![string]::IsNullOrWhiteSpace($ExplicitGoExe) -and (Test-Path -LiteralPath $ExplicitGoExe)) {
        return (Resolve-FullPath $ExplicitGoExe)
    }
    $runtimeGo = Join-Path $RuntimeRoot "tools\go\bin\go.exe"
    if (Test-Path -LiteralPath $runtimeGo) {
        return (Resolve-FullPath $runtimeGo)
    }
    $cmd = Get-Command go -ErrorAction SilentlyContinue
    if ($cmd -and $cmd.Source) {
        return $cmd.Source
    }
    throw "Go executable was not found. Pass -GoExe or install go in PATH."
}

function Use-RuntimeGoCache {
    param([string]$RuntimeRoot)
    if ([string]::IsNullOrWhiteSpace($RuntimeRoot) -or !(Test-Path -LiteralPath $RuntimeRoot)) {
        return
    }

    $goPath = Join-Path $RuntimeRoot "go-path"
    if ([string]::IsNullOrWhiteSpace($env:GOPATH)) {
        if (!(Test-Path -LiteralPath $goPath)) {
            New-Item -ItemType Directory -Force -Path $goPath | Out-Null
        }
        $env:GOPATH = $goPath
    }

    $modCache = Join-Path $RuntimeRoot "go-mod-cache"
    if ([string]::IsNullOrWhiteSpace($env:GOMODCACHE) -and (Test-Path -LiteralPath $modCache)) {
        $env:GOMODCACHE = $modCache
    }

    $buildCache = Join-Path $RuntimeRoot "go-build-cache"
    if ([string]::IsNullOrWhiteSpace($env:GOCACHE)) {
        if (!(Test-Path -LiteralPath $buildCache)) {
            New-Item -ItemType Directory -Force -Path $buildCache | Out-Null
        }
        $env:GOCACHE = $buildCache
    }

    $tmpDir = Join-Path $RuntimeRoot "go-tmp"
    if ([string]::IsNullOrWhiteSpace($env:GOTMPDIR)) {
        if (!(Test-Path -LiteralPath $tmpDir)) {
            New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null
        }
        $env:GOTMPDIR = $tmpDir
    }

    if ([string]::IsNullOrWhiteSpace($env:GOPROXY) -and (Test-Path -LiteralPath $modCache)) {
        $env:GOPROXY = "off"
    }
    if ([string]::IsNullOrWhiteSpace($env:GOSUMDB) -and (Test-Path -LiteralPath $modCache)) {
        $env:GOSUMDB = "off"
    }
}

function Convert-AddrToBaseUrl {
    param([string]$ListenAddr)
    $value = $ListenAddr.Trim()
    if ($value -eq "") {
        $value = ":8080"
    }
    if ($value.StartsWith("http://") -or $value.StartsWith("https://")) {
        return $value.TrimEnd("/")
    }
    if ($value.StartsWith(":")) {
        return "http://127.0.0.1$($value)"
    }
    if ($value -match "^0\.0\.0\.0:(\d+)$") {
        return "http://127.0.0.1:$($Matches[1])"
    }
    if ($value -match "^\[::\]:(\d+)$") {
        return "http://127.0.0.1:$($Matches[1])"
    }
    return "http://$value"
}

if ([string]::IsNullOrWhiteSpace($ProjectRoot)) {
    $ProjectRoot = Resolve-FullPath (Join-Path $PSScriptRoot "..")
} else {
    $ProjectRoot = Resolve-FullPath $ProjectRoot
}
if ([string]::IsNullOrWhiteSpace($ApiDir)) {
    $ApiDir = Join-Path $ProjectRoot "services\go-api"
}
$ApiDir = Resolve-FullPath $ApiDir

if (!(Test-Path -LiteralPath (Join-Path $ApiDir "cmd\server"))) {
    throw "go-api cmd/server was not found under $ApiDir"
}

if ([string]::IsNullOrWhiteSpace($EnvFile)) {
    foreach ($candidate in @((Join-Path $ProjectRoot ".env.local"), (Join-Path $ProjectRoot ".env"), (Join-Path $ProjectRoot ".env.example"))) {
        if (Test-Path -LiteralPath $candidate) {
            $EnvFile = $candidate
            break
        }
    }
}
Load-EnvFile -Path $EnvFile

if (![string]::IsNullOrWhiteSpace($Addr)) {
    $env:GO_API_ADDR = $Addr
}
if ([string]::IsNullOrWhiteSpace($env:GO_API_ADDR)) {
    $env:GO_API_ADDR = ":8080"
}
if ([string]::IsNullOrWhiteSpace($env:APP_ENV)) {
    $env:APP_ENV = "local"
}
Use-RuntimeGoCache -RuntimeRoot $RuntimeRoot

if ([string]::IsNullOrWhiteSpace($LogDir)) {
    if (Test-Path -LiteralPath $RuntimeRoot) {
        $LogDir = Join-Path $RuntimeRoot "logs"
    } else {
        $LogDir = Join-Path $ProjectRoot "logs"
    }
}
if (!(Test-Path -LiteralPath $LogDir)) {
    New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
}

$GoExe = Resolve-GoExe -ExplicitGoExe $GoExe -RuntimeRoot $RuntimeRoot

if (!$SkipStop) {
    Stop-LocalBackend
    Stop-BackendPortOwner -ListenAddr $env:GO_API_ADDR -RuntimeRoot $RuntimeRoot
}

$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$stdout = Join-Path $LogDir "go-api-$stamp.out.log"
$stderr = Join-Path $LogDir "go-api-$stamp.err.log"

Write-Host "Starting go-api with $GoExe"
Write-Host "Working directory: $ApiDir"
Write-Host "GO_API_ADDR=$env:GO_API_ADDR"
if (![string]::IsNullOrWhiteSpace($env:GOMODCACHE)) {
    Write-Host "GOMODCACHE=$env:GOMODCACHE"
}
if (![string]::IsNullOrWhiteSpace($env:GOPATH)) {
    Write-Host "GOPATH=$env:GOPATH"
}
if (![string]::IsNullOrWhiteSpace($env:GOCACHE)) {
    Write-Host "GOCACHE=$env:GOCACHE"
}
if (![string]::IsNullOrWhiteSpace($env:GOTMPDIR)) {
    Write-Host "GOTMPDIR=$env:GOTMPDIR"
}
if (![string]::IsNullOrWhiteSpace($env:GOPROXY)) {
    Write-Host "GOPROXY=$env:GOPROXY"
}
if (![string]::IsNullOrWhiteSpace($env:GOSUMDB)) {
    Write-Host "GOSUMDB=$env:GOSUMDB"
}
if (![string]::IsNullOrWhiteSpace($env:DATABASE_URL)) {
    Write-Host "DATABASE_URL is set."
} else {
    Write-Warning "DATABASE_URL is not set. The server will use in-memory stores where repositories need a database."
}

$process = Start-Process -FilePath $GoExe `
    -ArgumentList @("run", "./cmd/server") `
    -WorkingDirectory $ApiDir `
    -RedirectStandardOutput $stdout `
    -RedirectStandardError $stderr `
    -WindowStyle Hidden `
    -PassThru

Write-Host "Started PID $($process.Id)"
Write-Host "stdout: $stdout"
Write-Host "stderr: $stderr"

if ($NoWait) {
    exit 0
}

$baseURL = Convert-AddrToBaseUrl $env:GO_API_ADDR
$healthURLs = @("$baseURL/health", "$baseURL/api/app/health")
$deadline = (Get-Date).AddSeconds($TimeoutSec)
$lastError = ""

while ((Get-Date) -lt $deadline) {
    $process.Refresh()
    if ($process.HasExited) {
        throw "go-api exited early with code $($process.ExitCode). Check $stderr"
    }
    foreach ($url in $healthURLs) {
        try {
            $resp = Invoke-RestMethod -Method Get -Uri $url -TimeoutSec 2
            Write-Host "Health check ok: $url"
            $resp | ConvertTo-Json -Depth 8
            exit 0
        } catch {
            $lastError = $_.Exception.Message
        }
    }
    Start-Sleep -Seconds 1
}

throw "go-api did not become healthy within $TimeoutSec seconds. Last error: $lastError. Check $stderr"
