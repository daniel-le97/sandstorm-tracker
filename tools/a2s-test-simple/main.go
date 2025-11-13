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
	A2S_INFO      = 0x54
	S2A_INFO      = 0x49
	A2S_PLAYER    = 0x55
	S2A_PLAYER    = 0x44
	A2S_RULES     = 0x56
	S2A_RULES     = 0x45
	S2A_CHALLENGE = 0x41
)

func main() {
	address := "127.0.0.1:27131"
	outputFile := "a2s_response.txt"

	// Allow address to be specified via command line
	if len(os.Args) > 1 {
		address = os.Args[1]
	}
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	fmt.Println("Testing A2S Queries for Insurgency: Sandstorm")
	fmt.Printf("Server: %s\n", address)
	fmt.Printf("Output file: %s\n", outputFile)
	fmt.Println("=============================================================")
	fmt.Println()

	// Create output buffer
	var output bytes.Buffer
	output.WriteString(fmt.Sprintf("A2S Query Results for %s\n", address))
	output.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format(time.RFC3339)))
	output.WriteString("=============================================================\n\n")

	// Test 1: Server Info Query
	fmt.Println("Test 1: A2S_INFO (Server Information)")
	infoResult := testServerInfo(address)
	output.WriteString("Test 1: A2S_INFO (Server Information)\n")
	output.WriteString(infoResult)
	output.WriteString("\n")
	output.WriteString(string(make([]byte, 70)))
	output.WriteString("\n\n")

	fmt.Println()
	fmt.Println(string(make([]byte, 70)))
	fmt.Println()

	// Test 2: Standard A2S with challenge
	fmt.Println("Test 2: A2S_PLAYER with challenge request")
	playerResult := testWithChallenge(address)
	output.WriteString("Test 2: A2S_PLAYER with challenge request\n")
	output.WriteString(playerResult)
	output.WriteString("\n")
	output.WriteString(string(make([]byte, 70)))
	output.WriteString("\n\n")

	fmt.Println()
	fmt.Println(string(make([]byte, 70)))
	fmt.Println()

	// Test 3: Query with -1 challenge directly
	fmt.Println("Test 4: Player query with -1 challenge (no challenge request)")
	minusOneResult := testWithMinusOne(address)
	output.WriteString("Test 4: Player query with -1 challenge (no challenge request)\n")
	output.WriteString(minusOneResult)
	output.WriteString("\n")
	output.WriteString(string(make([]byte, 70)))
	output.WriteString("\n\n")

	fmt.Println()
	fmt.Println(string(make([]byte, 70)))
	fmt.Println()

	// Test 4: Rules Query with challenge
	fmt.Println("Test 5: A2S_RULES (Server Rules/CVars)")
	rulesResult := testRules(address)
	output.WriteString("Test 5: A2S_RULES (Server Rules/CVars)\n")
	output.WriteString(rulesResult)

	// Write to file
	err := os.WriteFile(outputFile, output.Bytes(), 0644)
	if err != nil {
		fmt.Printf("\n❌ Failed to write output file: %v\n", err)
	} else {
		fmt.Printf("\n✅ Results written to %s\n", outputFile)
	}
}

func testServerInfo(address string) string {
	var result bytes.Buffer

	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to connect: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send A2S_INFO request
	msg := "  → Sending A2S_INFO request...\n"
	fmt.Print(msg)
	result.WriteString(msg)

	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_INFO)
	request.WriteString("Source Engine Query\x00")

	conn.Write(request.Bytes())

	// Read response
	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to read response: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}

	msg = fmt.Sprintf("  ← Received %d bytes\n", n)
	fmt.Print(msg)
	result.WriteString(msg)

	msg = fmt.Sprintf("  Raw response: %x\n", response[:min(50, n)])
	fmt.Print(msg)
	result.WriteString(msg)

	reader := bytes.NewReader(response[:n])

	// Skip header
	var header uint32
	binary.Read(reader, binary.LittleEndian, &header)

	// Read response type
	responseType, _ := reader.ReadByte()
	msg = fmt.Sprintf("  Response type: 0x%02x ('%c')\n", responseType, responseType)
	fmt.Print(msg)
	result.WriteString(msg)

	if responseType == S2A_INFO {
		msg = "✅ Success! Got server info\n"
		fmt.Print(msg)
		result.WriteString(msg)

		infoMsg := parseServerInfo(reader)
		fmt.Print(infoMsg)
		result.WriteString(infoMsg)
	} else {
		msg = fmt.Sprintf("❌ Expected S2A_INFO (0x49), got 0x%02x\n", responseType)
		fmt.Print(msg)
		result.WriteString(msg)
	}

	return result.String()
}

