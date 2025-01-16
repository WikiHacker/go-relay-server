#!/bin/bash

# Variables
OS=$(uname -s)
case "$OS" in
    Linux*) 
        BINARY_NAME="smtp-relay-linux"
        INSTALL_DIR="${2:-/opt/smtp-relay}"
        SERVICE_NAME="smtp-relay"
        SERVICE_TYPE="systemd"
        ;;
    Darwin*)
        BINARY_NAME="smtp-relay-macos"
        INSTALL_DIR="${2:-/usr/local/smtp-relay}"
        SERVICE_NAME="com.local.smtprelay"
        SERVICE_TYPE="launchd"
        ;;
    CYGWIN*|MINGW*|MSYS*)
        BINARY_NAME="smtp-relay-windows.exe"
        INSTALL_DIR="${2:-C:/Program Files/smtp-relay}"
        SERVICE_NAME="SMTPRelay"
        SERVICE_TYPE="windows"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

CONFIG_NAME="config.json"
LOG_DIR="${INSTALL_DIR}/logs"
CURRENT_USER=$(whoami)

# Function to display usage
usage() {
    echo "Usage: $0 {install|uninstall|start|stop|restart|status|logs} [install-dir]"
    echo "Example:"
    echo "  Install: $0 install /opt/smtp-relay"
    echo "  Uninstall: $0 uninstall"
    echo "  View logs: $0 logs"
    exit 1
}

# Function to verify installation
verify_installation() {
    if [[ ! -f "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
        echo "Error: Binary not found at ${INSTALL_DIR}/${BINARY_NAME}"
        exit 1
    fi
    
    if [[ ! -f "${INSTALL_DIR}/${CONFIG_NAME}" ]]; then
        echo "Error: Config file not found at ${INSTALL_DIR}/${CONFIG_NAME}"
        exit 1
    fi
    
    if [[ ! -d "${LOG_DIR}" ]]; then
        echo "Error: Log directory not found at ${LOG_DIR}"
        exit 1
    fi
}

# Function to create service file
create_service_file() {
    echo "Creating service configuration..."
    
    case "$SERVICE_TYPE" in
        systemd)
            SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
            sudo bash -c "cat > ${SERVICE_FILE}" <<EOF
[Unit]
Description=SMTP Relay Service
After=network.target

[Service]
Type=simple
User=${CURRENT_USER}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${BINARY_NAME} --config ${INSTALL_DIR}/${CONFIG_NAME}
Restart=always
RestartSec=5
StandardOutput=append:${LOG_DIR}/output.log
StandardError=append:${LOG_DIR}/error.log
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF
            ;;
            
        launchd)
            SERVICE_FILE="/Library/LaunchDaemons/${SERVICE_NAME}.plist"
            sudo bash -c "cat > ${SERVICE_FILE}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>${SERVICE_NAME}</string>
    <key>ProgramArguments</key>
    <array>
        <string>${INSTALL_DIR}/${BINARY_NAME}</string>
        <string>--config</string>
        <string>${INSTALL_DIR}/${CONFIG_NAME}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>${LOG_DIR}/output.log</string>
    <key>StandardErrorPath</key>
    <string>${LOG_DIR}/error.log</string>
    <key>WorkingDirectory</key>
    <string>${INSTALL_DIR}</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>GIN_MODE</key>
        <string>release</string>
    </dict>
</dict>
</plist>
EOF
            ;;
            
        windows)
            # Windows service creation using sc.exe
            sc create ${SERVICE_NAME} binPath= "\"${INSTALL_DIR}\\${BINARY_NAME}\" --config \"${INSTALL_DIR}\\${CONFIG_NAME}\"" start= auto || {
                echo "Error: Failed to create Windows service"
                exit 1
            }
            sc description ${SERVICE_NAME} "SMTP Relay Service" || {
                echo "Warning: Failed to set service description"
            }
            ;;
    esac
    echo "Service configuration created"
}

