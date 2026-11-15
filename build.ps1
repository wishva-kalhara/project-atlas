<#
.SYNOPSIS
    Builds harness.exe, tools.exe, and workflow-exec.exe and places the
    final artifacts in .dist/
.DESCRIPTION
    Each module is built with `go build` from its own directory (so the
    working-directory-relative paths in each binary resolve correctly), then
    the resulting .exe is copied into the repo-root .dist/ folder.

    The script is idempotent: .dist is wiped and rebuilt on every run.
#>
[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

# Resolve repo root from this script's location (works regardless of the
# current working directory).
$RepoRoot = Split-Path -Parent $PSCommandPath
$DistDir  = Join-Path $RepoRoot ".dist"

$modules = @(
    @{ Name = "harness";             Path = "harness";             Out = "harness.exe" }
    @{ Name = "tools";               Path = "tools";               Out = "tools.exe" }
    @{ Name = "workflow-automation"; Path = "workflow-automation"; Out = "workflow-exec.exe" }
)

Write-Host "==> Building artifacts into $DistDir"

# Fresh output directory.
if (Test-Path $DistDir) {
    Remove-Item -Recurse -Force $DistDir
}
New-Item -ItemType Directory -Force -Path $DistDir | Out-Null

foreach ($m in $modules) {
    $modDir = Join-Path $RepoRoot $m.Path
    $outBin = Join-Path $modDir $m.Out
    $dest   = Join-Path $DistDir $m.Out

    Write-Host "==> Building $($m.Name) ($($m.Out))"

    if (-not (Test-Path (Join-Path $modDir "go.mod"))) {
        Write-Error "No go.mod found in $modDir - skipping."
        continue
    }

    # Build from inside the module directory so relative paths work.
    Push-Location $modDir
    try {
        go build -o $m.Out .
        if ($LASTEXITCODE -ne 0) {
            Write-Error "go build failed for $($m.Name)"
            exit 1
        }
    } finally {
        Pop-Location
    }

    if (-not (Test-Path $outBin)) {
        Write-Error "Build produced no binary at $outBin"
        exit 1
    }

    Copy-Item -Path $outBin -Destination $DistDir -Force
    Write-Host "    -> $dest"
}

Write-Host ""
Write-Host "==> Done. Artifacts in $DistDir"
Get-ChildItem -File -Path $DistDir | ForEach-Object {
    $size = [math]::Round($_.Length / 1KB, 1)
    Write-Host ("    {0,-22} {1,8} KB" -f $_.Name, $size)
}