func parseServerInfo(reader *bytes.Reader) string {
	var result bytes.Buffer

	// Read protocol version
	protocol, _ := reader.ReadByte()
	result.WriteString(fmt.Sprintf("  Protocol: %d\n", protocol))

	// Read null-terminated strings
	readString := func() string {
		nameBytes := []byte{}
		for {
			b, err := reader.ReadByte()
			if err != nil || b == 0 {
				break
			}
			nameBytes = append(nameBytes, b)
		}
		return string(nameBytes)
	}

	name := readString()
	mapName := readString()
	folder := readString()
	game := readString()

	result.WriteString(fmt.Sprintf("  Server Name: %s\n", name))
	result.WriteString(fmt.Sprintf("  Map: %s\n", mapName))
	result.WriteString(fmt.Sprintf("  Folder: %s\n", folder))
	result.WriteString(fmt.Sprintf("  Game: %s\n", game))

	var appID uint16
	binary.Read(reader, binary.LittleEndian, &appID)
	result.WriteString(fmt.Sprintf("  App ID: %d\n", appID))

	var players, maxPlayers, bots byte
	binary.Read(reader, binary.LittleEndian, &players)
	binary.Read(reader, binary.LittleEndian, &maxPlayers)
	binary.Read(reader, binary.LittleEndian, &bots)

	result.WriteString(fmt.Sprintf("  Players: %d/%d (Bots: %d)\n", players, maxPlayers, bots))

	return result.String()
}

func testWithChallenge(address string) string {
	var result bytes.Buffer

	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to connect: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Step 1: Request challenge
	msg := "  → Sending challenge request...\n"
	fmt.Print(msg)
	result.WriteString(msg)

	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_PLAYER)
	binary.Write(request, binary.LittleEndian, int32(-1))

	conn.Write(request.Bytes())

	// Read challenge response
	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to read challenge response: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}

	msg = fmt.Sprintf("  ← Received %d bytes\n", n)
	fmt.Print(msg)
	result.WriteString(msg)

	msg = fmt.Sprintf("  Raw response: %x\n", response[:min(20, n)])
	fmt.Print(msg)
	result.WriteString(msg)

	reader := bytes.NewReader(response[:n])

	// Skip header
	var header uint32
	binary.Read(reader, binary.LittleEndian, &header)

	// Read response type
	responseType, _ := reader.ReadByte()
	msg = fmt.Sprintf("  Response type: 0x%02x ('%c')\n", responseType, responseType)
	fmt.Print(msg)
	result.WriteString(msg)

	// Some servers return player data directly without challenge
	if responseType == S2A_PLAYER {
		msg := "✅ Server returned player data directly (no challenge required)\n"
		fmt.Print(msg)
		result.WriteString(msg)

		playerMsg := parsePlayers(reader)
		fmt.Print(playerMsg)
		result.WriteString(playerMsg)
		return result.String()
	}

	if responseType != S2A_CHALLENGE {
		msg := fmt.Sprintf("❌ Expected S2A_CHALLENGE (0x41) or S2A_PLAYER (0x44), got 0x%02x\n", responseType)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}

	// Read challenge
	var challenge int32
	binary.Read(reader, binary.LittleEndian, &challenge)
	msg = fmt.Sprintf("  Challenge: %d (0x%08x)\n", challenge, challenge)
	fmt.Print(msg)
	result.WriteString(msg)

	// Step 2: Query with challenge
	msg = "  → Sending player query with challenge...\n"
	fmt.Print(msg)
	result.WriteString(msg)

	request2 := &bytes.Buffer{}
	binary.Write(request2, binary.LittleEndian, uint32(PACKET_HEADER))
	request2.WriteByte(A2S_PLAYER)
	binary.Write(request2, binary.LittleEndian, challenge)

	conn.Write(request2.Bytes())

	response2 := make([]byte, 1400)
	n2, err := conn.Read(response2)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to read player response: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}

	msg = fmt.Sprintf("  ← Received %d bytes\n", n2)
	fmt.Print(msg)
	result.WriteString(msg)

	msg = fmt.Sprintf("  Raw response: %x\n", response2[:min(50, n2)])
	fmt.Print(msg)
	result.WriteString(msg)

	reader2 := bytes.NewReader(response2[:n2])
	binary.Read(reader2, binary.LittleEndian, &header)
	playerResponseType, _ := reader2.ReadByte()
	msg = fmt.Sprintf("  Response type: 0x%02x ('%c')\n", playerResponseType, playerResponseType)
	fmt.Print(msg)
	result.WriteString(msg)

	if playerResponseType == S2A_PLAYER {
		msg := "✅ Success! Got player data\n"
		fmt.Print(msg)
		result.WriteString(msg)

		playerMsg := parsePlayers(reader2)
		fmt.Print(playerMsg)
		result.WriteString(playerMsg)
	} else {
		msg := fmt.Sprintf("❌ Expected S2A_PLAYER (0x44), got 0x%02x\n", playerResponseType)
		fmt.Print(msg)
		result.WriteString(msg)
	}

	return result.String()
}


