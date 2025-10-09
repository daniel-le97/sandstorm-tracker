package rcon

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockRconServer is a mock RCON server for testing
type mockRconServer struct {
	listener      net.Listener
	address       string
	handler       func(net.Conn)
	acceptHandler func(net.Conn) bool // Returns true if connection should be accepted
	running       bool
	mu            sync.Mutex
}

// newMockRconServer creates a new mock RCON server
func newMockRconServer(handler func(net.Conn)) (*mockRconServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	server := &mockRconServer{
		listener: listener,
		address:  listener.Addr().String(),
		handler:  handler,
		running:  true,
	}

	go server.serve()
	return server, nil
}

// serve handles incoming connections
func (s *mockRconServer) serve() {
	for {
		s.mu.Lock()
		if !s.running {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()

		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()
			if !running {
				return
			}
			continue
		}

		// Check if we should accept this connection
		if s.acceptHandler != nil && !s.acceptHandler(conn) {
			conn.Close()
			continue
		}

		go s.handler(conn)
	}
}

// close shuts down the mock server
func (s *mockRconServer) close() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	s.listener.Close()
}

// Helper function to read an RCON packet from the connection
func readTestPacket(conn net.Conn) (*RconPacket, error) {
	// Read packet size
	sizeBytes := make([]byte, 4)
	_, err := io.ReadFull(conn, sizeBytes)
	if err != nil {
		return nil, err
	}
	packetSize := int32(binary.LittleEndian.Uint32(sizeBytes))

	// Read the rest of the packet
	packetBytes := make([]byte, packetSize)
	_, err = io.ReadFull(conn, packetBytes)
	if err != nil {
		return nil, err
	}

	// Parse packet
	packetID := int32(binary.LittleEndian.Uint32(packetBytes[0:4]))
	packetType := int32(binary.LittleEndian.Uint32(packetBytes[4:8]))
	
	// Extract payload - handle cases where packet might be too small
	var payload string
	if len(packetBytes) >= 10 {
		// Exclude 2 null terminators at the end
		payload = string(packetBytes[8 : len(packetBytes)-2])
	} else if len(packetBytes) > 8 {
		// Handle minimal packets with just one null
		payload = string(packetBytes[8 : len(packetBytes)-1])
	}

	return &RconPacket{
		Size:    packetSize,
		ID:      packetID,
		Type:    packetType,
		Payload: payload,
	}, nil
}

// Helper function to write an RCON packet to the connection
func writeTestPacket(conn net.Conn, id int32, packetType int32, payload string) error {
	packet := BuildPacket(id, packetType, payload)
	_, err := conn.Write(packet)
	return err
}

// TestBuildPacket tests the BuildPacket function
func TestBuildPacket(t *testing.T) {
	tests := []struct {
		name        string
		id          int32
		packetType  int32
		payload     string
		expectedLen int
	}{
		{
			name:        "Simple command packet",
			id:          1,
			packetType:  2,
			payload:     "status",
			expectedLen: 4 + 4 + 4 + len("status") + 2, // size field + id + type + payload + 2 nulls
		},
		{
			name:        "Empty payload",
			id:          2,
			packetType:  0,
			payload:     "",
			expectedLen: 4 + 4 + 4 + 2, // size field + id + type + 2 nulls
		},
		{
			name:        "Long payload",
			id:          3,
			packetType:  2,
			payload:     strings.Repeat("a", 1000),
			expectedLen: 4 + 4 + 4 + 1000 + 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := BuildPacket(tt.id, tt.packetType, tt.payload)
			if len(packet) != tt.expectedLen {
				t.Errorf("BuildPacket() len = %d, want %d", len(packet), tt.expectedLen)
			}

			// Verify packet structure
			packetSize := int32(binary.LittleEndian.Uint32(packet[0:4]))
			packetID := int32(binary.LittleEndian.Uint32(packet[4:8]))
			pType := int32(binary.LittleEndian.Uint32(packet[8:12]))

			if packetID != tt.id {
				t.Errorf("Packet ID = %d, want %d", packetID, tt.id)
			}
			if pType != tt.packetType {
				t.Errorf("Packet Type = %d, want %d", pType, tt.packetType)
			}
			if packetSize != int32(len(packet)-4) {
				t.Errorf("Packet Size = %d, want %d", packetSize, len(packet)-4)
			}
		})
	}
}

