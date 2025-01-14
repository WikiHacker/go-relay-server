package main

import (
	"flag"
	"fmt"
	"go-relay-server/config"
	"go-relay-server/server"
	"log"
	"os"
)

var (
	startCmd   = flag.NewFlagSet("start", flag.ExitOnError)
	stopCmd    = flag.NewFlagSet("stop", flag.ExitOnError)
	restartCmd = flag.NewFlagSet("restart", flag.ExitOnError)
	statusCmd  = flag.NewFlagSet("status", flag.ExitOnError)
	versionCmd = flag.NewFlagSet("version", flag.ExitOnError)
)

// Define the banner constant
const banner = `
██╗    ██╗██╗██╗  ██╗██╗██╗  ██╗ █████╗  ██████╗██╗  ██╗███████╗██████╗ 
██║    ██║██║██║ ██╔╝██║██║  ██║██╔══██╗██╔════╝██║ ██╔╝██╔════╝██╔══██╗
██║ █╗ ██║██║█████╔╝ ██║███████║███████║██║     █████╔╝ █████╗  ██████╔╝
██║███╗██║██║██╔═██╗ ██║██╔══██║██╔══██║██║     ██╔═██╗ ██╔══╝  ██╔══██╗
╚███╔███╔╝██║██║  ██╗██║██║  ██║██║  ██║╚██████╗██║  ██╗███████╗██║  ██║
 ╚══╝╚══╝ ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
Multi-Relay SMTP Server - simple SMTP Server Written in GO!
Version: 1.0.0
`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(banner)
		fmt.Println("Usage: smtp-relay <command>")
		fmt.Println("\nCommands:")
		fmt.Println("  start\t\tStart the SMTP relay server")
		fmt.Println("  stop\t\tStop the SMTP relay server")
		fmt.Println("  restart\tRestart the SMTP relay server")
		fmt.Println("  status\tCheck server status")
		fmt.Println("  version\tShow version information")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		startCmd.Parse(os.Args[2:])
		startServer()
	case "stop":
		stopCmd.Parse(os.Args[2:])
		stopServer()
	case "restart":
		restartCmd.Parse(os.Args[2:])
		restartServer()
	case "status":
		statusCmd.Parse(os.Args[2:])
		checkStatus()
	case "version":
		versionCmd.Parse(os.Args[2:])
		fmt.Println(banner)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func startServer() {
	fmt.Println(banner)

	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	server, err := server.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-make(chan struct{})
	server.Stop()
}

func stopServer() {
	server, err := getServerInstance()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	fmt.Println("Stopping server...")
	server.Stop()
	fmt.Println("Server stopped successfully")
}

func restartServer() {
	server, err := getServerInstance()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	fmt.Println("Restarting server...")
	if err := server.Restart(); err != nil {
		log.Fatalf("Failed to restart server: %v", err)
	}
	fmt.Println("Server restarted successfully")
}

func checkStatus() {
	server, err := getServerInstance()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	status := server.Status()
	fmt.Printf("Server status: %s\n", status)
}

func getServerInstance() (*server.Server, error) {
	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	return server.NewServer(config)
}