func testWithMinusOne(address string) string {
	var result bytes.Buffer

	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to connect: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send player query with -1 (all bits set)
	msg := "  → Sending player query with challenge=-1...\n"
	fmt.Print(msg)
	result.WriteString(msg)

	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_PLAYER)
	binary.Write(request, binary.LittleEndian, int32(-1))

	conn.Write(request.Bytes())

	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to read response: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}

	msg = fmt.Sprintf("  ← Received %d bytes\n", n)
	fmt.Print(msg)
	result.WriteString(msg)

	msg = fmt.Sprintf("  Raw response: %x\n", response[:min(50, n)])
	fmt.Print(msg)
	result.WriteString(msg)

	reader := bytes.NewReader(response[:n])
	var header uint32
	binary.Read(reader, binary.LittleEndian, &header)
	responseType, _ := reader.ReadByte()
	msg = fmt.Sprintf("  Response type: 0x%02x ('%c')\n", responseType, responseType)
	fmt.Print(msg)
	result.WriteString(msg)

	if responseType == S2A_PLAYER {
		msg := "✅ Success! Got player data with -1 challenge\n"
		fmt.Print(msg)
		result.WriteString(msg)

		msg = fmt.Sprintf("  Response size: %d bytes (6 bytes = header only, >6 = has players)\n", n)
		fmt.Print(msg)
		result.WriteString(msg)

		playerMsg := parsePlayers(reader)
		fmt.Print(playerMsg)
		result.WriteString(playerMsg)
	} else if responseType == S2A_CHALLENGE {
		msg := "⚠️  Server sent challenge (try Test 1 instead)\n"
		fmt.Print(msg)
		result.WriteString(msg)
	} else {
		msg := fmt.Sprintf("❌ Unexpected response type: 0x%02x\n", responseType)
		fmt.Print(msg)
		result.WriteString(msg)
	}

	return result.String()
}

func testRules(address string) string {
	var result bytes.Buffer

	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to connect: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Step 1: Request challenge
	msg := "  → Sending challenge request...\n"
	fmt.Print(msg)
	result.WriteString(msg)

	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_RULES)
	binary.Write(request, binary.LittleEndian, int32(-1))

	conn.Write(request.Bytes())

	// Read challenge response
	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to read challenge response: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}

	msg = fmt.Sprintf("  ← Received %d bytes\n", n)
	fmt.Print(msg)
	result.WriteString(msg)

	reader := bytes.NewReader(response[:n])

	// Skip header
	var header uint32
	binary.Read(reader, binary.LittleEndian, &header)

	// Read response type
	responseType, _ := reader.ReadByte()
	msg = fmt.Sprintf("  Response type: 0x%02x ('%c')\n", responseType, responseType)
	fmt.Print(msg)
	result.WriteString(msg)

	// Some servers return rules directly without challenge
	if responseType == S2A_RULES {
		msg := "✅ Server returned rules directly (no challenge required)\n"
		fmt.Print(msg)
		result.WriteString(msg)

		rulesMsg := parseRules(reader)
		fmt.Print(rulesMsg)
		result.WriteString(rulesMsg)
		return result.String()
	}

	if responseType != S2A_CHALLENGE {
		msg := fmt.Sprintf("❌ Expected S2A_CHALLENGE (0x41) or S2A_RULES (0x45), got 0x%02x\n", responseType)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}

	// Read challenge
	var challenge int32
	binary.Read(reader, binary.LittleEndian, &challenge)
	msg = fmt.Sprintf("  Challenge: %d (0x%08x)\n", challenge, challenge)
	fmt.Print(msg)
	result.WriteString(msg)

	// Step 2: Query with challenge
	msg = "  → Sending rules query with challenge...\n"
	fmt.Print(msg)
	result.WriteString(msg)

	request2 := &bytes.Buffer{}
	binary.Write(request2, binary.LittleEndian, uint32(PACKET_HEADER))
	request2.WriteByte(A2S_RULES)
	binary.Write(request2, binary.LittleEndian, challenge)

	conn.Write(request2.Bytes())

	response2 := make([]byte, 4096) // Rules can be larger
	n2, err := conn.Read(response2)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to read rules response: %v\n", err)
		fmt.Print(msg)
		result.WriteString(msg)
		return result.String()
	}

	msg = fmt.Sprintf("  ← Received %d bytes\n", n2)
	fmt.Print(msg)
	result.WriteString(msg)

	msg = fmt.Sprintf("  Raw response: %x\n", response2[:min(50, n2)])
	fmt.Print(msg)
	result.WriteString(msg)

	reader2 := bytes.NewReader(response2[:n2])
	binary.Read(reader2, binary.LittleEndian, &header)
	rulesResponseType, _ := reader2.ReadByte()
	msg = fmt.Sprintf("  Response type: 0x%02x ('%c')\n", rulesResponseType, rulesResponseType)
	fmt.Print(msg)
	result.WriteString(msg)

	if rulesResponseType == S2A_RULES {
		msg := "✅ Success! Got rules data\n"
		fmt.Print(msg)
		result.WriteString(msg)

		rulesMsg := parseRules(reader2)
		fmt.Print(rulesMsg)
		result.WriteString(rulesMsg)
	} else {
		msg := fmt.Sprintf("❌ Expected S2A_RULES (0x45), got 0x%02x\n", rulesResponseType)
		fmt.Print(msg)
		result.WriteString(msg)
	}

	return result.String()
}