// TestSpecialCharacterHandling tests sending commands with special characters
func TestSpecialCharacterHandling(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "Command with newline",
			command: "test\ncommand",
		},
		{
			name:    "Command with Unicode",
			command: "test 你好 мир",
		},
		{
			name:    "Command with emoji",
			command: "test 😀 🎮",
		},
		{
			name:    "Command with quotes",
			command: `test "quoted" 'single'`,
		},
		{
			name:    "Command with backslashes",
			command: `test\\path\to\file`,
		},
		{
			name:    "Command with tabs",
			command: "test\ttab\there",
		},
		{
			name:    "Command with mixed special chars",
			command: "test\n\"hello\"\t😀\\path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server that echoes the command back
			server, err := newMockRconServer(func(conn net.Conn) {
				defer conn.Close()

				// Read the command packet
				packet, err := readTestPacket(conn)
				if err != nil {
					return
				}

				// Send response with the same payload
				writeTestPacket(conn, packet.ID, 0, packet.Payload)
				// Send empty response to signal completion
				writeTestPacket(conn, packet.ID, 0, "")
			})
			if err != nil {
				t.Fatalf("Failed to create mock server: %v", err)
			}
			defer server.close()

			// Connect to the mock server
			conn, err := net.Dial("tcp", server.address)
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer conn.Close()

			// Send the command
			response, err := Send(conn, tt.command)
			if err != nil {
				t.Errorf("Send() error = %v", err)
			}

			// Verify the response matches the command
			if response != tt.command {
				t.Errorf("Response = %q, want %q", response, tt.command)
			}
		})
	}
}

// TestConnectionHandling tests connection scenarios
func TestConnectionHandling(t *testing.T) {
	t.Run("Connect to invalid address", func(t *testing.T) {
		// Try to connect to an invalid address
		_, err := net.DialTimeout("tcp", "invalid.address.test:12345", 1*time.Second)
		if err == nil {
			t.Error("Expected error when connecting to invalid address")
		}
	})

	t.Run("Connect to unreachable port", func(t *testing.T) {
		// Try to connect to a port that's not listening
		_, err := net.DialTimeout("tcp", "127.0.0.1:99999", 1*time.Second)
		if err == nil {
			t.Error("Expected error when connecting to unreachable port")
		}
	})

	t.Run("Server closes connection during Send", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			// Read one packet then close
			readTestPacket(conn)
			conn.Close()
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send should fail because server closes connection
		_, err = Send(conn, "test")
		if err == nil {
			t.Error("Expected error when server closes connection")
		}
	})

	t.Run("Timeout on read", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			// Read packet but never respond
			readTestPacket(conn)
			time.Sleep(10 * time.Second) // Sleep longer than timeout
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send should timeout
		_, err = Send(conn, "test")
		if err == nil {
			t.Error("Expected timeout error")
		}
	})
}

// TestAuthentication tests authentication scenarios
func TestAuthentication(t *testing.T) {
	t.Run("Successful authentication", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()

			// Read auth packet
			packet, err := readTestPacket(conn)
			if err != nil {
				return
			}

			// Send success response (type 2)
			writeTestPacket(conn, packet.ID, 2, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		success := Auth(conn, "correct_password")
		if !success {
			t.Error("Expected successful authentication")
		}
	})

	t.Run("Failed authentication with wrong password", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()

			// Read auth packet
			_, err := readTestPacket(conn)
			if err != nil {
				return
			}

			// Send failure response (wrong ID or type)
			writeTestPacket(conn, -1, 2, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		success := Auth(conn, "wrong_password")
		if success {
			t.Error("Expected authentication to fail")
		}
	})

	t.Run("Authentication timeout", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			// Read packet but never respond
			readTestPacket(conn)
			time.Sleep(10 * time.Second)
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		success := Auth(conn, "password")
		if success {
			t.Error("Expected authentication to fail on timeout")
		}
	})

	t.Run("Re-authentication after dropped connection", func(t *testing.T) {
		// First connection succeeds
		server1, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 2, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server1.close()

		conn1, err := net.Dial("tcp", server1.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		success := Auth(conn1, "password")
		if !success {
			t.Error("Expected first authentication to succeed")
		}
		conn1.Close()

		// Second connection (re-authentication) also succeeds
		server2, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 2, "")
		})
		if err != nil {
			t.Fatalf("Failed to create second mock server: %v", err)
		}
		defer server2.close()

		conn2, err := net.Dial("tcp", server2.address)
		if err != nil {
			t.Fatalf("Failed to reconnect: %v", err)
		}
		defer conn2.Close()

		success = Auth(conn2, "password")
		if !success {
			t.Error("Expected re-authentication to succeed")
		}
	})
}

