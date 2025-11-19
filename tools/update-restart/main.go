package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	var (
		pid             = flag.Int("pid", 0, "PID of the app process to shutdown (optional if pidfile exists)")
		pidFile         = flag.String("pidfile", "sandstorm-tracker.pid", "Path to PID file (used if -pid not provided)")
		shutdownTimeout = flag.Duration("timeout", 30*time.Second, "Grace period for shutdown before force kill")
		outputDir       = flag.String("logs", "logs", "Directory for output logs")
		httpAddr        = flag.String("http", "0.0.0.0:8090", "HTTP address for app to listen on")
		appBinary       = flag.String("app", "sandstorm-tracker", "Name of the app binary to restart")
	)
	flag.Parse()

	logger := log.New(os.Stdout, "[UPDATE] ", log.LstdFlags)

	// Determine PID: use -pid flag if provided, otherwise read from file
	if *pid == 0 {
		// Try to read from PID file
		pidData, err := os.ReadFile(*pidFile)
		if err != nil {
			logger.Fatalf("PID not provided and unable to read PID file '%s': %v. Usage: update-restart [-pid <process-id>] [-pidfile <path>]", *pidFile, err)
		}

		pidStr := strings.TrimSpace(string(pidData))
		parsedPID, err := strconv.Atoi(pidStr)
		if err != nil {
			logger.Fatalf("Invalid PID in file '%s': '%s' is not a valid integer: %v", *pidFile, pidStr, err)
		}
		*pid = parsedPID
		logger.Printf("Read PID %d from file: %s", *pid, *pidFile)
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logger.Fatalf("Failed to create logs directory: %v", err)
	}

	logger.Println("Starting update and restart procedure")

	// Step 1: Stop the running app process by PID
	logger.Printf("Looking for process with PID %d...", *pid)
	proc, err := os.FindProcess(*pid)
	if err != nil {
		logger.Fatalf("Failed to find process with PID %d: %v", *pid, err)
	}

	logger.Printf("Found process PID %d. Sending graceful shutdown signal...", *pid)

	// Send SIGTERM for graceful shutdown
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		logger.Printf("Error sending SIGTERM: %v. Forcing kill.", err)
		proc.Kill()
	} else {
		// Wait for graceful shutdown
		done := make(chan error, 1)
		go func() {
			_, err := proc.Wait()
			done <- err
		}()

		select {
		case <-time.After(*shutdownTimeout):
			logger.Printf("Process did not exit within %v. Force killing.", *shutdownTimeout)
			proc.Kill()
			<-done
		case err := <-done:
			logger.Printf("Process exited gracefully: %v", err)
		}
	}

	// Small delay to ensure file system is ready
	time.Sleep(1 * time.Second)

	// Step 2: Run the update command
	logger.Println("Running update command...")
	updateCmd := exec.Command(*appBinary, "update")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	if err := updateCmd.Run(); err != nil {
		logger.Fatalf("Update command failed: %v", err)
	}
	logger.Println("Update completed successfully")

	// Small delay after update
	time.Sleep(1 * time.Second)

	// Step 3: Start the app
	logger.Println("Starting application...")

	outFile, err := os.Create(filepath.Join(*outputDir, "app_output.log"))
	if err != nil {
		logger.Fatalf("Failed to create output log: %v", err)
	}
	defer outFile.Close()

	errFile, err := os.Create(filepath.Join(*outputDir, "app_error.log"))
	if err != nil {
		logger.Fatalf("Failed to create error log: %v", err)
	}
	defer errFile.Close()

	cmd := exec.Command(*appBinary, "serve", fmt.Sprintf("--http=%s", *httpAddr))
	cmd.Stdout = outFile
	cmd.Stderr = errFile

	if err := cmd.Start(); err != nil {
		logger.Fatalf("Failed to start application: %v", err)
	}

	logger.Printf("Application started with PID %d", cmd.Process.Pid)
	logger.Println("Update and restart procedure completed successfully")
}
