package a2s

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// A2S Protocol constants
const (
	// Request headers
	A2S_INFO                     = 0x54 // 'T' - Request server info
	A2S_PLAYER                   = 0x55 // 'U' - Request player list
	A2S_RULES                    = 0x56 // 'V' - Request server rules
	A2S_SERVERQUERY_GETCHALLENGE = 0x57 // 'W' - Request challenge number

	// Response headers
	S2A_INFO_SRC      = 0x49 // 'I' - Server info response (Source engine)
	S2A_INFO_DETAILED = 0x6D // 'm' - Detailed server info (obsolete)
	S2A_PLAYER        = 0x44 // 'D' - Player list response
	S2A_RULES         = 0x45 // 'E' - Rules response
	S2A_CHALLENGE     = 0x41 // 'A' - Challenge response

	// Protocol
	PACKET_HEADER = 0xFFFFFFFF

	// Timeouts
	DEFAULT_TIMEOUT = 5 * time.Second
)

// Client represents an A2S query client
type Client struct {
	timeout time.Duration
}

// ServerInfo contains information about a Source engine server
type ServerInfo struct {
	Protocol    byte
	Name        string
	Map         string
	Folder      string
	Game        string
	ID          uint16
	Players     byte
	MaxPlayers  byte
	Bots        byte
	ServerType  byte
	Environment byte
	Visibility  byte
	VAC         byte
	Version     string

	// Extended Data Flag (EDF)
	Port         *uint16
	SteamID      *uint64
	SourceTVPort *uint16
	SourceTVName *string
	Keywords     *string
	GameID       *uint64
}

// Player represents a player on the server
type Player struct {
	Index    byte
	Name     string
	Score    int32
	Duration float32
}

// Rule represents a server configuration rule
type Rule struct {
	Name  string
	Value string
}

// NewClient creates a new A2S client with default timeout
func NewClient() *Client {
	return &Client{
		timeout: DEFAULT_TIMEOUT,
	}
}

// NewClientWithTimeout creates a new A2S client with custom timeout
func NewClientWithTimeout(timeout time.Duration) *Client {
	return &Client{
		timeout: timeout,
	}
}

// QueryInfo retrieves server information
func (c *Client) QueryInfo(address string) (*ServerInfo, error) {
	return c.QueryInfoContext(context.Background(), address)
}

