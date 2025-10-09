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
	"sync"
	"time"
)

// PocketBaseLogger interface matching PocketBase's app.Logger() methods
type PocketBaseLogger interface {
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

// Config struct
type ClientConfig struct {
	Timeout time.Duration
}

func Get() *ClientConfig {
	return &ClientConfig{
		Timeout: 5 * time.Second,
	}
}

type RconPacket struct {
	Size    int32
	ID      int32
	Type    int32
	Payload string
}

var (
	conf                       = Get()
	l         PocketBaseLogger = NoOpLogger{}
	idCounter int32            = 0
	idMutex   sync.Mutex
)

func generateID() int32 {
	idMutex.Lock()
	defer idMutex.Unlock()
	idCounter++
	return idCounter
}

func SetLogger(logger PocketBaseLogger) {
	l = logger
}

func Auth(conn net.Conn, password string) bool {
	authID := generateID()
	// Send the authentication packet.
	authPacket := BuildPacket(authID, 3, password)
	_, err := conn.Write(authPacket)
	if err != nil {
		l.Error("Error sending authentication packet", "error", err)
		return false
	}
	// Wait for the response.
	responsePacket := ReadPacket(conn)
	if responsePacket == nil {
		l.Error("Error reading authentication response")
		return false
	}
	if responsePacket.ID == authID && responsePacket.Type == 2 {
		l.Info("Authentication successful")
		return true
	} else {
		l.Error("Authentication failed")
		return false
	}
}

func Repl(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	prompt := fmt.Sprintf("RCON %s] ", conn.RemoteAddr().String())

	fmt.Print(prompt)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			l.Error("Error reading input", "error", err)
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

		err = SendAndPrint(conn, input)
		if err != nil {
			l.Error("Error executing command", "error", err)
		}

		fmt.Print(prompt)
	}
}

func SendAndPrint(conn net.Conn, command string) error {
	output, err := Send(conn, command)
	if err != nil {
		return fmt.Errorf("error sending command: %s", err.Error())
	}
	// Print output
	if strings.TrimSpace(output) == "" {
		l.Info("Server response empty")
		return nil
	}
	l.Info("Server response:")
	fmt.Printf("%s\n", output)
	return nil
}

func Send(conn net.Conn, command string) (string, error) {
	// Generate unique IDs for the command and empty packets
	commandID := generateID()

	// Send the command packet
	commandPacket := BuildPacket(commandID, 2, command)
	l.Debug("Sending command", "command", command)
	_, err := conn.Write(commandPacket)
	if err != nil {
		return "", fmt.Errorf("error sending command packet: %s", err.Error())
	}

	// Read and assemble response packets
	var (
		fullPayload     strings.Builder
		sentEmptyPacket bool
	)
	for {
		responsePacket := ReadPacket(conn)
		if responsePacket == nil {
			return "", fmt.Errorf("error reading command response")
		}

		if commandID != responsePacket.ID {
			l.Error("Received packet with unexpected ID", "id", responsePacket.ID)
			continue
		}

		if responsePacket.ID == commandID && responsePacket.Type == 0 {
			// Check if we've received the empty response packet
			if sentEmptyPacket && responsePacket.Payload == "" {
				break
			}
			fullPayload.WriteString(responsePacket.Payload)
			if !sentEmptyPacket {
				// Send the empty SERVERDATA_RESPONSE_VALUE packet
				emptyPacket := BuildPacket(commandID, 0, "")
				l.Debug("Sending empty packet to confirm response fully received")
				_, err = conn.Write(emptyPacket)
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

func BuildPacket(id int32, packetType int32, payload string) []byte {
	l.Debug("Building packet", "id", id, "type", packetType, "payload", payload)
	payloadBytes := []byte(payload)
	payloadBytes = append(payloadBytes, 0x00, 0x00) // Two null terminators for RCON protocol
	packetSize := int32(4 + 4 + len(payloadBytes))  // ID + Type + Payload
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, packetSize)
	binary.Write(buffer, binary.LittleEndian, id)
	binary.Write(buffer, binary.LittleEndian, packetType)
	buffer.Write(payloadBytes)
	return buffer.Bytes()
}

func ReadPacket(conn net.Conn) *RconPacket {
	conn.SetReadDeadline(time.Now().Add(conf.Timeout))

	// Read the packet size
	sizeBytes := make([]byte, 4)
	_, err := io.ReadFull(conn, sizeBytes)
	if err != nil {
		l.Error("Error reading packet size", "error", err)
		return nil
	}
	packetSize := int32(binary.LittleEndian.Uint32(sizeBytes))

	// Validate packet size to prevent excessive memory allocation
	if packetSize < 0 || packetSize > 16384 {
		l.Error("Invalid packet size", "size", packetSize)
		return nil
	}

	// Read the rest of the packet
	packetBytes := make([]byte, packetSize)
	_, err = io.ReadFull(conn, packetBytes)
	if err != nil {
		l.Error("Error reading packet data", "error", err)
		return nil
	}

	// Validate minimum packet size (ID + Type + at least one null terminator)
	if len(packetBytes) < 9 {
		l.Error("Packet too small", "size", len(packetBytes))
		return nil
	}

	// Parse the packet
	packetID := int32(binary.LittleEndian.Uint32(packetBytes[0:4]))
	packetType := int32(binary.LittleEndian.Uint32(packetBytes[4:8]))
	payload := string(packetBytes[8 : len(packetBytes)-2]) // Exclude the two null terminators

	packet := &RconPacket{
		Size:    packetSize,
		ID:      packetID,
		Type:    packetType,
		Payload: payload,
	}
	l.Debug("Received packet", "id", packetID, "type", packetType, "payload", payload)
	return packet
}
