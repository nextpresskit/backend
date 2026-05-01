# Interactive deploy wizard (Windows / PowerShell): Nginx snippets, optional TLS, optional build.
# Run from repo root:
#   .\scripts\deploy.ps1
#
# For full Linux server deploy (git, systemd, Certbot), use WSL or SSH and run ./scripts/deploy (bash).

$ErrorActionPreference = "Stop"
$RootDir = Split-Path -Parent $PSScriptRoot
Set-Location $RootDir

function Get-EnvFileValue([string]$Root, [string]$Key) {
    $p = Join-Path $Root ".env"
    if (-not (Test-Path $p)) { return $null }
    foreach ($line in Get-Content $p) {
        $t = $line.Trim()
        if ($t.Length -eq 0 -or $t.StartsWith("#")) { continue }
        $prefix = "$Key="
        if (-not $t.StartsWith($prefix)) { continue }
        $v = $t.Substring($prefix.Length).Trim()
        if ($v.Length -ge 2 -and $v.StartsWith('"') -and $v.EndsWith('"')) {
            return $v.Substring(1, $v.Length - 2)
        }
        if ($v.Length -ge 2 -and $v.StartsWith("'") -and $v.EndsWith("'")) {
            return $v.Substring(1, $v.Length - 2)
        }
        return $v
    }
    return $null
}

function Prompt-Default($Default, [string]$Message) {
    $line = Read-Host "$Message [$Default]"
    if ([string]::IsNullOrWhiteSpace($line)) { return $Default }
    return $line
}

function Prompt-Yn($DefaultY, [string]$Message) {
    while ($true) {
        $def = if ($DefaultY) { "y" } else { "n" }
        $line = Read-Host "$Message (y/n) [$def]"
        if ([string]::IsNullOrWhiteSpace($line)) { $line = $def }
        $c = $line.Substring(0, 1).ToLowerInvariant()
        if ($c -eq "y") { return $true }
        if ($c -eq "n") { return $false }
        Write-Host "Please answer y or n." -ForegroundColor Yellow
    }
}

function Get-Slug([string]$s) {
    $t = $s.ToLowerInvariant() -replace '[^a-z0-9._-]+', '-'
    $t = $t.Trim('-')
    if ($t.Length -gt 48) { $t = $t.Substring(0, 48) }
    if ([string]::IsNullOrWhiteSpace($t)) { return "site" }
    return $t
}

function Tier-FromChoice([string]$choice) {
    $c = $choice.Trim().ToLowerInvariant()
    switch ($c) {
        "1" { return "production" }
        "2" { return "staging" }
        "3" { return "dev" }
        "4" { return "local" }
        "production" { return "production" }
        "staging" { return "staging" }
        "dev" { return "dev" }
        "local" { return "local" }
        Default { return "local" }
    }
}

function Branch-ForTier([string]$tier) {
    switch ($tier) {
        "production" { return "main" }
        "staging" { return "staging" }
        "dev" { return "dev" }
        Default { return "" }
    }
}

$genLabel = Get-EnvFileValue $RootDir "APP_NAME"
if ([string]::IsNullOrWhiteSpace($genLabel)) { $genLabel = "NextPressKit" }

Write-Host "$genLabel — interactive deploy (PowerShell)" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tip: On Linux servers use bash ./scripts/deploy from the repo." -ForegroundColor Yellow
Write-Host ""

Write-Host "Environment tier:"
Write-Host "  1) production  2) staging  3) dev  4) local"
$tierChoice = Prompt-Default "4" "Choose tier (1-4 or name)"
$tier = Tier-FromChoice $tierChoice

$defaultDomain = "api.example.com"
$defaultPort = "9090"
switch ($tier) {
    "production" { $defaultDomain = "api.example.com"; $defaultPort = "9090" }
    "staging" { $defaultDomain = "api-staging.example.com"; $defaultPort = "9091" }
    "dev" { $defaultDomain = "api-dev.example.com"; $defaultPort = "9092" }
    "local" { $defaultDomain = "nextpresskit.local"; $defaultPort = "9090" }
}

$publicHost = Get-EnvFileValue $RootDir "NEXTPRESS_PUBLIC_HOST"
if (-not [string]::IsNullOrWhiteSpace($publicHost)) {
    $defaultDomain = $publicHost
}

$serviceUnit = Get-EnvFileValue $RootDir "APP_SERVICE_UNIT"
if ([string]::IsNullOrWhiteSpace($serviceUnit)) { $serviceUnit = "nextpresskit-backend" }
$sslSubdir = Get-EnvFileValue $RootDir "APP_LOCAL_SSL_SUBDIR"
if ([string]::IsNullOrWhiteSpace($sslSubdir)) { $sslSubdir = "nextpresskit-ssl" }
$sslShare = Join-Path $env:USERPROFILE ".local\share\$sslSubdir"

