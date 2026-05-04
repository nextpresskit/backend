# NextPressKit dev CLI for Windows (PowerShell). From repo root:
#   .\scripts\nextpresskit.ps1 setup
#   .\scripts\nextpresskit.ps1 run
# Requires: Go on PATH, PostgreSQL for migrate/seed.

$ErrorActionPreference = "Stop"
$RootDir = Split-Path -Parent $PSScriptRoot
Set-Location $RootDir
$env:CGO_ENABLED = if ($env:CGO_ENABLED) { $env:CGO_ENABLED } else { "0" }

function Get-AppPort {
    $envFile = Join-Path $RootDir ".env"
    if (-not (Test-Path $envFile)) { return "9090" }
    foreach ($line in Get-Content $envFile) {
        if ($line -match '^\s*APP_PORT\s*=\s*(.+)$') {
            return $matches[1].Trim()
        }
    }
    "9090"
}

function Get-DevRuntimeBasename {
    $envFile = Join-Path $RootDir ".env"
    if (-not (Test-Path $envFile)) { return "nextpresskit" }
    foreach ($line in Get-Content $envFile) {
        if ($line -match '^\s*APP_DEV_RUNTIME_BASENAME\s*=\s*(.+)$') {
            $b = $matches[1].Trim()
            if (-not [string]::IsNullOrWhiteSpace($b)) { return $b }
        }
    }
    "nextpresskit"
}

function Test-PortListen([int]$Port) {
    $c = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
    return $null -ne $c
}

function Assert-PortFree([int]$Port) {
    if (Test-PortListen $Port) {
        Write-Host "Port $Port is already in use. Stop the process or change APP_PORT in .env." -ForegroundColor Yellow
        Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue |
            ForEach-Object { Get-Process -Id $_.OwningProcess -ErrorAction SilentlyContinue }
        exit 1
    }
}

function Get-Version {
    Push-Location $RootDir
    try {
        $out = & git describe --tags --always --dirty 2>$null
        if ($LASTEXITCODE -eq 0 -and $out) { return $out }
    } finally {
        Pop-Location
    }
    return "dev"
}

function Show-Help {
    @"
NextPressKit dev CLI (Windows PowerShell). On Linux/macOS use: ./scripts/nextpresskit

Commands:
  help install deps tidy build build-all setup
  migrate-up migrate-down migrate-version migrate-steps <N> seed
  run start stop deploy checks
  test test-coverage test-integration security-check graphql clean postman-sync

Examples:
  .\scripts\nextpresskit.ps1 setup
  .\scripts\nextpresskit.ps1 run
"@
}

function Ensure-Go {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Error "Go is not installed or not on PATH."
    }
}

$cmd = if ($args.Count -ge 1) { $args[0] } else { "help" }
if ($cmd -eq "-h" -or $cmd -eq "--help") { $cmd = "help" }

