# install.ps1 — Install hush binary on Windows from GitHub Releases.
#
# Usage:
#   irm https://raw.githubusercontent.com/biao29/hush/main/scripts/install.ps1 | iex
#   $env:HUSH_VERSION = "v0.1.0"; irm ... | iex
#   $env:HUSH_BIN_DIR = "$HOME\bin"; irm ... | iex

$ErrorActionPreference = "Stop"

$Repo = "biao29/hush"
$BinDir = if ($env:HUSH_BIN_DIR) { $env:HUSH_BIN_DIR } else { "$env:LOCALAPPDATA\hush\bin" }
$Version = if ($env:HUSH_VERSION) { $env:HUSH_VERSION } else { "latest" }

function Info($msg)  { Write-Host "  ✓ $msg" -ForegroundColor Green }
function Warn($msg)  { Write-Host "  ! $msg" -ForegroundColor Yellow }
function Fail($msg)  { Write-Host "  ✗ $msg" -ForegroundColor Red; exit 1 }

# Detect architecture
$Arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    "X64"   { "amd64" }
    "Arm64" { "arm64" }
    default { Fail "Unsupported architecture: $_" }
}

# Resolve latest version
if ($Version -eq "latest") {
    $Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $Release.tag_name
    if (-not $Version) { Fail "Could not determine latest version" }
}

$Archive = "hush-windows-$Arch.zip"
$Url = "https://github.com/$Repo/releases/download/$Version/$Archive"
$ChecksumsUrl = "https://github.com/$Repo/releases/download/$Version/checksums.txt"

$TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) "hush-install-$([System.Guid]::NewGuid().ToString('N').Substring(0,8))"
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

try {
    Info "Downloading hush $Version (windows/$Arch)..."
    Invoke-WebRequest -Uri $Url -OutFile "$TmpDir\$Archive" -UseBasicParsing
    Invoke-WebRequest -Uri $ChecksumsUrl -OutFile "$TmpDir\checksums.txt" -UseBasicParsing

    # Verify checksum
    Info "Verifying checksum..."
    $Expected = (Get-Content "$TmpDir\checksums.txt" | Where-Object { $_ -match $Archive }) -replace "\s+.*", ""
    $Actual = (Get-FileHash "$TmpDir\$Archive" -Algorithm SHA256).Hash.ToLower()
    if ($Expected -ne $Actual) {
        Fail "Checksum mismatch: expected $Expected, got $Actual"
    }
    Info "Checksum verified"

    # Extract and install
    Info "Installing to $BinDir..."
    Expand-Archive -Path "$TmpDir\$Archive" -DestinationPath $TmpDir -Force
    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
    Copy-Item "$TmpDir\hush.exe" "$BinDir\hush.exe" -Force

    Info "hush $Version installed to $BinDir\hush.exe"

    # Check PATH
    $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($UserPath -notlike "*$BinDir*") {
        Write-Host ""
        Warn "$BinDir is not in your PATH. Add it:"
        Write-Host ""
        Write-Host "  `$env:PATH = `"$BinDir;`$env:PATH`""
        Write-Host ""
        Write-Host "  # Or permanently:"
        Write-Host "  [Environment]::SetEnvironmentVariable('PATH', `"$BinDir;`$([Environment]::GetEnvironmentVariable('PATH', 'User'))`", 'User')"
    }
}
finally {
    Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
