package rcon

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

// Logger interface for pluggable logging
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

func (l NoOpLogger) Debug(msg string, args ...any) {}
func (l NoOpLogger) Info(msg string, args ...any)  {}
func (l NoOpLogger) Warn(msg string, args ...any)  {}
func (l NoOpLogger) Error(msg string, args ...any) {}

// Config struct for RCON client
type ClientConfig struct {
	Timeout time.Duration
	Logger  Logger
}

func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		Timeout: 5 * time.Second,
		Logger:  NoOpLogger{},
	}
}

type RconPacket struct {
	Size    int32
	ID      int32
	Type    int32
	Payload string
}

var idCounter int32 = 0

func generateID() int32 {
	idCounter++
	return idCounter
}

// RconClient wraps a connection and config for testability
type RconClient struct {
	Conn   net.Conn
	Config *ClientConfig
}

func NewRconClient(conn net.Conn, config *ClientConfig) *RconClient {
	if config == nil {
		config = DefaultConfig()
	}
	return &RconClient{Conn: conn, Config: config}
}

// SetLogger allows changing the logger at runtime
func (c *RconClient) SetLogger(logger Logger) {
	c.Config.Logger = logger
}

// Auth authenticates with the RCON server
func (c *RconClient) Auth(password string) bool {
	authID := generateID()
	authPacket := BuildPacket(authID, 3, password)
	_, err := c.Conn.Write(authPacket)
	if err != nil {
		c.Config.Logger.Error("Error sending authentication packet", "error", err)
		return false
	}
	responsePacket, err := c.ReadPacket()
	if err != nil {
		c.Config.Logger.Error("Error reading authentication response", "error", err)
		return false
	}
	if responsePacket.ID == authID && responsePacket.Type == 2 {
		c.Config.Logger.Info("Authentication successful")
		return true
	} else {
		c.Config.Logger.Error("Authentication failed")
		return false
	}
}

// Repl provides a simple interactive shell for RCON
func (c *RconClient) Repl() {
	reader := bufio.NewReader(os.Stdin)
	prompt := fmt.Sprintf("RCON %s] ", c.Conn.RemoteAddr().String())

	fmt.Print(prompt)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			c.Config.Logger.Error("Error reading input", "error", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Print(prompt)
			continue
		}

		if input == "exit" || input == "quit" {
			break
		}

		err = c.SendAndPrint(input)
		if err != nil {
			c.Config.Logger.Error("Error executing command", "error", err)
		}

		fmt.Print(prompt)
	}
}

// SendAndPrint sends a command and prints the response
func (c *RconClient) SendAndPrint(command string) error {
	output, err := c.Send(command)
	if err != nil {
		return fmt.Errorf("error sending command: %s", err.Error())
	}
	if strings.TrimSpace(output) == "" {
		c.Config.Logger.Info("Server response empty")
		return nil
	}
	c.Config.Logger.Info("Server response:")
	fmt.Printf("%s\n", output)
	return nil
}

// Send sends a command and returns the response
func (c *RconClient) Send(command string) (string, error) {
	commandID := generateID()
	commandPacket := BuildPacket(commandID, 2, command)
	c.Config.Logger.Debug("Sending command", "command", command)
	_, err := c.Conn.Write(commandPacket)
	if err != nil {
		return "", fmt.Errorf("error sending command packet: %s", err.Error())
	}

	var (
		fullPayload     strings.Builder
		sentEmptyPacket bool
	)
	for {
		responsePacket, err := c.ReadPacket()
		if err != nil {
			return "", fmt.Errorf("error reading command response: %w", err)
		}

		if commandID != responsePacket.ID {
			c.Config.Logger.Error("Received packet with unexpected ID", "id", responsePacket.ID)
			continue
		}

		if responsePacket.ID == commandID && responsePacket.Type == 0 {
			if sentEmptyPacket && responsePacket.Payload == "" {
				break
			}
			fullPayload.WriteString(responsePacket.Payload)
			if !sentEmptyPacket {
				emptyPacket := BuildPacket(commandID, 0, "")
				c.Config.Logger.Debug("Sending empty packet to confirm response fully received")
				_, err = c.Conn.Write(emptyPacket)
				if err != nil {
					return fullPayload.String(), fmt.Errorf("error sending empty packet: %s", err.Error())
				}
				sentEmptyPacket = true
			}
		} else {
			return fullPayload.String(), fmt.Errorf("unhandled case in Send")
		}
	}

	return fullPayload.String(), nil
}

// BuildPacket creates a binary RCON packet
func BuildPacket(id int32, packetType int32, payload string) []byte {
	payloadBytes := []byte(payload)
	payloadBytes = append(payloadBytes, 0x00)      // Null terminator for the payload
	packetSize := int32(4 + 4 + len(payloadBytes)) // ID + Type + Payload
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, packetSize)
	binary.Write(buffer, binary.LittleEndian, id)
	binary.Write(buffer, binary.LittleEndian, packetType)
	buffer.Write(payloadBytes)
	return buffer.Bytes()
}

// ReadPacket reads a packet from the connection
func (c *RconClient) ReadPacket() (*RconPacket, error) {
	c.Conn.SetReadDeadline(time.Now().Add(c.Config.Timeout))

	sizeBytes := make([]byte, 4)
	_, err := io.ReadFull(c.Conn, sizeBytes)
	if err != nil {
		c.Config.Logger.Error("Error reading packet size", "error", err)
		return nil, err
	}
	packetSize := int32(binary.LittleEndian.Uint32(sizeBytes))

	packetBytes := make([]byte, packetSize)
	_, err = io.ReadFull(c.Conn, packetBytes)
	if err != nil {
		c.Config.Logger.Error("Error reading packet data", "error", err)
		return nil, err
	}

	if len(packetBytes) < 10 {
		return nil, fmt.Errorf("packet too short")
	}
	packetID := int32(binary.LittleEndian.Uint32(packetBytes[0:4]))
	packetType := int32(binary.LittleEndian.Uint32(packetBytes[4:8]))
	payload := string(packetBytes[8 : len(packetBytes)-2]) // Exclude the two null terminators

	packet := &RconPacket{
		Size:    packetSize,
		ID:      packetID,
		Type:    packetType,
		Payload: payload,
	}
	c.Config.Logger.Debug("Received packet", "id", packetID, "type", packetType, "payload", payload)
	return packet, nil
}