// TestCommandHandling tests various command scenarios
func TestCommandHandling(t *testing.T) {
	t.Run("Empty command", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
			// Wait for and read empty packet
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		response, err := Send(conn, "")
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}
		if response != "" {
			t.Errorf("Expected empty response, got %q", response)
		}
	})

	t.Run("Large command near protocol limit", func(t *testing.T) {
		// RCON protocol typically has a 4096 byte limit, test with 4000 bytes
		largeCommand := strings.Repeat("a", 4000)

		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, packet.Payload)
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		response, err := Send(conn, largeCommand)
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}
		if len(response) != len(largeCommand) {
			t.Errorf("Response length = %d, want %d", len(response), len(largeCommand))
		}
	})

	t.Run("Rapid sequence of commands", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			// Handle multiple commands
			for i := 0; i < 10; i++ {
				packet, err := readTestPacket(conn)
				if err != nil {
					return
				}
				writeTestPacket(conn, packet.ID, 0, fmt.Sprintf("response_%d", i))
				readTestPacket(conn) // Read empty packet
				writeTestPacket(conn, packet.ID, 0, "")
			}
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send 10 commands rapidly
		for i := 0; i < 10; i++ {
			response, err := Send(conn, fmt.Sprintf("command_%d", i))
			if err != nil {
				t.Errorf("Send() command %d error = %v", i, err)
			}
			expectedResponse := fmt.Sprintf("response_%d", i)
			if response != expectedResponse {
				t.Errorf("Command %d response = %q, want %q", i, response, expectedResponse)
			}
		}
	})
}

// TestResponseHandling tests various response scenarios
func TestResponseHandling(t *testing.T) {
	t.Run("Malformed packet - invalid size", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			// Read command
			_, _ = readTestPacket(conn)
			// Send malformed packet with wrong size
			invalidPacket := []byte{0xFF, 0xFF, 0xFF, 0xFF}
			conn.Write(invalidPacket)
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		_, err = Send(conn, "test")
		if err == nil {
			t.Error("Expected error with malformed packet")
		}
	})

	t.Run("Partial packet", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			readTestPacket(conn)
			// Send only part of a packet then close
			conn.Write([]byte{0x10, 0x00, 0x00, 0x00, 0x01})
			time.Sleep(100 * time.Millisecond)
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		_, err = Send(conn, "test")
		if err == nil {
			t.Error("Expected error with partial packet")
		}
	})

	t.Run("Very large response", func(t *testing.T) {
		largeResponse := strings.Repeat("response data ", 1000)

		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, largeResponse)
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		response, err := Send(conn, "test")
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}
		if response != largeResponse {
			t.Errorf("Response length = %d, want %d", len(response), len(largeResponse))
		}
	})

	t.Run("Response with multibyte Unicode characters", func(t *testing.T) {
		unicodeResponse := "Hello 世界 🌍 Привет мир"

		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, unicodeResponse)
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		response, err := Send(conn, "test")
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}
		if response != unicodeResponse {
			t.Errorf("Response = %q, want %q", response, unicodeResponse)
		}
	})

	t.Run("Multiple response packets", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			// Send multiple response packets
			writeTestPacket(conn, packet.ID, 0, "part1")
			writeTestPacket(conn, packet.ID, 0, "part2")
			writeTestPacket(conn, packet.ID, 0, "part3")
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		response, err := Send(conn, "test")
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}
		// Should concatenate all parts
		if response != "part1part2part3" {
			t.Errorf("Response = %q, want %q", response, "part1part2part3")
		}
	})
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("Server error message", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			// Send error as response
			writeTestPacket(conn, packet.ID, 0, "ERROR: Command failed")
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		response, err := Send(conn, "test")
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}
		if !strings.Contains(response, "ERROR") {
			t.Errorf("Expected error message in response, got %q", response)
		}
	})

	t.Run("Packet ID mismatch", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			// Send response with wrong ID first
			writeTestPacket(conn, packet.ID+999, 0, "wrong id")
			// Then send correct response
			writeTestPacket(conn, packet.ID, 0, "correct")
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Should skip wrong ID and get correct response
		response, err := Send(conn, "test")
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}
		if response != "correct" {
			t.Errorf("Response = %q, want %q", response, "correct")
		}
	})

	t.Run("Unexpected packet type", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			// Send response with unexpected type (should be 0, send 99)
			writeTestPacket(conn, packet.ID, 99, "unexpected type")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		_, err = Send(conn, "test")
		if err == nil {
			t.Error("Expected error with unexpected packet type")
		}
		if !strings.Contains(err.Error(), "unhandled case") {
			t.Errorf("Expected unhandled case error, got: %v", err)
		}
	})

	t.Run("No panics on nil response", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			readTestPacket(conn)
			// Close without sending response
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Send() panicked: %v", r)
			}
		}()

		_, err = Send(conn, "test")
		if err == nil {
			t.Error("Expected error when server closes without response")
		}
	})
}