func parseRules(reader *bytes.Reader) string {
	var result bytes.Buffer

	var ruleCount uint16
	if err := binary.Read(reader, binary.LittleEndian, &ruleCount); err != nil {
		msg := fmt.Sprintf("  ❌ Failed to read rule count: %v\n", err)
		result.WriteString(msg)
		return result.String()
	}

	result.WriteString(fmt.Sprintf("  Rule count: %d\n", ruleCount))
	result.WriteString(fmt.Sprintf("  Remaining buffer: %d bytes\n", reader.Len()))

	if reader.Len() == 0 {
		result.WriteString("  ℹ️  No rule data in response\n")
		return result.String()
	}

	// Read null-terminated string helper
	readString := func() string {
		nameBytes := []byte{}
		for {
			b, err := reader.ReadByte()
			if err != nil || b == 0 {
				break
			}
			nameBytes = append(nameBytes, b)
		}
		return string(nameBytes)
	}

	result.WriteString("  📋 Server Rules/CVars:\n")
	actualCount := 0
	for reader.Len() > 0 && actualCount < int(ruleCount) {
		name := readString()
		value := readString()

		if name == "" {
			break
		}

		result.WriteString(fmt.Sprintf("     %s = %s\n", name, value))
		actualCount++
	}

	if actualCount > 0 {
		result.WriteString(fmt.Sprintf("  ✅ Successfully parsed %d rules\n", actualCount))
	}

	return result.String()
}

func parsePlayers(reader *bytes.Reader) string {
	var result bytes.Buffer

	var playerCount byte
	if err := binary.Read(reader, binary.LittleEndian, &playerCount); err != nil {
		msg := fmt.Sprintf("  ❌ Failed to read player count: %v\n", err)
		result.WriteString(msg)
		return result.String()
	}

	result.WriteString(fmt.Sprintf("  Player count (reported): %d\n", playerCount))
	result.WriteString(fmt.Sprintf("  Remaining buffer: %d bytes\n", reader.Len()))

	if reader.Len() == 0 {
		result.WriteString("  ℹ️  No player data in response (server is empty)\n")
		return result.String()
	}

	// NOTE: Insurgency may report 0 for player count even when data follows
	// So we iterate until buffer is empty rather than trusting the count
	result.WriteString("  📋 Players found:\n")
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

		result.WriteString(fmt.Sprintf("     [%d] %s - Score: %d, Time: %.1fs\n", index, name, score, duration))
		actualCount++
	}

	if actualCount > 0 {
		result.WriteString(fmt.Sprintf("  ✅ Successfully parsed %d players\n", actualCount))
		if actualCount != int(playerCount) {
			result.WriteString(fmt.Sprintf("  ⚠️  INSURGENCY BUG: reported count=%d, actual=%d (this is why we iterate!)\n", playerCount, actualCount))
		}
	}

	return result.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
