# qurl installer script for Windows
# Usage:
#   iwr -useb https://raw.githubusercontent.com/bkeane/qurl/main/install.ps1 | iex
#   iwr -useb https://raw.githubusercontent.com/bkeane/qurl/main/install.ps1 -OutFile install.ps1; .\install.ps1 v0.1.0

param(
    [string]$Version = "latest"
)

$ErrorActionPreference = 'Stop'

$GITHUB_REPO = "bkeane/qurl"
$BINARY_NAME = "qurl.exe"
$INSTALL_DIR = "$env:LOCALAPPDATA\Programs\qurl"

# Detect architecture
function Get-Architecture {
    $arch = [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")
    switch ($arch) {
        "AMD64" { return "x86_64" }
        "ARM64" { return "arm64" }
        default {
            Write-Host "Unsupported architecture: $arch" -ForegroundColor Red
            exit 1
        }
    }
}

# Get latest release version from GitHub
function Get-LatestVersion {
    try {
        $releases = Invoke-RestMethod -Uri "https://api.github.com/repos/$GITHUB_REPO/releases/latest"
        return $releases.tag_name
    }
    catch {
        Write-Host "Failed to get latest version: $_" -ForegroundColor Red
        exit 1
    }
}

# Download and install qurl
function Install-Qurl {
    $arch = Get-Architecture

    # Use provided version or get latest
    if ($Version -eq "latest") {
        $versionToInstall = Get-LatestVersion
    } else {
        $versionToInstall = $Version
    }

    Write-Host "Installing qurl $versionToInstall for Windows $arch..." -ForegroundColor Cyan

    $downloadUrl = "https://github.com/$GITHUB_REPO/releases/download/$versionToInstall/qurl_Windows_$arch.zip"
    $tempZip = "$env:TEMP\qurl.zip"
    $tempDir = "$env:TEMP\qurl-extract"

    # Download
    Write-Host "Downloading from $downloadUrl..."
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempZip
    }
    catch {
        Write-Host "Failed to download: $_" -ForegroundColor Red
        exit 1
    }

    # Extract
    Write-Host "Extracting..."
    if (Test-Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force
    }
    Expand-Archive -Path $tempZip -DestinationPath $tempDir -Force

    # Create install directory
    if (!(Test-Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
    }

    # Install binary
    Move-Item -Path "$tempDir\qurl.exe" -Destination "$INSTALL_DIR\$BINARY_NAME" -Force

    # Clean up
    Remove-Item -Path $tempZip -Force
    Remove-Item -Path $tempDir -Recurse -Force

    Write-Host "✓ qurl installed successfully to $INSTALL_DIR\$BINARY_NAME" -ForegroundColor Green
}

# Add to PATH if needed
function Update-Path {
    $currentPath = [System.Environment]::GetEnvironmentVariable("Path", "User")

    if ($currentPath -notlike "*$INSTALL_DIR*") {
        Write-Host "`n⚠ Adding $INSTALL_DIR to PATH..." -ForegroundColor Yellow

        # Add to user PATH
        $newPath = "$currentPath;$INSTALL_DIR"
        [System.Environment]::SetEnvironmentVariable("Path", $newPath, "User")

        # Update current session
        $env:Path = "$env:Path;$INSTALL_DIR"

        Write-Host "✓ PATH updated. You may need to restart your terminal." -ForegroundColor Green
    } else {
        Write-Host "✓ $INSTALL_DIR is already in PATH" -ForegroundColor Green
    }
}

# Show completion instructions
function Show-CompletionInstructions {
    Write-Host "`nShell completions:" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "PowerShell completions:"
    Write-Host "  Add to your `$PROFILE:"
    Write-Host "  qurl completion powershell | Out-String | Invoke-Expression"
    Write-Host ""
    Write-Host "Or install completions permanently:"
    Write-Host "  qurl completion powershell > $env:USERPROFILE\Documents\WindowsPowerShell\Modules\qurl-completion.ps1"
    Write-Host "  Add to `$PROFILE: Import-Module qurl-completion"
    Write-Host ""
}

# Main installation
function Main {
    Write-Host "Installing qurl..." -ForegroundColor Cyan
    Write-Host ""

    # Install qurl
    Install-Qurl

    # Update PATH
    Update-Path

    # Show completion instructions
    Show-CompletionInstructions

    Write-Host "Get started with:" -ForegroundColor Green
    Write-Host '  $env:QURL_OPENAPI = "https://your-api.com/openapi.json"'
    Write-Host "  qurl /api/endpoint"
    Write-Host ""
    Write-Host "For more information, visit: https://github.com/$GITHUB_REPO" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Note: You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
}

# Run installation
Main