// QueryInfoContext retrieves server information with context support
func (c *Client) QueryInfoContext(ctx context.Context, address string) (*ServerInfo, error) {
	conn, err := net.DialTimeout("udp", address, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(c.timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Check context before sending
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Build A2S_INFO request
	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_INFO)
	request.WriteString("Source Engine Query\x00")

	// Send request
	if _, err := conn.Write(request.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return parseServerInfo(response[:n])
}

// QueryPlayers retrieves the list of players on the server
func (c *Client) QueryPlayers(address string) ([]Player, error) {
	return c.QueryPlayersContext(context.Background(), address)
}

// QueryPlayersContext retrieves the list of players on the server with context support
func (c *Client) QueryPlayersContext(ctx context.Context, address string) ([]Player, error) {
	conn, err := net.DialTimeout("udp", address, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(c.timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Try direct query with challenge -1 first
	// Some games (like Insurgency: Sandstorm) skip the challenge-response and return player data directly
	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_PLAYER)
	binary.Write(request, binary.LittleEndian, int32(-1))

	if _, err := conn.Write(request.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to send initial request: %w", err)
	}

	// Read response
	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read initial response: %w", err)
	}

	// Check response type
	if n < 6 {
		return nil, fmt.Errorf("response too short: %d bytes", n)
	}

	// Skip header (4 bytes) and read response type
	responseType := response[4]

	// If we got player data directly (Insurgency: Sandstorm behavior), parse it
	if responseType == S2A_PLAYER {
		return parsePlayers(response[:n])
	}

	// If we got a challenge, use it for a second request (standard behavior)
	if responseType == S2A_CHALLENGE {
		// Parse the challenge number
		reader := bytes.NewReader(response[:n])
		var header uint32
		binary.Read(reader, binary.LittleEndian, &header)
		reader.ReadByte() // skip response type
		var challenge int32
		if err := binary.Read(reader, binary.LittleEndian, &challenge); err != nil {
			return nil, fmt.Errorf("failed to read challenge: %w", err)
		}

		// Build second request with the challenge
		request2 := &bytes.Buffer{}
		binary.Write(request2, binary.LittleEndian, uint32(PACKET_HEADER))
		request2.WriteByte(A2S_PLAYER)
		binary.Write(request2, binary.LittleEndian, challenge)

		if _, err := conn.Write(request2.Bytes()); err != nil {
			return nil, fmt.Errorf("failed to send challenge request: %w", err)
		}

		// Read player data response
		response2 := make([]byte, 1400)
		n2, err := conn.Read(response2)
		if err != nil {
			return nil, fmt.Errorf("failed to read player response: %w", err)
		}

		return parsePlayers(response2[:n2])
	}

	return nil, fmt.Errorf("unexpected response type: 0x%02x", responseType)
}

// QueryRules retrieves server rules/cvars
func (c *Client) QueryRules(address string) ([]Rule, error) {
	return c.QueryRulesContext(context.Background(), address)
}

// QueryRulesContext retrieves server rules/cvars with context support
func (c *Client) QueryRulesContext(ctx context.Context, address string) ([]Rule, error) {
	conn, err := net.DialTimeout("udp", address, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(c.timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// First, get challenge number
	challenge, err := c.getChallengeContext(ctx, conn, A2S_RULES)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}

	// Build A2S_RULES request with challenge
	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(A2S_RULES)
	binary.Write(request, binary.LittleEndian, challenge)

	// Send request
	if _, err := conn.Write(request.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return parseRules(response[:n])
}

// getChallenge requests a challenge number from the server
func (c *Client) getChallenge(conn net.Conn, queryType byte) (int32, error) {
	return c.getChallengeContext(context.Background(), conn, queryType)
}

// getChallengeContext requests a challenge number from the server with context support
func (c *Client) getChallengeContext(ctx context.Context, conn net.Conn, queryType byte) (int32, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Build challenge request
	request := &bytes.Buffer{}
	binary.Write(request, binary.LittleEndian, uint32(PACKET_HEADER))
	request.WriteByte(queryType)
	binary.Write(request, binary.LittleEndian, int32(-1)) // -1 to request challenge

	// Send request
	if _, err := conn.Write(request.Bytes()); err != nil {
		return 0, fmt.Errorf("failed to send challenge request: %w", err)
	}

	// Read response
	response := make([]byte, 1400)
	n, err := conn.Read(response)
	if err != nil {
		return 0, fmt.Errorf("failed to read challenge response: %w", err)
	}

	// Parse challenge
	reader := bytes.NewReader(response[:n])

	// Skip header
	var header uint32
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return 0, fmt.Errorf("failed to read header: %w", err)
	}

	// Read response type
	responseType, err := reader.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("failed to read response type: %w", err)
	}

	if responseType != S2A_CHALLENGE {
		return 0, fmt.Errorf("unexpected response type: %02x", responseType)
	}

	// Read challenge number
	var challenge int32
	if err := binary.Read(reader, binary.LittleEndian, &challenge); err != nil {
		return 0, fmt.Errorf("failed to read challenge: %w", err)
	}

	return challenge, nil
}

// parseServerInfo parses the server info response
func parseServerInfo(data []byte) (*ServerInfo, error) {
	reader := bytes.NewReader(data)
	info := &ServerInfo{}

	// Skip header
	var header uint32
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Read response type
	responseType, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read response type: %w", err)
	}

	if responseType != S2A_INFO_SRC {
		return nil, fmt.Errorf("unexpected response type: %02x", responseType)
	}

	// Read protocol version
	if info.Protocol, err = reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read protocol: %w", err)
	}

	// Read null-terminated strings
	if info.Name, err = readString(reader); err != nil {
		return nil, fmt.Errorf("failed to read name: %w", err)
	}
	if info.Map, err = readString(reader); err != nil {
		return nil, fmt.Errorf("failed to read map: %w", err)
	}
	if info.Folder, err = readString(reader); err != nil {
		return nil, fmt.Errorf("failed to read folder: %w", err)
	}
	if info.Game, err = readString(reader); err != nil {
		return nil, fmt.Errorf("failed to read game: %w", err)
	}

	// Read app ID
	if err := binary.Read(reader, binary.LittleEndian, &info.ID); err != nil {
		return nil, fmt.Errorf("failed to read ID: %w", err)
	}

	// Read player counts
	if info.Players, err = reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read players: %w", err)
	}
	if info.MaxPlayers, err = reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read max players: %w", err)
	}
	if info.Bots, err = reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read bots: %w", err)
	}

	// Read server type
	if info.ServerType, err = reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read server type: %w", err)
	}
	if info.Environment, err = reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read environment: %w", err)
	}
	if info.Visibility, err = reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read visibility: %w", err)
	}
	if info.VAC, err = reader.ReadByte(); err != nil {
		return nil, fmt.Errorf("failed to read VAC: %w", err)
	}

	// Read version
	if info.Version, err = readString(reader); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	// Read EDF (Extra Data Flag) if present
	if reader.Len() > 0 {
		edf, err := reader.ReadByte()
		if err != nil {
			return info, nil // Not critical, return what we have
		}

		if edf&0x80 != 0 { // Port
			var port uint16
			if err := binary.Read(reader, binary.LittleEndian, &port); err == nil {
				info.Port = &port
			}
		}
		if edf&0x10 != 0 { // SteamID
			var steamID uint64
			if err := binary.Read(reader, binary.LittleEndian, &steamID); err == nil {
				info.SteamID = &steamID
			}
		}
		if edf&0x40 != 0 { // SourceTV
			var port uint16
			if err := binary.Read(reader, binary.LittleEndian, &port); err == nil {
				info.SourceTVPort = &port
			}
			if name, err := readString(reader); err == nil {
				info.SourceTVName = &name
			}
		}
		if edf&0x20 != 0 { // Keywords
			if keywords, err := readString(reader); err == nil {
				info.Keywords = &keywords
			}
		}
		if edf&0x01 != 0 { // GameID
			var gameID uint64
			if err := binary.Read(reader, binary.LittleEndian, &gameID); err == nil {
				info.GameID = &gameID
			}
		}
	}

	return info, nil
}

