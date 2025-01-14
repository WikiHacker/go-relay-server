# Windows System Setup Script for SMTP Relay

param (
    [string]$installDir = "C:\Program Files\SMTP-Relay"
)

function Test-Admin {
    $currentUser = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    return $currentUser.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Install-Prerequisites {
    Write-Host "Checking and installing prerequisites..."

    # Check if .NET is installed
    if (-not (Get-Command "dotnet" -ErrorAction SilentlyContinue)) {
        Write-Host "Installing .NET Runtime..."
        winget install Microsoft.DotNet.Runtime.6
    }

    # Check if Go is installed
    if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
        Write-Host "Installing Go..."
        winget install GoLang.Go
    }

    # Check if Git is installed
    if (-not (Get-Command "git" -ErrorAction SilentlyContinue)) {
        Write-Host "Installing Git..."
        winget install Git.Git
    }

    # Check if OpenSSL is installed
    if (-not (Get-Command "openssl" -ErrorAction SilentlyContinue)) {
        Write-Host "Installing OpenSSL..."
        winget install ShiningLight.OpenSSL
    }

    Write-Host "Prerequisites installed successfully"
}

function Configure-Firewall {
    Write-Host "Configuring Windows Firewall..."
    
    # Allow SMTP port
    New-NetFirewallRule -DisplayName "SMTP Relay" `
        -Direction Inbound `
        -LocalPort 2525 `
        -Protocol TCP `
        -Action Allow | Out-Null

    Write-Host "Firewall configured successfully"
}

function Create-InstallationDirectory {
    if (-not (Test-Path $installDir)) {
        Write-Host "Creating installation directory..."
        New-Item -ItemType Directory -Path $installDir | Out-Null
    }
}

function Set-EnvironmentVariables {
    Write-Host "Setting environment variables..."
    
    # Add Go to PATH if not already present
    $goPath = "$env:ProgramFiles\Go\bin"
    if ($env:Path -notmatch [regex]::Escape($goPath)) {
        [System.Environment]::SetEnvironmentVariable(
            "Path",
            "$env:Path;$goPath",
            [System.EnvironmentVariableTarget]::Machine
        )
    }

    # Set SMTP_RELAY_HOME
    [System.Environment]::SetEnvironmentVariable(
        "SMTP_RELAY_HOME",
        $installDir,
        [System.EnvironmentVariableTarget]::Machine
    )

    Write-Host "Environment variables configured"
}

function Show-Help {
    Write-Host @"
SMTP Relay Setup Script

Usage: setup-installation.ps1 [install-dir]

Options:
  install-dir  Installation directory (default: C:\Program Files\SMTP-Relay)

This script will:
1. Install required dependencies (.NET, Go, Git, OpenSSL)
2. Configure Windows Firewall
3. Create installation directory
4. Set environment variables
"@
}

# Main execution
if (-not (Test-Admin)) {
    Write-Host "This script must be run as administrator" -ForegroundColor Red
    exit 1
}

if ($args[0] -eq "help") {
    Show-Help
    exit 0
}

try {
    Install-Prerequisites
    Configure-Firewall
    Create-InstallationDirectory
    Set-EnvironmentVariables
    
    Write-Host @"
Setup completed successfully!

Next steps:
1. Build the application using compile.ps1
2. Install the service using manage-service.ps1 install
"@
}
catch {
    Write-Host "Setup failed: $_" -ForegroundColor Red
    exit 1
}