$genConfig = Prompt-Yn $true "Generate Nginx config files?"
$sslMode = "none"
$certbotEmail = "admin@example.com"
$sslCert = ""
$sslKey = ""
$listenSsl = "443"
$redirectHttps = $true
$serverName = $defaultDomain
$deployDir = $RootDir
$appPort = $defaultPort
$listenPort = if ($tier -eq "local") { "8080" } else { "80" }
$includeSystemd = $false

if ($genConfig) {
    $serverName = Prompt-Default $defaultDomain "Nginx server_name (public hostname)"
    $defaultDeploy = if ($tier -eq "local") { $RootDir } else { "C:\$serviceUnit-$tier" }
    $deployDir = Prompt-Default $defaultDeploy "Absolute app root (uploads path)"
    $deployDir = $deployDir -replace '\\', '/'
    $appPort = Prompt-Default $defaultPort "APP_PORT"
    $listenPort = Prompt-Default $(if ($tier -eq "local") { "8080" } else { "80" }) "Nginx HTTP listen port"

    Write-Host "TLS / SSL:" -ForegroundColor Cyan
    Write-Host "  1) none  2) Let's Encrypt (instructions + HTTP vhost)  3) HTTPS with PEM files"
    $sslPick = Prompt-Default "1" "Choose (1-3)"
    switch ($sslPick.Trim()) {
        "2" {
            $sslMode = "letsencrypt"
            $certbotEmail = Prompt-Default $certbotEmail "Email for Let's Encrypt"
        }
        "3" {
            $sslMode = "pem"
            $sslCert = Prompt-Default "$(Join-Path $sslShare 'cert.pem')" "Certificate PEM path"
            $sslKey = Prompt-Default "$(Join-Path $sslShare 'key.pem')" "Private key PEM path"
            $listenSsl = Prompt-Default "443" "HTTPS listen port"
            $redirectHttps = Prompt-Yn $true "Add HTTP -> HTTPS redirect on port $listenPort ?"
        }
        Default { $sslMode = "none" }
    }

    $includeSystemd = Prompt-Yn $false "Write systemd unit stub (for WSL/Linux reference)?"
}

$slug = Get-Slug "$tier-$serverName"
$genDir = Join-Path $RootDir "deploy\generated\$slug"
New-Item -ItemType Directory -Force -Path $genDir | Out-Null

