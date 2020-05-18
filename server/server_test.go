package server

import (
	"net"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"

	"github.com/yondero/smtp/storage"
)

func connect(t *testing.T) (*textproto.Conn, *Session) {
	client, server := net.Pipe()

	t.Cleanup(func() {
		client.Close()
		server.Close()
	})

	text := textproto.NewConn(client)
	storage := storage.NewMemoryStorage()
	session := NewSession(server, storage)

	go session.Serve()

	_, _, err := text.ReadResponse(220)
	if err != nil {
		t.Errorf("Failed to receive greeting!")
	}

	return text, session
}

func command(t *testing.T, text *textproto.Conn, cmd string, code int) string {
	err := text.PrintfLine(cmd)
	if err != nil {
		t.Errorf("Error sending cmd=%s!", cmd)
	}

	_, message, err := text.ReadResponse(code)
	if err != nil {
		t.Errorf("Error receiving cmd=%s!", cmd)
	}

	return message
}

func TestExitsAfterTooManyErrors(t *testing.T) {
	text, _ := connect(t)

	command(t, text, "ERROR", 500)

	command(t, text, "ERROR", 500)

	command(t, text, "ERROR", 221)
}

func TestHELO(t *testing.T) {
	text, session := connect(t)

	rcpt1 := &mail.Address{Address: "user@example.com"}
	rcpt2 := &mail.Address{Address: "user@example.com"}

	session.Sender = &mail.Address{Address: "user@example.com"}
	session.Rcpts = []*mail.Address{rcpt1, rcpt2}

	message := command(t, text, "HELO", 250)
	if !strings.Contains(message, "yondero.co") {
		t.Errorf("Missing host!")
	}

	if session.Sender != nil {
		t.Errorf("Sender should be nil!")
	}

	if session.Rcpts != nil {
		t.Errorf("Rcpts should be nil!")
	}
}

func TestEHLO(t *testing.T) {
	text, session := connect(t)

	rcpt1 := &mail.Address{Address: "user@example.com"}
	rcpt2 := &mail.Address{Address: "user@example.com"}

	session.Sender = &mail.Address{Address: "user@example.com"}
	session.Rcpts = []*mail.Address{rcpt1, rcpt2}

	message := command(t, text, "EHLO", 250)
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
}

func TestMAIL(t *testing.T) {
	text, session := connect(t)

	command(t, text, "MAIL", 500)
	if session.Sender != nil {
		t.Errorf("Sender should not be set!")
	}

	command(t, text, "MAIL FROM:<invalid>", 500)
	if session.Sender != nil {
		t.Errorf("Sender should not be set!")
	}

	command(t, text, "MAIL FROM:<>", 250)
	if session.Sender != nil {
		t.Errorf("Sender should not be set!")
	}

	command(t, text, "MAIL FROM:<user@example.com>", 250)
	if session.Sender == nil {
		t.Errorf("Sender not set!")
	}
}

func TestRCPT(t *testing.T) {
	text, session := connect(t)

	command(t, text, "RCPT", 500)
	if session.Rcpts != nil {
		t.Errorf("Rcpts should not be set!")
	}

	command(t, text, "RCPT TO:<invalid>", 500)
	if session.Rcpts != nil {
		t.Errorf("Rcpts should not be set!")
	}

	command(t, text, "RCPT TO:<user@example.com>", 250)
	if session.Rcpts == nil {
		t.Errorf("Rcpts not set!")
	}
}

func TestDATA(t *testing.T) {
	text, session := connect(t)

	command(t, text, "DATA", 500)

	rcpt := &mail.Address{Address: "user@example.com"}
	session.Rcpts = []*mail.Address{rcpt}

	command(t, text, "DATA", 354)

	command(t, text, "Hello.\r\n.", 250)
}

func TestRSET(t *testing.T) {
	text, session := connect(t)

	rcpt1 := &mail.Address{Address: "user@example.com"}
	rcpt2 := &mail.Address{Address: "user@example.com"}

	session.Sender = &mail.Address{Address: "user@example.com"}
	session.Rcpts = []*mail.Address{rcpt1, rcpt2}

	command(t, text, "RSET", 250)
	if session.Sender != nil {
		t.Errorf("Sender should be nil!")
	}

	if session.Rcpts != nil {
		t.Errorf("Rcpts should be nil!")
	}
}

func TestNOOP(t *testing.T) {
	text, _ := connect(t)

	command(t, text, "NOOP", 250)
}

func TestQUIT(t *testing.T) {
	text, _ := connect(t)

	command(t, text, "QUIT", 221)
}
