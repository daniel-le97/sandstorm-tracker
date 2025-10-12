// +build integration

package rcon

import (
    "os/exec"
    "testing"
    "time"
    "net"
)

func waitForPort(address string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        conn, err := net.Dial("tcp", address)
        if err == nil {
            conn.Close()
            return nil
        }
        time.Sleep(500 * time.Millisecond)
    }
    return &net.OpError{Op: "dial", Net: "tcp", Addr: nil, Err: err}
}

func TestRconClientWithSpawnedServer(t *testing.T) {
    // Adjust these paths and credentials for your setup
    serverPath := "/path/to/InsurgencyServer-Linux-Shipping"
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
    client, err := DialRcon(rconAddr, rconPassword, DefaultConfig())
    if err != nil {
        t.Fatalf("Failed to connect to RCON: %v", err)
    }
    defer client.Close()

    // Send a command (e.g., "ServerInfo")
    resp, err := client.Send("ServerInfo")
    if err != nil {
        t.Fatalf("Failed to send command: %v", err)
    }
    if resp == "" {
        t.Error("Expected non-empty response from server")
    }
}