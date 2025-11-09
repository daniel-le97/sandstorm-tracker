package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const (
	PACKET_HEADER = 0xFFFFFFFF
	A2S_PLAYER    = 0x55
)

func main() {
	address := "127.0.0.1:27131"

	fmt.Println("Testing A2S_PLAYER query with detailed hex dump")
	fmt.Println("=================================================\n")

	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send player query with -1 challenge
	fmt.Println("→ Sending A2S_PLAYER with challenge=-1")
	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_PLAYER)
	binary.Write(request, binary.LittleEndian, int32(-1))

	fmt.Printf("  Request bytes: %x\n", request.Bytes())

	conn.Write(request.Bytes())

	// Read response
	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("\n← Received %d bytes\n", n)
	fmt.Printf("  Full hex dump:\n")

	// Print in rows of 16 bytes
	for i := 0; i < n; i += 16 {
		end := i + 16
		if end > n {
			end = n
		}

		// Print offset
		fmt.Printf("  %04x: ", i)

		// Print hex
		for j := i; j < end; j++ {
			fmt.Printf("%02x ", response[j])
		}

		// Padding if last row
		for j := end; j < i+16; j++ {
			fmt.Print("   ")
		}

		// Print ASCII
		fmt.Print(" |")
		for j := i; j < end; j++ {
			if response[j] >= 32 && response[j] <= 126 {
				fmt.Printf("%c", response[j])
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")
	}

	// Parse header and response type
	fmt.Println("\n  Parsed structure:")
	reader := bytes.NewReader(response[:n])

	var header uint32
	binary.Read(reader, binary.LittleEndian, &header)
	fmt.Printf("  Header: 0x%08x\n", header)

	responseType, _ := reader.ReadByte()
	fmt.Printf("  Response Type: 0x%02x ('%c')\n", responseType, responseType)

	if responseType == 0x44 { // S2A_PLAYER
		playerCount, _ := reader.ReadByte()
		fmt.Printf("  Player Count: %d\n", playerCount)

		fmt.Printf("\n  Remaining bytes: %d\n", reader.Len())
		if reader.Len() > 0 {
			remaining := make([]byte, reader.Len())
			reader.Read(remaining)
			fmt.Printf("  Remaining data: %x\n", remaining)
		}
	}
}
