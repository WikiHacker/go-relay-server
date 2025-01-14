#!/bin/bash

# Variables
OS=$(uname -s)
case "$OS" in
    Linux*) 
        BINARY_NAME="smtp-relay-linux"
        INSTALL_DIR="${1:-/opt/smtp-relay}"
        SERVICE_NAME="smtp-relay"
        ;;
    Darwin*)
        BINARY_NAME="smtp-relay-macos"
        INSTALL_DIR="${1:-/usr/local/smtp-relay}"
        SERVICE_NAME="com.local.smtprelay"
        ;;
    CYGWIN*|MINGW*|MSYS*)
        BINARY_NAME="smtp-relay-windows.exe"
        INSTALL_DIR="${1:-C:/Program Files/smtp-relay}"
        SERVICE_NAME="SMTPRelay"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Additional directories
CERT_DIR="${INSTALL_DIR}/certs"
LOG_DIR="${INSTALL_DIR}/logs"
CONFIG_NAME="config.json"
BINARY_SOURCE="./${BINARY_NAME}" # Path to the binary in the current directory
CONFIG_SOURCE="./${CONFIG_NAME}" # Path to the config file in the current directory

# Function to display usage
usage() {
    echo "Usage: $0 [install-dir]"
    echo "Example:"
    echo "  Default: $0"
    echo "  Custom: $0 /path/to/your/directory"
    exit 1
}

# Function to create the installation directory
create_install_dir() {
    echo "Creating installation directory structure..."
    case "$OS" in
        CYGWIN*|MINGW*|MSYS*)
            # Windows requires special handling for Program Files
            mkdir -p "${INSTALL_DIR}" || {
                echo "Error: Failed to create directory ${INSTALL_DIR}."
                echo "You may need to run this script as Administrator."
                exit 1
            }
            mkdir -p "${CERT_DIR}" || {
                echo "Error: Failed to create certificate directory."
                exit 1
            }
            mkdir -p "${LOG_DIR}" || {
                echo "Error: Failed to create log directory."
                exit 1
            }
            ;;
        *)
            mkdir -p "${INSTALL_DIR}" || {
                echo "Error: Failed to create directory ${INSTALL_DIR}."
                exit 1
            }
            mkdir -p "${CERT_DIR}" || {
                echo "Error: Failed to create certificate directory."
                exit 1
            }
            mkdir -p "${LOG_DIR}" || {
                echo "Error: Failed to create log directory."
                exit 1
            }
            ;;
    esac
    echo "Directory structure created:"
    echo "  - Installation: ${INSTALL_DIR}"
    echo "  - Certificates: ${CERT_DIR}"
    echo "  - Logs: ${LOG_DIR}"
}

# Function to copy the binary
copy_binary() {
    echo "Copying binary to ${INSTALL_DIR}..."
    if [[ ! -f "${BINARY_SOURCE}" ]]; then
        echo "Error: Binary not found at ${BINARY_SOURCE}. Please compile the Go program first."
        exit 1
    fi
    
    case "$OS" in
        CYGWIN*|MINGW*|MSYS*)
            # Windows requires special handling for Program Files
            cp "${BINARY_SOURCE}" "${INSTALL_DIR}/${BINARY_NAME}" || {
                echo "Error: Failed to copy binary."
                echo "You may need to run this script as Administrator."
                exit 1
            }
            ;;
        *)
            cp "${BINARY_SOURCE}" "${INSTALL_DIR}/${BINARY_NAME}" || {
                echo "Error: Failed to copy binary."
                exit 1
            }
            ;;
    esac
    echo "Binary copied."
}

# Function to copy the config file
copy_config() {
    echo "Copying config file to ${INSTALL_DIR}..."
    if [[ ! -f "${CONFIG_SOURCE}" ]]; then
        echo "Error: Config file not found at ${CONFIG_SOURCE}."
        exit 1
    fi
    
    case "$OS" in
        CYGWIN*|MINGW*|MSYS*)
            # Windows requires special handling for Program Files
            cp "${CONFIG_SOURCE}" "${INSTALL_DIR}/${CONFIG_NAME}" || {
                echo "Error: Failed to copy config file."
                echo "You may need to run this script as Administrator."
                exit 1
            }
            ;;
        *)
            cp "${CONFIG_SOURCE}" "${INSTALL_DIR}/${CONFIG_NAME}" || {
                echo "Error: Failed to copy config file."
                exit 1
            }
            ;;
    esac
    echo "Config file copied."
}

# Function to set permissions
set_permissions() {
    echo "Setting permissions..."
    case "$OS" in
        CYGWIN*|MINGW*|MSYS*)
            # Windows doesn't use chmod, but we can set ACLs
            icacls "${INSTALL_DIR}/${BINARY_NAME}" /grant:r "$(whoami):(RX)" || {
                echo "Error: Failed to set permissions on the binary."
                echo "You may need to run this script as Administrator."
                exit 1
            }
            icacls "${INSTALL_DIR}/${CONFIG_NAME}" /grant:r "$(whoami):(R)" || {
                echo "Error: Failed to set permissions on the config file."
                echo "You may need to run this script as Administrator."
                exit 1
            }
            ;;
        *)
            chmod +x "${INSTALL_DIR}/${BINARY_NAME}" || {
                echo "Error: Failed to set executable permissions on the binary."
                exit 1
            }
            chmod 644 "${INSTALL_DIR}/${CONFIG_NAME}" || {
                echo "Error: Failed to set permissions on the config file."
                exit 1
            }
            ;;
    esac
    echo "Permissions set."
}

# Function to install as service
install_service() {
    echo "Installing as system service..."
    case "$OS" in
        Linux*)
            # Create systemd service file
            SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
            cat <<EOF | sudo tee "${SERVICE_FILE}" > /dev/null
[Unit]
Description=SMTP Relay Service
After=network.target

[Service]
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${BINARY_NAME} --config ${INSTALL_DIR}/config.json
Restart=always
StandardOutput=append:${LOG_DIR}/output.log
StandardError=append:${LOG_DIR}/error.log

[Install]
WantedBy=multi-user.target
EOF

            sudo systemctl daemon-reload
            sudo systemctl enable ${SERVICE_NAME}
            sudo systemctl start ${SERVICE_NAME}
            ;;
            
        Darwin*)
            # Create launchd plist
            PLIST_FILE="/Library/LaunchDaemons/${SERVICE_NAME}.plist"
            cat <<EOF | sudo tee "${PLIST_FILE}" > /dev/null
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
        <string>${INSTALL_DIR}/config.json</string>
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
</dict>
</plist>
EOF

            sudo launchctl load -w "${PLIST_FILE}"
            ;;
            
        CYGWIN*|MINGW*|MSYS*)
            # Create Windows service
            sc create ${SERVICE_NAME} binPath= "\"${INSTALL_DIR}\\${BINARY_NAME}\" --config \"${INSTALL_DIR}\\config.json\"" start= auto || {
                echo "Error: Failed to create Windows service."
                echo "You may need to run this script as Administrator."
                exit 1
            }
            sc description ${SERVICE_NAME} "SMTP Relay Service" || {
                echo "Warning: Failed to set service description"
            }
            net start ${SERVICE_NAME} || {
                echo "Error: Failed to start Windows service."
                exit 1
            }
            ;;
    esac
    echo "Service installed and started successfully."
}

# Main script logic
if [[ $# -gt 1 ]]; then
    usage
fi

echo "Setting up installation..."
create_install_dir
copy_binary
copy_config
set_permissions
install_service
echo "Setup complete. Files are ready in ${INSTALL_DIR}."
echo "Service '${SERVICE_NAME}' has been installed and started."
