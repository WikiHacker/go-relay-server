# Windows Service Management Script for SMTP Relay

param (
    [string]$action = "help",
    [string]$installDir = "C:\Program Files\SMTP-Relay"
)

function Get-ServiceStatus {
    $service = Get-Service -Name "SMTP-Relay" -ErrorAction SilentlyContinue
    if ($service) {
        return $service.Status
    }
    return $null
}

function Install-Service {
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir | Out-Null
    }

    # Copy executable and config
    Copy-Item -Path "..\build\smtp-relay-windows.exe" -Destination "$installDir\smtp-relay.exe"
    Copy-Item -Path "..\config\config.json" -Destination "$installDir\config.json"

    # Create service
    New-Service -Name "SMTP-Relay" `
        -BinaryPathName "$installDir\smtp-relay.exe --config $installDir\config.json" `
        -DisplayName "SMTP Relay Service" `
        -Description "SMTP Relay Service for email delivery" `
        -StartupType Automatic | Out-Null

    Write-Host "Service installed successfully"
}

function Uninstall-Service {
    $service = Get-Service -Name "SMTP-Relay" -ErrorAction SilentlyContinue
    if ($service) {
        Stop-Service -Name "SMTP-Relay" -Force
        sc.exe delete "SMTP-Relay" | Out-Null
    }

    if (Test-Path $installDir) {
        Remove-Item -Recurse -Force $installDir
    }

    Write-Host "Service uninstalled successfully"
}

function Start-Service {
    Start-Service -Name "SMTP-Relay"
    Write-Host "Service started"
}

function Stop-Service {
    Stop-Service -Name "SMTP-Relay" -Force
    Write-Host "Service stopped"
}

function Restart-Service {
    Restart-Service -Name "SMTP-Relay" -Force
    Write-Host "Service restarted"
}

function Show-Logs {
    $logPath = "$installDir\logs"
    if (Test-Path $logPath) {
        Get-Content "$logPath\output.log" -Tail 100
    }
    else {
        Write-Host "No logs found"
    }
}

function Show-Status {
    $status = Get-ServiceStatus
    if ($status) {
        Write-Host "Service status: $status"
    }
    else {
        Write-Host "Service is not installed"
    }
}

function Show-Help {
    Write-Host @"
SMTP Relay Service Management

Usage: manage-service.ps1 <action> [install-dir]

Actions:
  install     Install the service
  uninstall   Remove the service
  start       Start the service
  stop        Stop the service
  restart     Restart the service
  status      Show service status
  logs        View service logs
  help        Show this help message

Options:
  install-dir  Installation directory (default: C:\Program Files\SMTP-Relay)
"@
}

# Main execution
switch ($action.ToLower()) {
    "install"    { Install-Service }
    "uninstall"  { Uninstall-Service }
    "start"      { Start-Service }
    "stop"       { Stop-Service }
    "restart"    { Restart-Service }
    "status"     { Show-Status }
    "logs"       { Show-Logs }
    default      { Show-Help }
}
