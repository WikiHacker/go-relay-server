@echo off
echo Compiling for Windows...
go build -o smtp-relay-windows.exe main.go

echo Compiling for Linux...
set GOOS=linux
set GOARCH=amd64
go build -o smtp-relay-linux main.go

echo Compiling for MacOS...
set GOOS=darwin
set GOARCH=amd64
go build -o smtp-relay-macos main.go

echo Compilation complete!
pause