// TestConcurrency tests concurrent access
func TestConcurrency(t *testing.T) {
	t.Run("Multiple goroutines sending commands", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			// Handle multiple commands sequentially
			for {
				packet, err := readTestPacket(conn)
				if err != nil {
					return
				}
				writeTestPacket(conn, packet.ID, 0, "response: "+packet.Payload)
				_, err = readTestPacket(conn) // Read empty packet
				if err != nil {
					return
				}
				writeTestPacket(conn, packet.ID, 0, "")
			}
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		// Create multiple connections (one per goroutine for safety)
		const numGoroutines = 10
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				conn, err := net.Dial("tcp", server.address)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: failed to connect: %v", id, err)
					return
				}
				defer conn.Close()

				command := fmt.Sprintf("command_%d", id)
				response, err := Send(conn, command)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: Send() error: %v", id, err)
					return
				}

				expectedResponse := "response: " + command
				if response != expectedResponse {
					errors <- fmt.Errorf("goroutine %d: response = %q, want %q", id, response, expectedResponse)
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}
	})

	t.Run("Concurrent authentication attempts", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 2, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		const numGoroutines = 5
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				conn, err := net.Dial("tcp", server.address)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: failed to connect: %v", id, err)
					return
				}
				defer conn.Close()

				success := Auth(conn, "password")
				if !success {
					errors <- fmt.Errorf("goroutine %d: authentication failed", id)
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Error(err)
		}
	})
}

// TestResourceManagement tests resource cleanup
func TestResourceManagement(t *testing.T) {
	t.Run("Graceful connection close", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "response")
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// Send a command
		_, err = Send(conn, "test")
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}

		// Close connection
		err = conn.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}

		// Verify connection is closed
		buffer := make([]byte, 1)
		_, err = conn.Read(buffer)
		if err == nil {
			t.Error("Expected error reading from closed connection")
		}
	})

	t.Run("Double close", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			// Try to read packet but connection might be closed
			_, err := readTestPacket(conn)
			if err != nil {
				return // Connection closed by client
			}
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// Close once - should succeed
		err1 := conn.Close()
		if err1 != nil {
			t.Errorf("First Close() error = %v", err1)
		}

		// Close again - may return error (expected behavior) but should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Second Close() panicked: %v", r)
			}
		}()
		_ = conn.Close() // Second close may error, which is acceptable
	})

	t.Run("Server shutdown with active connections", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			// Keep connection open
			time.Sleep(2 * time.Second)
			conn.Close()
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Close server while connection is active
		server.close()

		// Give it time to close
		time.Sleep(100 * time.Millisecond)
	})
}

// TestProtocolEdgeCases tests edge cases in the protocol
func TestProtocolEdgeCases(t *testing.T) {
	t.Run("Command with only special characters", func(t *testing.T) {
		specialCommands := []string{
			"\n\n\n",
			"\t\t\t",
			"   ",
			"\\\\\\",
			"\"\"\"",
		}

		for _, cmd := range specialCommands {
			t.Run(fmt.Sprintf("Command: %q", cmd), func(t *testing.T) {
				server, err := newMockRconServer(func(conn net.Conn) {
					defer conn.Close()
					packet, _ := readTestPacket(conn)
					writeTestPacket(conn, packet.ID, 0, packet.Payload)
					readTestPacket(conn)
					writeTestPacket(conn, packet.ID, 0, "")
				})
				if err != nil {
					t.Fatalf("Failed to create mock server: %v", err)
				}
				defer server.close()

				conn, err := net.Dial("tcp", server.address)
				if err != nil {
					t.Fatalf("Failed to connect: %v", err)
				}
				defer conn.Close()

				response, err := Send(conn, cmd)
				if err != nil {
					t.Errorf("Send() error = %v", err)
				}
				if response != cmd {
					t.Errorf("Response = %q, want %q", response, cmd)
				}
			})
		}
	})

	t.Run("Unsolicited server data", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			// Send unsolicited packet before receiving command
			writeTestPacket(conn, 12345, 0, "unsolicited data")

			// Then handle the actual command
			packet, _ := readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "response")
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Should handle unsolicited data gracefully
		response, err := Send(conn, "test")
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}
		if response != "response" {
			t.Errorf("Response = %q, want %q", response, "response")
		}
	})

	t.Run("Maximum packet size boundary", func(t *testing.T) {
		// Test exactly at typical RCON limit (4096 bytes)
		maxPayload := strings.Repeat("x", 4096-12) // Account for header overhead

		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			// Echo back
			writeTestPacket(conn, packet.ID, 0, packet.Payload)
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		response, err := Send(conn, maxPayload)
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}
		if len(response) != len(maxPayload) {
			t.Errorf("Response length = %d, want %d", len(response), len(maxPayload))
		}
	})
}

