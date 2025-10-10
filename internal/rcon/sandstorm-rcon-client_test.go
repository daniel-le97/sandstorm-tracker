package rcon

// import (
// 	"bytes"
// 	"net"
// 	"strings"
// 	"testing"
// 	"time"
// )

// type mockConn struct {
// 	readBuf           *bytes.Buffer
// 	writeBuf          *bytes.Buffer
// 	closed            bool
// 	setDeadlineCalled bool
// }

// func newMockConn(readData []byte) *mockConn {
// 	return &mockConn{
// 		readBuf:  bytes.NewBuffer(readData),
// 		writeBuf: &bytes.Buffer{},
// 	}
// }

// func (m *mockConn) Read(b []byte) (int, error) {
// 	return m.readBuf.Read(b)
// }
// func (m *mockConn) Write(b []byte) (int, error) {
// 	return m.writeBuf.Write(b)
// }
// func (m *mockConn) Close() error {
// 	m.closed = true
// 	return nil
// }
// func (m *mockConn) LocalAddr() net.Addr                { return nil }
// func (m *mockConn) RemoteAddr() net.Addr               { return nil }
// func (m *mockConn) SetDeadline(t time.Time) error      { m.setDeadlineCalled = true; return nil }
// func (m *mockConn) SetReadDeadline(t time.Time) error  { m.setDeadlineCalled = true; return nil }
// func (m *mockConn) SetWriteDeadline(t time.Time) error { m.setDeadlineCalled = true; return nil }

// // testLogger captures logs for assertions

// type testLogger struct {
// 	entries []string
// }

// func (l *testLogger) Debug(msg string, args ...any) { l.entries = append(l.entries, "DEBUG: "+msg) }
// func (l *testLogger) Info(msg string, args ...any)  { l.entries = append(l.entries, "INFO: "+msg) }
// func (l *testLogger) Warn(msg string, args ...any)  { l.entries = append(l.entries, "WARN: "+msg) }
// func (l *testLogger) Error(msg string, args ...any) { l.entries = append(l.entries, "ERROR: "+msg) }

// func TestBuildAndParsePacket(t *testing.T) {
// 	id := int32(42)
// 	typeVal := int32(2)
// 	payload := "hello world"
// 	packet := BuildPacket(id, typeVal, payload)
// 	if len(packet) < 12 {
// 		t.Fatalf("packet too short")
// 	}
// 	// Simulate reading the packet
// 	mock := newMockConn(packet)
// 	client := NewRconClient(mock, DefaultConfig())
// 	p, err := client.ReadPacket()
// 	if err != nil {
// 		t.Fatalf("ReadPacket error: %v", err)
// 	}
// 	if p.ID != id || p.Type != typeVal || p.Payload != payload {
// 		t.Errorf("Packet mismatch: got %+v", p)
// 	}
// }

// func TestAuthSuccess(t *testing.T) {
// 	idCounter = 0 // reset global counter for deterministic test
// 	id := int32(1)
// 	// Build a valid auth response packet
// 	resp := BuildPacket(id, 2, "")
// 	mock := newMockConn(resp)
// 	logger := &testLogger{}
// 	client := NewRconClient(mock, &ClientConfig{Timeout: time.Second, Logger: logger})
// 	ok := client.Auth("pw")
// 	if !ok {
// 		t.Error("expected auth to succeed")
// 	}
// 	if len(logger.entries) == 0 || logger.entries[len(logger.entries)-1] != "INFO: Authentication successful" {
// 		t.Error("expected success log entry")
// 	}
// }

// func TestAuthFailure(t *testing.T) {
// 	idCounter = 0
// 	id := int32(1)
// 	// Build a wrong type packet (not type 2)
// 	resp := BuildPacket(id, 0, "fail")
// 	mock := newMockConn(resp)
// 	logger := &testLogger{}
// 	client := NewRconClient(mock, &ClientConfig{Timeout: time.Second, Logger: logger})
// 	ok := client.Auth("pw")
// 	if ok {
// 		t.Error("expected auth to fail")
// 	}
// 	found := false
// 	for _, e := range logger.entries {
// 		if e == "ERROR: Authentication failed" {
// 			found = true
// 		}
// 	}
// 	if !found {
// 		t.Error("expected failure log entry")
// 	}
// }

// func TestSendCommandAndResponse(t *testing.T) {
// 	idCounter = 0
// 	id := int32(1)
// 	// Simulate a response: first a payload, then an empty packet
// 	resp := append(BuildPacket(id, 0, "result"), BuildPacket(id, 0, "")...)
// 	mock := newMockConn(resp)
// 	logger := &testLogger{}
// 	client := NewRconClient(mock, &ClientConfig{Timeout: time.Second, Logger: logger})
// 	out, err := client.Send("status")
// 	if err != nil {
// 		t.Fatalf("Send error: %v", err)
// 	}
// 	if out != "result" {
// 		t.Errorf("expected 'result', got '%s'", out)
// 	}
// }

