# Tifybe CLI installer — Windows
#   irm https://tifybe.com/install.ps1 | iex
# Downloads the latest release binary, verifies its SHA-256 checksum when
# available, installs to %LOCALAPPDATA%\Programs\tifybe and adds it to PATH.
$ErrorActionPreference = "Stop"

$repo = "emirhannsarial/tifybe-cli"
$base = "https://github.com/$repo/releases/latest/download"
$asset = "tifybe-windows-amd64.exe"

$dir = Join-Path $env:LOCALAPPDATA "Programs\tifybe"
New-Item -ItemType Directory -Force -Path $dir | Out-Null
$exe = Join-Path $dir "tifybe.exe"
$tmp = Join-Path $env:TEMP "tifybe-download.exe"

Write-Host "Downloading $asset (latest release)..."
Invoke-WebRequest -Uri "$base/$asset" -OutFile $tmp -UseBasicParsing

# Checksum verification — releases from v1.1.2 onward publish .sha256 files.
try {
  $shaFile = Join-Path $env:TEMP "tifybe-download.sha256"
  Invoke-WebRequest -Uri "$base/$asset.sha256" -OutFile $shaFile -UseBasicParsing
  $expected = (Get-Content $shaFile -Raw).Split(" ")[0].Trim().ToLower()
  $actual = (Get-FileHash -Algorithm SHA256 $tmp).Hash.ToLower()
  if ($expected -ne $actual) {
    Remove-Item $tmp -Force
    throw "Checksum mismatch - aborting install."
  }
  Write-Host "Checksum verified."
  Remove-Item $shaFile -Force
} catch [System.Net.WebException] {
  Write-Host "note: no checksum published for this release; skipping verification."
}

Move-Item -Force $tmp $exe

# Add to the user PATH if missing (takes effect in new terminals).
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$dir*") {
  [Environment]::SetEnvironmentVariable("Path", "$userPath;$dir", "User")
  Write-Host "Added $dir to your PATH (open a new terminal to use it)."
}

Write-Host ""
Write-Host "OK tifybe installed to $exe" -ForegroundColor Green
& $exe --version 2>$null
Write-Host ""
Write-Host "Get started:  tifybe listen 8080"
