package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	PACKET_HEADER = 0xFFFFFFFF
	A2S_PLAYER    = 0x55
	S2A_PLAYER    = 0x44
	S2A_CHALLENGE = 0x41
)

func main() {
	address := "127.0.0.1:27131"

	// Allow address to be specified via command line
	if len(os.Args) > 1 {
		address = os.Args[1]
	}

	fmt.Println("Testing A2S Player Query Variations for Insurgency: Sandstorm")
	fmt.Printf("Server: %s\n", address)
	fmt.Println("=============================================================")
	fmt.Println()

	// Test 1: Standard A2S with challenge
	fmt.Println("Test 1: Standard A2S protocol with challenge request")
	testWithChallenge(address)

	fmt.Println()
	fmt.Println(string(make([]byte, 70)))
	fmt.Println()

	// Test 2: Direct query without challenge (some games accept this)
	fmt.Println("Test 2: Direct player query (no challenge)")
	testWithoutChallenge(address)

	fmt.Println()
	fmt.Println(string(make([]byte, 70)))
	fmt.Println()

	// Test 3: Query with -1 challenge directly
	fmt.Println("Test 3: Player query with -1 challenge (no challenge request)")
	testWithMinusOne(address)
}

func testWithChallenge(address string) {
	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Step 1: Request challenge
	fmt.Println("  → Sending challenge request...")
	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_PLAYER)
	binary.Write(request, binary.LittleEndian, int32(-1))

	conn.Write(request.Bytes())

	// Read challenge response
	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Printf("❌ Failed to read challenge response: %v\n", err)
		return
	}

	fmt.Printf("  ← Received %d bytes\n", n)
	fmt.Printf("  Raw response: %x\n", response[:min(20, n)])

	reader := bytes.NewReader(response[:n])

	// Skip header
	var header uint32
	binary.Read(reader, binary.LittleEndian, &header)

	// Read response type
	responseType, _ := reader.ReadByte()
	fmt.Printf("  Response type: 0x%02x ('%c')\n", responseType, responseType)

	if responseType != S2A_CHALLENGE {
		fmt.Printf("❌ Expected S2A_CHALLENGE (0x41), got 0x%02x\n", responseType)
		return
	}

	// Read challenge
	var challenge int32
	binary.Read(reader, binary.LittleEndian, &challenge)
	fmt.Printf("  Challenge: %d (0x%08x)\n", challenge, challenge)

	// Step 2: Query with challenge
	fmt.Println("  → Sending player query with challenge...")
	request2 := &bytes.Buffer{}
	binary.Write(request2, binary.LittleEndian, uint32(PACKET_HEADER))
	request2.WriteByte(A2S_PLAYER)
	binary.Write(request2, binary.LittleEndian, challenge)

	conn.Write(request2.Bytes())

	response2 := make([]byte, 1400)
	n2, err := conn.Read(response2)
	if err != nil {
		fmt.Printf("❌ Failed to read player response: %v\n", err)
		return
	}

	fmt.Printf("  ← Received %d bytes\n", n2)
	fmt.Printf("  Raw response: %x\n", response2[:min(50, n2)])

	reader2 := bytes.NewReader(response2[:n2])
	binary.Read(reader2, binary.LittleEndian, &header)
	playerResponseType, _ := reader2.ReadByte()
	fmt.Printf("  Response type: 0x%02x ('%c')\n", playerResponseType, playerResponseType)

	if playerResponseType == S2A_PLAYER {
		fmt.Println("✅ Success! Got player data")
		parsePlayers(reader2)
	} else {
		fmt.Printf("❌ Expected S2A_PLAYER (0x44), got 0x%02x\n", playerResponseType)
	}
}

func testWithoutChallenge(address string) {
	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send player query without challenge
	fmt.Println("  → Sending player query (no challenge)...")
	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_PLAYER)

	conn.Write(request.Bytes())

	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("  ← Received %d bytes\n", n)
	fmt.Printf("  Raw response: %x\n", response[:min(50, n)])

	reader := bytes.NewReader(response[:n])
	var header uint32
	binary.Read(reader, binary.LittleEndian, &header)
	responseType, _ := reader.ReadByte()
	fmt.Printf("  Response type: 0x%02x ('%c')\n", responseType, responseType)

	if responseType == S2A_PLAYER {
		fmt.Println("✅ Success! Got player data without challenge")
		parsePlayers(reader)
	} else if responseType == S2A_CHALLENGE {
		fmt.Println("⚠️  Server requires challenge (standard behavior)")
	} else {
		fmt.Printf("❌ Unexpected response type: 0x%02x\n", responseType)
	}
}

func testWithMinusOne(address string) {
	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send player query with -1 (all bits set)
	fmt.Println("  → Sending player query with challenge=-1...")
	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_PLAYER)
	binary.Write(request, binary.LittleEndian, int32(-1))

	conn.Write(request.Bytes())

	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("  ← Received %d bytes\n", n)
	fmt.Printf("  Raw response: %x\n", response[:min(50, n)])

	reader := bytes.NewReader(response[:n])
	var header uint32
	binary.Read(reader, binary.LittleEndian, &header)
	responseType, _ := reader.ReadByte()
	fmt.Printf("  Response type: 0x%02x ('%c')\n", responseType, responseType)

	if responseType == S2A_PLAYER {
		fmt.Println("✅ Success! Got player data with -1 challenge")
		fmt.Printf("  Response size: %d bytes (6 bytes = header only, >6 = has players)\n", n)
		parsePlayers(reader)
	} else if responseType == S2A_CHALLENGE {
		fmt.Println("⚠️  Server sent challenge (try Test 1 instead)")
	} else {
		fmt.Printf("❌ Unexpected response type: 0x%02x\n", responseType)
	}
}

func parsePlayers(reader *bytes.Reader) {
	var playerCount byte
	if err := binary.Read(reader, binary.LittleEndian, &playerCount); err != nil {
		fmt.Printf("  ❌ Failed to read player count: %v\n", err)
		return
	}

	fmt.Printf("  Player count (reported): %d\n", playerCount)
	fmt.Printf("  Remaining buffer: %d bytes\n", reader.Len())

	if reader.Len() == 0 {
		fmt.Println("  ℹ️  No player data in response (server is empty)")
		return
	}

	// NOTE: Insurgency may report 0 for player count even when data follows
	// So we iterate until buffer is empty rather than trusting the count
	fmt.Println("  📋 Players found:")
	actualCount := 0
	for reader.Len() > 0 && actualCount < 100 {
		var index byte
		if err := binary.Read(reader, binary.LittleEndian, &index); err != nil {
			break
		}

		// Read null-terminated string
		nameBytes := []byte{}
		for {
			b, err := reader.ReadByte()
			if err != nil || b == 0 {
				break
			}
			nameBytes = append(nameBytes, b)
		}
		name := string(nameBytes)

		var score int32
		if err := binary.Read(reader, binary.LittleEndian, &score); err != nil {
			break
		}

		var duration float32
		if err := binary.Read(reader, binary.LittleEndian, &duration); err != nil {
			break
		}

		fmt.Printf("     [%d] %s - Score: %d, Time: %.1fs\n", index, name, score, duration)
		actualCount++
	}

	if actualCount > 0 {
		fmt.Printf("  ✅ Successfully parsed %d players\n", actualCount)
		if actualCount != int(playerCount) {
			fmt.Printf("  ⚠️  INSURGENCY BUG: reported count=%d, actual=%d (this is why we iterate!)\n", playerCount, actualCount)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