// parsePlayers parses the player list response
func parsePlayers(data []byte) ([]Player, error) {
	reader := bytes.NewReader(data)

	// Skip header
	var header uint32
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Read response type
	responseType, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read response type: %w", err)
	}

	if responseType != S2A_PLAYER {
		return nil, fmt.Errorf("unexpected response type: %02x", responseType)
	}

	// Read player count
	// NOTE: Insurgency: Sandstorm has been observed to return 0 for player count
	// even when player data follows. We read the count but iterate until buffer is empty.
	playerCount, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read player count: %w", err)
	}

	players := make([]Player, 0, max(int(playerCount), 32)) // Allocate for at least 32 if count is 0

	// Iterate until buffer is exhausted (like SAW does)
	// This handles the case where Insurgency reports 0 players but sends data anyway
	for reader.Len() > 0 {
		player := Player{}

		// Read index
		if player.Index, err = reader.ReadByte(); err != nil {
			// End of buffer or malformed data
			break
		}

		// Read name
		if player.Name, err = readString(reader); err != nil {
			// End of buffer or malformed data
			break
		}

		// Read score
		if err := binary.Read(reader, binary.LittleEndian, &player.Score); err != nil {
			// End of buffer or malformed data
			break
		}

		// Read duration
		if err := binary.Read(reader, binary.LittleEndian, &player.Duration); err != nil {
			// End of buffer or malformed data
			break
		}

		players = append(players, player)
	}

	return players, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// parseRules parses the server rules response
func parseRules(data []byte) ([]Rule, error) {
	reader := bytes.NewReader(data)

	// Skip header
	var header uint32
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Read response type
	responseType, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read response type: %w", err)
	}

	if responseType != S2A_RULES {
		return nil, fmt.Errorf("unexpected response type: %02x", responseType)
	}

	// Read rule count
	var ruleCount uint16
	if err := binary.Read(reader, binary.LittleEndian, &ruleCount); err != nil {
		return nil, fmt.Errorf("failed to read rule count: %w", err)
	}

	rules := make([]Rule, 0, ruleCount)

	for i := uint16(0); i < ruleCount; i++ {
		rule := Rule{}

		// Read name
		if rule.Name, err = readString(reader); err != nil {
			return rules, fmt.Errorf("failed to read rule name: %w", err)
		}

		// Read value
		if rule.Value, err = readString(reader); err != nil {
			return rules, fmt.Errorf("failed to read rule value: %w", err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// readString reads a null-terminated string from the reader
func readString(reader io.Reader) (string, error) {
	var result []byte
	buf := make([]byte, 1)

	for {
		if _, err := reader.Read(buf); err != nil {
			return "", err
		}
		if buf[0] == 0 {
			break
		}
		result = append(result, buf[0])
	}

	return string(result), nil
}
