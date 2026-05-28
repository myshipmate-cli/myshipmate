# Shipmate Installer for Windows (PowerShell)
# Usage: irm myshipmate.cc/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "shipmate/cli"
$Version = "v0.1.0"
$InstallDir = "$env:LOCALAPPDATA\Shipmate"

Write-Host ""
Write-Host "  ╔═══════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "  ║       🚀 SHIPMATE INSTALLER           ║" -ForegroundColor Cyan
Write-Host "  ║   The Smart Deployer for Developers   ║" -ForegroundColor Cyan
Write-Host "  ╚═══════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq "Arm64") {
        "arm64"
    } else {
        "amd64"
    }
} else {
    Write-Host "  ✗ 32-bit systems are not supported" -ForegroundColor Red
    exit 1
}

Write-Host "  ✓ Detected: windows_$Arch" -ForegroundColor Green

# Create install directory
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Write-Host "  ✓ Created $InstallDir" -ForegroundColor Green
}

# Download binary
$BinaryName = "shipmate_windows_${Arch}.exe"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$BinaryName"
$BinaryPath = "$InstallDir\shipmate.exe"

Write-Host "  ↓ Downloading Shipmate $Version..." -ForegroundColor Yellow

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $BinaryPath -UseBasicParsing
    Write-Host "  ✓ Downloaded to $BinaryPath" -ForegroundColor Green
} catch {
    Write-Host "  ⚠ Pre-built binary not available yet." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  To install from source, you need Go:" -ForegroundColor White
    Write-Host "    1. Install Go from https://go.dev/dl/" -ForegroundColor White
    Write-Host "    2. Run: go install github.com/shipmate/cli@latest" -ForegroundColor White
    Write-Host ""
    exit 1
}

# Add to PATH
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "User")
    Write-Host "  ✓ Added to PATH" -ForegroundColor Green
    Write-Host ""
    Write-Host "  ⚠ Restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "  ✓ Shipmate installed!" -ForegroundColor Green
Write-Host ""
Write-Host "  Get started:" -ForegroundColor White
Write-Host "    cd your-project" -ForegroundColor Gray
Write-Host "    shipmate" -ForegroundColor Gray
Write-Host ""
