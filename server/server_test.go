package main_test

import (
	"io"
	"net"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	// Wait for the server to start
	conn, err := net.Dial("tcp", "localhost:6969")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Test basic commands
	testCommands(t, conn)
}

func testCommands(t *testing.T, conn net.Conn) {
	commands := []string{
		"SET key1 value1",
		"GET key1",
		"DEL key1",
		"GET key1",
		"PING",
	}

	expectedResponses := []string{
		"OK",
		"value1",
		"OK",
		"Key not found",
		"PONG",
	}

	for i, cmd := range commands {
		if _, err := io.WriteString(conn, cmd); err != nil {
			t.Fatalf("Failed to send command '%s': %v", cmd, err)
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			t.Fatalf("Failed to read response for command '%s': %v", cmd, err)
		}

		response := strings.TrimSpace(string(buffer[:n]))
		if response != expectedResponses[i] {
			t.Errorf("Expected response '%s' for command '%s', got '%s'", expectedResponses[i], cmd, response)
		}
	}
}
