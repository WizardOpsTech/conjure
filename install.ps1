# Conjure installer — Windows (PowerShell 5.1+)
# Usage: irm https://raw.githubusercontent.com/WizardOpsTech/conjure/main/install.ps1 | iex
$ErrorActionPreference = 'Stop'

# Windows arm64 is not supported by the published binaries
if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') {
    Write-Error "Windows arm64 is not currently supported. Use WSL2 and the Linux installer instead:`n  curl -sfL https://raw.githubusercontent.com/WizardOpsTech/conjure/main/install.sh | sh"
    exit 1
}

# Fetch latest release info from the GitHub API
Write-Host "Fetching latest Conjure release..."
$release = Invoke-RestMethod "https://api.github.com/repos/WizardOpsTech/conjure/releases/latest"
$tag     = $release.tag_name                # e.g. "v1.0.0"

$archive = "conjure_${tag}_Windows_x86_64.zip"
$url     = "https://github.com/WizardOpsTech/conjure/releases/download/$tag/$archive"

Write-Host "Installing conjure $tag for Windows x86_64..."

# Download to a unique temp directory
$tmp     = Join-Path $env:TEMP "conjure-install-$([System.IO.Path]::GetRandomFileName())"
$zipPath = Join-Path $tmp "conjure.zip"
New-Item -ItemType Directory -Force -Path $tmp | Out-Null

try {
    Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing
    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force

    # Install to %USERPROFILE%\bin — no admin required
    $installDir = Join-Path $env:USERPROFILE "bin"
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
    Move-Item -Force (Join-Path $tmp "conjure.exe") (Join-Path $installDir "conjure.exe")

    # Add the install directory to the user PATH if it is not already there
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($null -eq $userPath) { $userPath = "" }
    if ($userPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable("PATH", "$userPath;$installDir", "User")
        Write-Host ""
        Write-Host "conjure $tag installed to $installDir"
        Write-Host "Added $installDir to your PATH."
        Write-Host "Restart your terminal, then run: conjure --version"
    } else {
        Write-Host ""
        Write-Host "conjure $tag installed to $installDir"
        Write-Host "Run: conjure --version"
    }
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
