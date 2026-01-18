# ostop installation script for Windows
# Usage: irm https://ostop.mkla.dev/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$REPO = "Vegasq/ostop"
$BINARY_NAME = "ostop.exe"
$INSTALL_DIR = if ($env:OSTOP_INSTALL_DIR) { $env:OSTOP_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\ostop" }

function Write-Info {
    param([string]$Message)
    Write-Host "==> " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "Warning: " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "Error: " -ForegroundColor Red -NoNewline
    Write-Host $Message
    exit 1
}

function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Write-Error-Custom "Unsupported architecture: $arch" }
    }
}

function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest"
        return $release.tag_name
    }
    catch {
        Write-Error-Custom "Failed to fetch latest version: $_"
    }
}

function Install-Ostop {
    Write-Info "Installing ostop - OpenSearch Terminal UI"
    Write-Host ""

    $arch = Get-Architecture
    $version = Get-LatestVersion

    Write-Info "Detected Architecture: $arch"
    Write-Info "Latest version: $version"

    # Construct download URL
    $filename = "ostop-${version}-windows-${arch}.zip"
    $downloadUrl = "https://github.com/$REPO/releases/download/${version}/${filename}"

    Write-Info "Downloading from: $downloadUrl"

    # Create temporary directory
    $tmpDir = Join-Path $env:TEMP "ostop-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        # Download
        $zipPath = Join-Path $tmpDir $filename
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing

        # Extract
        Write-Info "Extracting..."
        Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force

        # Find the binary
        $binaryPath = Get-ChildItem -Path $tmpDir -Filter $BINARY_NAME -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1

        if (-not $binaryPath) {
            Write-Error-Custom "Binary not found in archive"
        }

        # Create install directory if it doesn't exist
        if (-not (Test-Path $INSTALL_DIR)) {
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        }

        # Install
        Write-Info "Installing to $INSTALL_DIR..."
        $destPath = Join-Path $INSTALL_DIR $BINARY_NAME
        Copy-Item -Path $binaryPath.FullName -Destination $destPath -Force

        Write-Info "Installation complete! 🎉"
        Write-Host ""

        # Check if install directory is in PATH
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$INSTALL_DIR*") {
            Write-Warn "$INSTALL_DIR is not in your PATH"
            Write-Host ""
            Write-Host "To add it to your PATH, run:" -ForegroundColor Cyan
            Write-Host "`$env:Path += `";$INSTALL_DIR`"" -ForegroundColor White
            Write-Host ""
            Write-Host "To add it permanently, run PowerShell as Administrator and execute:" -ForegroundColor Cyan
            Write-Host "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';$INSTALL_DIR', 'User')" -ForegroundColor White
            Write-Host ""
            Write-Host "Or add it manually through System Properties > Environment Variables" -ForegroundColor Cyan
        }
        else {
            Write-Info "Run 'ostop --version' to verify installation"
            Write-Info "Run 'ostop --help' to see usage instructions"
        }
    }
    finally {
        # Cleanup
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Main execution
try {
    Install-Ostop
}
catch {
    Write-Error-Custom "Installation failed: $_"
}
