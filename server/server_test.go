package server

import (
	"net"
	"net/textproto"
	"testing"
)

func testConnect(t *testing.T) *textproto.Conn {
	go NewServer("localhost").Listen("localhost:3000")

	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		t.Errorf("Failed to connect!")
	}

	return textproto.NewConn(conn)
}

func testCommand(t *testing.T, text *textproto.Conn, cmd string, code int) {
	err := text.PrintfLine(cmd)
	if err != nil {
		t.Errorf("Failed to send!")
	}

	_, _, err = text.ReadResponse(code)
	if err != nil {
		t.Errorf("Failed to receive!")
	}
}

func TestSession(t *testing.T) {
	text := testConnect(t)

	_, _, err := text.ReadResponse(220)
	if err != nil {
		t.Errorf("Failed to receive greeting!")
	}

	testCommand(t, text, "EHLO", 250)
	testCommand(t, text, "HELO", 250)
	testCommand(t, text, "NOOP", 250)
	testCommand(t, text, "QUIT", 221)

	text.Close()
}
