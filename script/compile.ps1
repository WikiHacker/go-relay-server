# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Go is not installed. Please install Go first."
    exit 1
}

# Create output directory
New-Item -ItemType Directory -Path build -Force | Out-Null

# Compile for Windows
Write-Host "Compiling for Windows..."
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -o build/smtp-relay-windows.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "Windows compilation failed"
    exit 1
}

# Compile for Linux
Write-Host "Compiling for Linux..."
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o build/smtp-relay-linux.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "Linux compilation failed"
    exit 1
}

# Compile for MacOS
Write-Host "Compiling for MacOS..."
$env:GOOS="darwin"
$env:GOARCH="amd64"
go build -o build/smtp-relay-macos.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "MacOS compilation failed"
    exit 1
}

Write-Host "Compilation complete! Binaries are in the build/ directory"
