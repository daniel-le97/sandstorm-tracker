package rcon

import (
	// "log"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func waitForPort(address string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := net.Dial("tcp", address)
		if err == nil {
			conn.Close()
			return nil
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	return lastErr
}

func TestRconClientWithSpawnedServer(t *testing.T) {
	err := godotenv.Load("..\\..\\.env")
	if err != nil {
		// panic(fmt.Sprintf("unable to load .env file: %v", err))
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("unable to get working directory: %v", err))
		}
		fmt.Printf("Working directory: %s\n", wd)
	}

	// Adjust these paths and credentials for your setup
	serverPath := os.Getenv("INSURGENCY_SERVER_PATH")
	if serverPath == "" {
		t.Skip("Skipping test: INSURGENCY_SERVER_PATH environment variable not set")
	}

	serverArgs := []string{
		"Oilfield?Scenario=Scenario_Oilfield_Checkpoint_Security",
		"Port=27102",
		"QueryPort=27131",
		"RCONEnabled=True",
		"RCONPort=27015",
		"RCONPassword=MyRconPassword",
	}
	rconAddr := "127.0.0.1:27015"
	rconPassword := "MyRconPassword"

	// Start the server
	cmd := exec.Command(serverPath, serverArgs...)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Wait for RCON port to be open
	if err := waitForPort(rconAddr, 60*time.Second); err != nil {
		t.Fatalf("RCON port did not open: %v", err)
	}

	// Connect and authenticate
	conn, err := net.Dial("tcp", rconAddr)
	if err != nil {
		t.Fatalf("Failed to connect to RCON: %v", err)
	}
	defer conn.Close()

	client := NewRconClient(conn, DefaultConfig())
	if !client.Auth(rconPassword) {
		t.Fatalf("Failed to authenticate to RCON")
	}

	// Send a command (e.g., "ServerInfo")
	resp, err := client.Send("listplayers")
	if err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}
	if resp == "" {
		t.Error("Expected non-empty response from server")
	}
}
