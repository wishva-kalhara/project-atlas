# install.ps1 - Atlas E2E Workflow Orchestrator Installer
$ErrorActionPreference = "Stop"

# --- CONFIGURATION ---
$RepoOwner = "YOUR_GITHUB_USERNAME"
$RepoName  = "YOUR_REPO_NAME"
# Uses 'latest' release automatically
$BaseUrl   = "https://github.com/$RepoOwner/$RepoName/releases/latest/download"
$Binaries  = @("atlas.exe", "tools.exe", "workflow-exec.exe", "harness.exe")
# ---------------------

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "   ATLAS E2E WORKFLOW ORCHESTRATOR SETUP  " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# 1. Ask the user for the installation path (with a sensible default)
$DefaultPath = Join-Path $env:LOCALAPPDATA "Atlas"
$InputPath = Read-Host "Enter installation path (Press Enter for default: $DefaultPath)"

if ([string]::IsNullOrWhiteSpace($InputPath)) {
    $InstallDir = $DefaultPath
} else {
    # Strip surrounding quotes if the user pasted a path like "C:\My Tools\Atlas"
    $InstallDir = $InputPath.Trim().Trim('"').Trim("'")
}

# 2. Create the directory if it doesn't exist
if (-not (Test-Path -Path $InstallDir)) {
    Write-Host "==> Creating directory: $InstallDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# Resolve full absolute path
$InstallDir = (Resolve-Path -Path $InstallDir).Path

# 3. Download all 4 binaries from GitHub Releases
Write-Host "==> Downloading Atlas binaries..." -ForegroundColor Cyan
foreach ($Bin in $Binaries) {
    $DownloadUrl = "$BaseUrl/$Bin"
    $OutFile     = Join-Path $InstallDir $Bin

    Write-Host "    Downloading $Bin..." -NoNewline
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $OutFile -UseBasicParsing
        Write-Host " Done!" -ForegroundColor Green
    } catch {
        Write-Host " Failed!" -ForegroundColor Red
        Write-Error "Could not download $Bin from $DownloadUrl. Verify the release exists and is public."
        exit 1
    }
}

# 4. Add the installation directory to the User PATH environment variable
Write-Host "==> Configuring Environment Variables (User PATH)..." -ForegroundColor Cyan
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")

# Check if the path is already in User PATH (case-insensitive)
$PathEntries = $UserPath -split ";" | Where-Object { $_ -ne "" }
if ($PathEntries -notcontains $InstallDir) {
    $NewPath = ($PathEntries + $InstallDir) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    
    # Also update the current PowerShell session's PATH immediately
    $env:Path = "$env:Path;$InstallDir"
    Write-Host "    Added '$InstallDir' to User PATH." -ForegroundColor Green
} else {
    Write-Host "    '$InstallDir' is already in User PATH." -ForegroundColor DarkGray
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host "You can now run 'atlas' from any terminal." -ForegroundColor White
Write-Host ""