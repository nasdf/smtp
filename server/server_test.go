package server

import (
	"net"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"
)

func connect(t *testing.T) (*textproto.Conn, *Session) {
	client, server := net.Pipe()
	text := textproto.NewConn(client)

	session := NewSession(server)
	go session.Serve()

	_, _, err := text.ReadResponse(220)
	if err != nil {
		t.Errorf("Failed to receive greeting!")
	}

	return text, session
}

func command(t *testing.T, cmd string, code int, text *textproto.Conn) string {
	err := text.PrintfLine(cmd)
	if err != nil {
		t.Errorf("Failed to send!")
	}

	_, message, err := text.ReadResponse(code)
	if err != nil {
		t.Errorf("Failed to receive!")
	}

	return message
}

func TestHELO(t *testing.T) {
	text, session := connect(t)
	defer text.Close()

	rcpt1 := &mail.Address{Address: "user@example.com"}
	rcpt2 := &mail.Address{Address: "user@example.com"}

	session.Sender = &mail.Address{Address: "user@example.com"}
	session.Rcpts = []*mail.Address{rcpt1, rcpt2}
	session.Data = []string{"Hello."}

	message := command(t, "HELO", 250, text)
	if !strings.Contains(message, "yondero.co") {
		t.Errorf("Missing host!")
	}

	if session.Sender != nil {
		t.Errorf("Sender should be nil!")
	}

	if session.Rcpts != nil {
		t.Errorf("Rcpts should be nil!")
	}

	if session.Data != nil {
		t.Errorf("Data should be nil!")
	}
}

func TestEHLO(t *testing.T) {
	text, session := connect(t)
	defer text.Close()

	rcpt1 := &mail.Address{Address: "user@example.com"}
	rcpt2 := &mail.Address{Address: "user@example.com"}

	session.Sender = &mail.Address{Address: "user@example.com"}
	session.Rcpts = []*mail.Address{rcpt1, rcpt2}
	session.Data = []string{"Hello."}
	
	message := command(t, "EHLO", 250, text)
	if !strings.Contains(message, "yondero.co") {
		t.Errorf("Missing host!")
	}

	if !strings.Contains(message, "STARTTLS") {
		t.Errorf("Missing STARTTLS!")
	}

	if session.Sender != nil {
		t.Errorf("Sender should be nil!")
	}

	if session.Rcpts != nil {
		t.Errorf("Rcpts should be nil!")
	}

	if session.Data != nil {
		t.Errorf("Data should be nil!")
	}
}

func TestMAIL(t *testing.T) {
	text, session := connect(t)
	defer text.Close()

	command(t, "MAIL", 500, text)
	if session.Sender != nil {
		t.Errorf("Sender should not be set!")
	}

	command(t, "MAIL FROM:<invalid>", 500, text)
	if session.Sender != nil {
		t.Errorf("Sender should not be set!")
	}

	command(t, "MAIL FROM:<>", 250, text)
	if session.Sender != nil {
		t.Errorf("Sender should not be set!")
	}

	command(t, "MAIL FROM:<user@example.com>", 250, text)
	if session.Sender == nil {
		t.Errorf("Sender not set!")
	}
}

func TestRCPT(t *testing.T) {
	text, session := connect(t)
	defer text.Close()

	command(t, "RCPT", 500, text)
	if session.Rcpts != nil {
		t.Errorf("Rcpts should not be set!")
	}

	command(t, "RCPT TO:<invalid>", 500, text)
	if session.Rcpts != nil {
		t.Errorf("Rcpts should not be set!")
	}

	command(t, "RCPT TO:<user@example.com>", 250, text)
	if session.Rcpts == nil {
		t.Errorf("Rcpts not set!")
	}
}

func TestDATA(t *testing.T) {
	text, session := connect(t)
	defer text.Close()

	command(t, "DATA", 500, text)
	if session.Data != nil {
		t.Errorf("Data should not be set!")
	}

	rcpt := &mail.Address{Address: "user@example.com"}
	session.Rcpts = []*mail.Address{rcpt}

	command(t, "DATA", 354, text)
	if session.Data != nil {
		t.Errorf("Data should not be set!")
	}

	command(t, "Hello.\r\n.", 250, text)
	if session.Data == nil {
		t.Errorf("Data should be set!")
	}
}

func TestNOOP(t *testing.T) {
	text, _ := connect(t)
	defer text.Close()

	command(t, "NOOP", 250, text)
}

func TestQUIT(t *testing.T) {
	text, _ := connect(t)
	defer text.Close()

	command(t, "QUIT", 221, text)
}