// func TestSendHandlesUnexpectedID(t *testing.T) {
// 	idCounter = 0
// 	id := int32(1)
// 	// Simulate a packet with wrong ID, then a correct one
// 	wrong := BuildPacket(99, 0, "bad")
// 	good := append(BuildPacket(id, 0, "ok"), BuildPacket(id, 0, "")...)
// 	resp := append(wrong, good...)
// 	mock := newMockConn(resp)
// 	logger := &testLogger{}
// 	client := NewRconClient(mock, &ClientConfig{Timeout: time.Second, Logger: logger})
// 	out, err := client.Send("cmd")
// 	if err != nil {
// 		t.Fatalf("Send error: %v", err)
// 	}
// 	if out != "ok" {
// 		t.Errorf("expected 'ok', got '%s'", out)
// 	}
// 	found := false
// 	for _, e := range logger.entries {
// 		if e == "ERROR: Received packet with unexpected ID" {
// 			found = true
// 		}
// 	}
// 	if !found {
// 		t.Error("expected unexpected ID log entry")
// 	}
// }

// func TestReadPacketShort(t *testing.T) {
// 	mock := newMockConn([]byte{0, 0, 0, 0}) // size=0, so no payload
// 	client := NewRconClient(mock, DefaultConfig())
// 	_, err := client.ReadPacket()
// 	if err == nil {
// 		t.Error("expected error for short packet")
// 	}
// }

// func TestReadPacketIOError(t *testing.T) {
// 	mock := newMockConn([]byte{}) // nothing to read
// 	client := NewRconClient(mock, DefaultConfig())
// 	_, err := client.ReadPacket()
// 	if err == nil {
// 		t.Error("expected error for io.EOF")
// 	}
// }

// func TestBuildAndParsePacket_SpecialCharacters(t *testing.T) {
// 	id := int32(99)
// 	typeVal := int32(2)
// 	payload := "hélo 🌟 你好 \u2603 \x00"
// 	packet := BuildPacket(id, typeVal, payload)
// 	mock := newMockConn(packet)
// 	client := NewRconClient(mock, DefaultConfig())
// 	p, err := client.ReadPacket()
// 	if err != nil {
// 		t.Fatalf("ReadPacket error: %v", err)
// 	}
// 	if p.Payload != payload {
// 		t.Errorf("expected payload '%s', got '%s'", payload, p.Payload)
// 	}
// }

// func TestBuildAndParsePacket_LargePayload(t *testing.T) {
// 	id := int32(100)
// 	typeVal := int32(2)
// 	large := strings.Repeat("A", 4096) + "Ω"
// 	packet := BuildPacket(id, typeVal, large)
// 	mock := newMockConn(packet)
// 	client := NewRconClient(mock, DefaultConfig())
// 	p, err := client.ReadPacket()
// 	if err != nil {
// 		t.Fatalf("ReadPacket error: %v", err)
// 	}
// 	if p.Payload != large {
// 		t.Errorf("expected large payload, got len=%d", len(p.Payload))
// 	}
// }

// func TestBuildAndParsePacket_NullBytesInPayload(t *testing.T) {
// 	id := int32(101)
// 	typeVal := int32(2)
// 	payload := "foo\x00bar"
// 	packet := BuildPacket(id, typeVal, payload)
// 	mock := newMockConn(packet)
// 	client := NewRconClient(mock, DefaultConfig())
// 	p, err := client.ReadPacket()
// 	if err != nil {
// 		t.Fatalf("ReadPacket error: %v", err)
// 	}
// 	if p.Payload != payload {
// 		t.Errorf("expected payload with null byte, got '%v'", []byte(p.Payload))
// 	}
// }

// func TestSendAndPrint_EmptyResponse(t *testing.T) {
// 	idCounter = 0
// 	id := int32(1)
// 	resp := append(BuildPacket(id, 0, ""), BuildPacket(id, 0, "")...)
// 	mock := newMockConn(resp)
// 	logger := &testLogger{}
// 	client := NewRconClient(mock, &ClientConfig{Timeout: time.Second, Logger: logger})
// 	err := client.SendAndPrint("empty")
// 	if err != nil {
// 		t.Fatalf("SendAndPrint error: %v", err)
// 	}
// 	found := false
// 	for _, e := range logger.entries {
// 		if e == "INFO: Server response empty" {
// 			found = true
// 		}
// 	}
// 	if !found {
// 		t.Error("expected empty response log entry")
// 	}
// }

// func TestBuildPacket_RoundTripIntegrity(t *testing.T) {
// 	id := int32(123)
// 	typeVal := int32(2)
// 	payloads := []string{"abc", "héllo 🌟", strings.Repeat("Z", 1000), "\x00\x01\x02"}
// 	for _, payload := range payloads {
// 		packet := BuildPacket(id, typeVal, payload)
// 		mock := newMockConn(packet)
// 		client := NewRconClient(mock, DefaultConfig())
// 		p, err := client.ReadPacket()
// 		if err != nil {
// 			t.Fatalf("ReadPacket error: %v", err)
// 		}
// 		if p.Payload != payload {
// 			t.Errorf("round-trip failed: want '%v', got '%v'", []byte(payload), []byte(p.Payload))
// 		}
// 	}
// }