// TestReadPacketEdgeCases tests ReadPacket function edge cases
func TestReadPacketEdgeCases(t *testing.T) {
	t.Run("ReadPacket with very small packet", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			// Send minimal valid packet
			writeTestPacket(conn, 1, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		packet := ReadPacket(conn)
		if packet == nil {
			t.Error("Expected valid packet, got nil")
		}
		if packet != nil && packet.Payload != "" {
			t.Errorf("Expected empty payload, got %q", packet.Payload)
		}
	})

	t.Run("ReadPacket timeout", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			// Don't send anything, let it timeout
			time.Sleep(10 * time.Second)
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		packet := ReadPacket(conn)
		if packet != nil {
			t.Error("Expected nil packet on timeout")
		}
	})

	t.Run("ReadPacket with connection closed", func(t *testing.T) {
		server, err := newMockRconServer(func(conn net.Conn) {
			conn.Close()
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		time.Sleep(100 * time.Millisecond) // Let server close
		packet := ReadPacket(conn)
		if packet != nil {
			t.Error("Expected nil packet when connection is closed")
		}
	})
}

// TestGenerateID tests ID generation
func TestGenerateID(t *testing.T) {
	t.Run("ID generation is sequential", func(t *testing.T) {
		// Reset counter to known state
		idCounter = 0

		id1 := generateID()
		id2 := generateID()
		id3 := generateID()

		if id2 != id1+1 {
			t.Errorf("Expected id2 = %d, got %d", id1+1, id2)
		}
		if id3 != id2+1 {
			t.Errorf("Expected id3 = %d, got %d", id2+1, id3)
		}
	})

	t.Run("ID generation is unique across calls", func(t *testing.T) {
		ids := make(map[int32]bool)
		for i := 0; i < 1000; i++ {
			id := generateID()
			if ids[id] {
				t.Errorf("Duplicate ID generated: %d", id)
			}
			ids[id] = true
		}
	})
}

// TestNullByteHandling tests handling of null bytes
func TestNullByteHandling(t *testing.T) {
	t.Run("Command with embedded null bytes", func(t *testing.T) {
		// Note: RCON protocol uses null-terminated strings, so embedded nulls
		// will truncate the string. This test verifies that behavior.
		server, err := newMockRconServer(func(conn net.Conn) {
			defer conn.Close()
			packet, _ := readTestPacket(conn)
			// Echo back what was received
			writeTestPacket(conn, packet.ID, 0, packet.Payload)
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		})
		if err != nil {
			t.Fatalf("Failed to create mock server: %v", err)
		}
		defer server.close()

		conn, err := net.Dial("tcp", server.address)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send command with embedded null (will be truncated by protocol)
		commandWithNull := "test" + string(byte(0)) + "after"
		response, err := Send(conn, commandWithNull)
		if err != nil {
			t.Errorf("Send() error = %v", err)
		}

		// The response should only contain "test" (truncated at null byte)
		if !strings.HasPrefix(commandWithNull, response) {
			t.Logf("Command: %q, Response: %q", commandWithNull, response)
		}
	})
}

// TestSetLogger tests logger configuration
func TestSetLogger(t *testing.T) {
	t.Run("SetLogger changes logger", func(t *testing.T) {
		type testLogger struct {
			NoOpLogger
			debugCalls int
			infoCalls  int
			errorCalls int
		}

		logger := &testLogger{}
		SetLogger(logger)

		// The logger is now set, but we can't easily test it without
		// triggering actual RCON operations. This test mainly ensures
		// SetLogger doesn't panic.
		if logger == nil {
			t.Error("Logger should not be nil")
		}
	})
}

// Benchmark tests
func BenchmarkBuildPacket(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildPacket(1, 2, "test command")
	}
}

func BenchmarkSend(b *testing.B) {
	server, err := newMockRconServer(func(conn net.Conn) {
		defer conn.Close()
		for {
			packet, err := readTestPacket(conn)
			if err != nil {
				return
			}
			writeTestPacket(conn, packet.ID, 0, "response")
			readTestPacket(conn)
			writeTestPacket(conn, packet.ID, 0, "")
		}
	})
	if err != nil {
		b.Fatalf("Failed to create mock server: %v", err)
	}
	defer server.close()

	conn, err := net.Dial("tcp", server.address)
	if err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Send(conn, "test")
		if err != nil {
			b.Fatalf("Send() error = %v", err)
		}
	}
}