switch ($cmd) {
    "help" { Show-Help }
    "install" {
        Ensure-Go
        go mod download
        $envPath = Join-Path $RootDir ".env"
        if (-not (Test-Path $envPath)) {
            Copy-Item (Join-Path $RootDir ".env.example") $envPath
            Write-Host "Created .env from .env.example — edit DB_* and JWT_SECRET."
        } else {
            Write-Host ".env already exists."
        }
    }
    "deps" { Ensure-Go; go mod download }
    "tidy" { Ensure-Go; go mod tidy }
    "build" {
        Ensure-Go
        New-Item -ItemType Directory -Force -Path (Join-Path $RootDir "bin") | Out-Null
        $v = Get-Version
        go build -ldflags "-X main.version=$v" -o (Join-Path $RootDir "bin\server.exe") ./cmd/api
        Write-Host "Built bin\server.exe"
    }
    "build-all" {
        Ensure-Go
        New-Item -ItemType Directory -Force -Path (Join-Path $RootDir "bin") | Out-Null
        $v = Get-Version
        go build -ldflags "-X main.version=$v" -o (Join-Path $RootDir "bin\server.exe") ./cmd/api
        go build -o (Join-Path $RootDir "bin\migrate.exe") ./cmd/migrate
        go build -o (Join-Path $RootDir "bin\seed.exe") ./cmd/seed
        Write-Host "Built bin\server.exe, bin\migrate.exe, bin\seed.exe"
    }
    "setup" {
        & $PSCommandPath install
        & $PSCommandPath build-all
        & $PSCommandPath migrate-up
        & $PSCommandPath seed
        Write-Host "Local HTTPS + nginx automation runs on Linux/macOS (./scripts/nextpresskit setup in Git Bash/WSL). On native Windows use mkcert + proxy manually or WSL."
        Write-Host "Setup complete. Run: .\scripts\nextpresskit.ps1 run"
    }
    "migrate-up" { Ensure-Go; go run ./cmd/migrate -command=up }
    "migrate-down" {
        Write-Error "migrate-down is removed (GORM AutoMigrate). For dev reset use: db-fresh"
    }
    "migrate-version" {
        Write-Host "No migration version row: schema is internal/platform/dbmigrate + cmd/migrate up."
    }
    "migrate-steps" {
        Write-Error "migrate-steps is removed (no numbered SQL migrations)."
    }
    "migrate-drop" {
        Ensure-Go
        Write-Warning "This drops all tables."
        $c = Read-Host "Type y to continue"
        if ($c -ne "y") { exit 1 }
        $env:ALLOW_SCHEMA_DROP = "1"
        go run ./cmd/migrate -command=drop
    }
    "db-fresh" {
        & $PSCommandPath migrate-drop
        & $PSCommandPath migrate-up
    }
    "seed" { Ensure-Go; go run ./cmd/seed }
    "run" {
        Ensure-Go
        if (-not (Test-Path (Join-Path $RootDir ".env"))) {
            Write-Error ".env missing. Run: .\scripts\nextpresskit.ps1 install"
        }
        $port = [int](Get-AppPort)
        Assert-PortFree $port
        go run ./cmd/api
    }
    "start" {
        Ensure-Go
        $rt = Get-DevRuntimeBasename
        $pidFile = Join-Path $RootDir ".tmp\$rt-api.pid"
        $logFile = Join-Path $RootDir ".tmp\$rt-api.log"
        New-Item -ItemType Directory -Force -Path (Split-Path $pidFile) | Out-Null
        if (Test-Path $pidFile) {
            $oldId = Get-Content $pidFile -ErrorAction SilentlyContinue
            if ($oldId -and (Get-Process -Id $oldId -ErrorAction SilentlyContinue)) {
                Write-Host "API already running (pid=$oldId)."
                exit 0
            }
            Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
        }
        $port = [int](Get-AppPort)
        Assert-PortFree $port
        $goExe = (Get-Command go -ErrorAction Stop).Source
        $proc = Start-Process -FilePath $goExe -ArgumentList "run","./cmd/api" -WorkingDirectory $RootDir `
            -WindowStyle Hidden -RedirectStandardOutput $logFile -RedirectStandardError $logFile -PassThru
        Set-Content -Path $pidFile -Value $proc.Id
        Start-Sleep -Seconds 1
        if (Get-Process -Id $proc.Id -ErrorAction SilentlyContinue) {
            Write-Host "API started in background (pid=$($proc.Id))."
            Write-Host "Logs: $logFile"
        } else {
            Write-Host "API failed to start. See $logFile" -ForegroundColor Red
            Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
            exit 1
        }
    }
    "stop" {
        $rt = Get-DevRuntimeBasename
        $pidFile = Join-Path $RootDir ".tmp\$rt-api.pid"
        if (Test-Path $pidFile) {
            $apiPid = Get-Content $pidFile
            $running = Get-Process -Id $apiPid -ErrorAction SilentlyContinue
            if ($running) {
                Stop-Process -Id $apiPid -Force -ErrorAction SilentlyContinue
                Write-Host "API stopped."
            } else {
                Write-Host "Stale pid file removed."
            }
            Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
        } else {
            Write-Host "No .tmp\$rt-api.pid (nothing to stop from start)."
        }
    }
    "deploy" {
        & (Join-Path $RootDir "scripts\deploy.ps1")
    }
    "checks" {
        Ensure-Go
        go test ./...
        go vet ./...
        go test -tags=integration -v ./internal/platform/database
        go run github.com/getkin/kin-openapi/cmd/validate@latest (Join-Path $RootDir "docs\openapi.yaml")
        go run golang.org/x/vuln/cmd/govulncheck@latest ./...
    }
    "test" { Ensure-Go; go test -v ./... }
    "test-coverage" { Ensure-Go; go test -cover ./... }
    "test-integration" { Ensure-Go; go test -tags=integration -v ./internal/platform/database }
    "security-check" { Ensure-Go; go run golang.org/x/vuln/cmd/govulncheck@latest ./... }
    "graphql" { Ensure-Go; go run github.com/99designs/gqlgen generate }
    "postman-sync" {
        $py = Get-Command python3 -ErrorAction SilentlyContinue
        if (-not $py) { $py = Get-Command python -ErrorAction SilentlyContinue }
        if (-not $py) {
            Write-Error "Python 3 is required for postman-sync (install Python or use Git Bash: ./scripts/nextpresskit postman-sync)."
        }
        $extra = @()
        if ($args.Count -gt 1) { $extra = $args[1..($args.Count - 1)] }
        & $py.Source (Join-Path $RootDir "scripts\sync-postman.py") @extra
    }
    "clean" {
        Remove-Item -Recurse -Force (Join-Path $RootDir "bin") -ErrorAction SilentlyContinue
        go clean
    }
    default {
        Write-Host "Unknown command: $cmd" -ForegroundColor Red
        Show-Help
        exit 1
    }
}