$nginxPath = Join-Path $genDir "nginx-$serviceUnit-$tier.conf"
if ($genConfig) {
    $nginxLines = New-Object System.Collections.Generic.List[string]
    $nginxLines.Add("# Generated by scripts/deploy.ps1 — $genLabel ($tier)")
    if ($sslMode -eq "none" -or $sslMode -eq "letsencrypt") {
        $nginxLines.Add("server {")
        $nginxLines.Add("    listen $listenPort;")
        $nginxLines.Add("    server_name $serverName;")
        $nginxLines.Add("")
        $nginxLines.Add("    location /uploads/ {")
        $nginxLines.Add("        alias $deployDir/storage/uploads/;")
        $nginxLines.Add("        expires 30d;")
        $nginxLines.Add("        add_header Cache-Control `"public, max-age=2592000`";")
        $nginxLines.Add("    }")
        $nginxLines.Add("")
        $nginxLines.Add("    location / {")
        $nginxLines.Add("        proxy_pass         http://127.0.0.1:$appPort;")
        $nginxLines.Add("        proxy_set_header   Host `$host;")
        $nginxLines.Add("        proxy_set_header   X-Real-IP `$remote_addr;")
        $nginxLines.Add("        proxy_set_header   X-Forwarded-For `$proxy_add_x_forwarded_for;")
        $nginxLines.Add("        proxy_set_header   X-Forwarded-Proto `$scheme;")
        $nginxLines.Add("    }")
        $nginxLines.Add("}")
    }
    elseif ($sslMode -eq "pem") {
        if ($redirectHttps) {
            $nginxLines.Add("server {")
            $nginxLines.Add("    listen $listenPort;")
            $nginxLines.Add("    server_name $serverName;")
            $nginxLines.Add("    return 301 https://`$host`$request_uri;")
            $nginxLines.Add("}")
            $nginxLines.Add("")
        }
        $nginxLines.Add("server {")
        $nginxLines.Add("    listen $listenSsl ssl http2;")
        $nginxLines.Add("    server_name $serverName;")
        $nginxLines.Add("    ssl_certificate     $sslCert;")
        $nginxLines.Add("    ssl_certificate_key $sslKey;")
        $nginxLines.Add("")
        $nginxLines.Add("    location /uploads/ {")
        $nginxLines.Add("        alias $deployDir/storage/uploads/;")
        $nginxLines.Add("        expires 30d;")
        $nginxLines.Add("        add_header Cache-Control `"public, max-age=2592000`";")
        $nginxLines.Add("    }")
        $nginxLines.Add("")
        $nginxLines.Add("    location / {")
        $nginxLines.Add("        proxy_pass         http://127.0.0.1:$appPort;")
        $nginxLines.Add("        proxy_set_header   Host `$host;")
        $nginxLines.Add("        proxy_set_header   X-Real-IP `$remote_addr;")
        $nginxLines.Add("        proxy_set_header   X-Forwarded-For `$proxy_add_x_forwarded_for;")
        $nginxLines.Add("        proxy_set_header   X-Forwarded-Proto `$scheme;")
        $nginxLines.Add("    }")
        $nginxLines.Add("}")
    }
    Set-Content -Path $nginxPath -Value ($nginxLines -join "`n") -Encoding UTF8
}

if ($includeSystemd) {
    $runUser = Prompt-Default "www-data" "systemd User"
    $runGroup = Prompt-Default "www-data" "systemd Group"
    $unitPath = Join-Path $genDir "$serviceUnit@$tier.service"
    $unit = @"
# Generated by scripts/deploy.ps1 — use on Linux/WSL per docs/DEPLOYMENT.md

[Unit]
Description=$genLabel ($tier)
After=network.target

[Service]
Type=simple
WorkingDirectory=$deployDir
Environment=APP_ENV=$tier
EnvironmentFile=$deployDir/.env
ExecStart=$deployDir/bin/server
Restart=always
RestartSec=5
User=$runUser
Group=$runGroup

[Install]
WantedBy=multi-user.target
"@
    Set-Content -Path $unitPath -Value $unit -Encoding UTF8
}

$readmePath = Join-Path $genDir "README.md"
$tlsNote = switch ($sslMode) {
    "letsencrypt" { "Let's Encrypt: use Certbot on the Linux host after installing this HTTP vhost. See docs/DEPLOYMENT.md. Email for registration: $certbotEmail" }
    "pem" { "HTTPS PEM paths are embedded in the Nginx config." }
    Default { "HTTP only; add TLS per docs." }
}
$readme = @"
# Generated $genLabel deploy (Windows)

- **Tier:** $tier
- **TLS:** $sslMode — $tlsNote
- **App root:** $deployDir

## Nginx

Merge ``nginx-$serviceUnit-$tier.conf`` into your Nginx install or use WSL2 and run ``./scripts/deploy`` (bash) on Ubuntu.

## API on Windows

Ensure ``.env`` exists. From repo root: ``go run ./cmd/api`` or build with ``go build -o bin/server.exe ./cmd/api``.

"@
Set-Content -Path $readmePath -Value $readme -Encoding UTF8

Write-Host ""
Write-Host "Wrote: $genDir" -ForegroundColor Green

if (Prompt-Yn $true "Run build + migrate now (requires Go and .env)?") {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Host "Go not found; skipping." -ForegroundColor Red
    }
    elseif (-not (Test-Path (Join-Path $RootDir ".env"))) {
        Write-Host ".env missing; skipping." -ForegroundColor Red
    }
    else {
        $branch = Branch-ForTier $tier
        if ($branch -and (Test-Path (Join-Path $RootDir ".git"))) {
            if (Prompt-Yn $true "Git fetch / checkout / pull $branch ?") {
                git -C $RootDir fetch origin
                git -C $RootDir checkout $branch
                git -C $RootDir pull origin $branch
            }
        }
        Push-Location $RootDir
        try {
            go env -w CGO_ENABLED=0 2>$null
            go mod download
            New-Item -ItemType Directory -Force -Path "bin" | Out-Null
            go build -v -o bin/server ./cmd/api
            go build -v -o bin/migrate ./cmd/migrate
            go build -v -o bin/seed ./cmd/seed
            Write-Host "Running migrations..." -ForegroundColor Green
            $migrateBin = Join-Path $RootDir "bin\migrate.exe"
            if (-not (Test-Path $migrateBin)) { $migrateBin = Join-Path $RootDir "bin\migrate" }
            & $migrateBin -command=up
            $seedBin = Join-Path $RootDir "bin\seed.exe"
            if (-not (Test-Path $seedBin)) { $seedBin = Join-Path $RootDir "bin\seed" }
            if ($env:RUN_SEED_ON_DEPLOY -eq "true") {
                & $seedBin
            }
            elseif (Prompt-Yn $false "Run seed?") {
                & $seedBin
            }
        }
        finally {
            Pop-Location
        }
        Write-Host "Build/migrate finished." -ForegroundColor Green
    }
}

Write-Host "Done. See $readmePath" -ForegroundColor Cyan