# Function to install the service
install_service() {
    verify_installation
    create_service_file
    
    case "$SERVICE_TYPE" in
        systemd)
            sudo systemctl daemon-reload
            sudo systemctl enable ${SERVICE_NAME}
            sudo systemctl start ${SERVICE_NAME}
            ;;
            
        launchd)
            sudo launchctl load -w ${SERVICE_FILE}
            sudo launchctl start ${SERVICE_NAME}
            ;;
            
        windows)
            net start ${SERVICE_NAME} || {
                echo "Error: Failed to start Windows service"
                exit 1
            }
            ;;
    esac
    
    echo "Service installed and started successfully"
}

# Function to uninstall the service
uninstall_service() {
    case "$SERVICE_TYPE" in
        systemd)
            sudo systemctl stop ${SERVICE_NAME} || true
            sudo systemctl disable ${SERVICE_NAME} || true
            sudo rm -f /etc/systemd/system/${SERVICE_NAME}.service
            sudo systemctl daemon-reload
            ;;
            
        launchd)
            sudo launchctl stop ${SERVICE_NAME} || true
            sudo launchctl unload ${SERVICE_FILE} || true
            sudo rm -f ${SERVICE_FILE}
            ;;
            
        windows)
            net stop ${SERVICE_NAME} || true
            sc delete ${SERVICE_NAME} || {
                echo "Error: Failed to delete Windows service"
                exit 1
            }
            ;;
    esac
    
    echo "Service uninstalled successfully"
}

# Function to start the service
start_service() {
    case "$SERVICE_TYPE" in
        systemd)
            sudo systemctl start ${SERVICE_NAME}
            ;;
            
        launchd)
            sudo launchctl start ${SERVICE_NAME}
            ;;
            
        windows)
            net start ${SERVICE_NAME}
            ;;
    esac
    
    echo "Service started"
}

# Function to stop the service
stop_service() {
    case "$SERVICE_TYPE" in
        systemd)
            sudo systemctl stop ${SERVICE_NAME}
            ;;
            
        launchd)
            sudo launchctl stop ${SERVICE_NAME}
            ;;
            
        windows)
            net stop ${SERVICE_NAME}
            ;;
    esac
    
    echo "Service stopped"
}

# Function to restart the service
restart_service() {
    case "$SERVICE_TYPE" in
        systemd)
            sudo systemctl restart ${SERVICE_NAME}
            ;;
            
        launchd)
            sudo launchctl stop ${SERVICE_NAME}
            sudo launchctl start ${SERVICE_NAME}
            ;;
            
        windows)
            net stop ${SERVICE_NAME}
            net start ${SERVICE_NAME}
            ;;
    esac
    
    echo "Service restarted"
}

# Function to check service status
status_service() {
    case "$SERVICE_TYPE" in
        systemd)
            sudo systemctl status ${SERVICE_NAME}
            ;;
            
        launchd)
            sudo launchctl list | grep ${SERVICE_NAME}
            ;;
            
        windows)
            sc query ${SERVICE_NAME}
            ;;
    esac
}

# Function to view logs
view_logs() {
    case "$SERVICE_TYPE" in
        systemd)
            journalctl -u ${SERVICE_NAME} -n 100 --no-pager
            ;;
            
        *)
            echo "Output log:"
            tail -n 100 ${LOG_DIR}/output.log
            echo -e "\nError log:"
            tail -n 100 ${LOG_DIR}/error.log
            ;;
    esac
}

# Main script logic
if [[ $# -lt 1 ]]; then
    usage
fi

case "$1" in
    install)
        install_service
        ;;
    uninstall)
        uninstall_service
        ;;
    start)
        start_service
        ;;
    stop)
        stop_service
        ;;
    restart)
        restart_service
        ;;
    status)
        status_service
        ;;
    logs)
        view_logs
        ;;
    *)
        usage
        ;;
esac
