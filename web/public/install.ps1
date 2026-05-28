# Shipmate Installer for Windows
# Usage: iwr myshipmate.fly.dev/install.ps1 | iex
# Or: curl -L myshipmate.fly.dev/install.ps1 -o install.ps1; .\install.ps1

$ErrorActionPreference = "Stop"

$Repo = "myshipmate-cli/myshipmate"
$GitHubAPI = "https://api.github.com/repos/$Repo/releases/latest"
$InstallDir = "$env:LOCALAPPDATA\Shipmate"
$BinaryName = "shipmate.exe"

function Write-Info($msg) { Write-Host "  ✓ $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "  ⚠ $msg" -ForegroundColor Yellow }
function Write-Err($msg) { Write-Host "  ✗ $msg" -ForegroundColor Red; exit 1 }

Write-Host ""
Write-Host "  ╔═══════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "  ║       🚀 SHIPMATE INSTALLER           ║" -ForegroundColor Cyan
Write-Host "  ║   The Smart Deployer for Developers   ║" -ForegroundColor Cyan
Write-Host "  ╚═══════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Auto-detect latest version
Write-Host "  Checking latest version..." -NoNewline
try {
    $release = Invoke-RestMethod -Uri $GitHubAPI -UseBasicParsing
    $version = $release.tag_name
    Write-Host " $version" -ForegroundColor Green
} catch {
    $version = "v0.1.0"
    Write-Warn "Could not detect latest version, defaulting to $version"
}

# Determine architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$platform = "windows_$arch"
$downloadURL = "https://github.com/$Repo/releases/download/$version/shipmate_$platform.exe"

Write-Info "Platform: $platform"
Write-Info "Downloading from: $downloadURL"

# Create install directory
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Write-Info "Created install directory: $InstallDir"
}

# Download binary
$targetPath = Join-Path $InstallDir $BinaryName
try {
    Invoke-WebRequest -Uri $downloadURL -OutFile $targetPath -UseBasicParsing
    Write-Info "Downloaded to: $targetPath"
} catch {
    Write-Err "Download failed: $_"
}

# Add to PATH if not already there
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Info "Added $InstallDir to your PATH"
    Write-Warn "Restart PowerShell for PATH changes to take effect"
} else {
    Write-Info "Already in PATH"
}

Write-Host ""
Write-Host "  Shipmate installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "  Get started:" -ForegroundColor Cyan
Write-Host "    cd your-project"
Write-Host "    shipmate review"
Write-Host ""
Write-Host "  Run 'shipmate --help' for all commands."
Write-Host ""
