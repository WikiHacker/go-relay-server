#!/bin/bash

# Check if Go is installed
if ! command -v go &> /dev/null
then
    echo "Go is not installed. Please install Go first."
    exit 1
fi

# Create output directory
mkdir -p build

# Compile for Windows
echo "Compiling for Windows..."
GOOS=windows GOARCH=amd64 go build -o build/smtp-relay-windows.exe ../server/server.go
if [ $? -ne 0 ]; then
    echo "Windows compilation failed"
    exit 1
fi

# Compile for Linux  
echo "Compiling for Linux..."
GOOS=linux GOARCH=amd64 go build -o build/smtp-relay-linux.exe ../server/server.go
if [ $? -ne 0 ]; then
    echo "Linux compilation failed" 
    exit 1
fi

# Compile for MacOS
echo "Compiling for MacOS..."
GOOS=darwin GOARCH=amd64 go build -o build/smtp-relay-macos.exe ../server/server.go
if [ $? -ne 0 ]; then
    echo "MacOS compilation failed"
    exit 1
fi

echo "Compilation complete! Binaries are in the build/ directory"
