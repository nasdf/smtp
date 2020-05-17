package server

import (
	"fmt"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
)

// Session contains info about client connections
type Session struct {
	Conn   net.Conn
	Text   *textproto.Conn
	Sender *mail.Address
	Rcpts  []*mail.Address
	Data   []string
}

// NewSession creates a new session
func NewSession(conn net.Conn) *Session {
	text := textproto.NewConn(conn)
	return &Session{Conn: conn, Text: text}
}

// Serve handles session commands
func (s *Session) Serve() error {
	err := s.Text.PrintfLine("220 yondero.co ESMTP service ready")
	if err != nil {
		return err
	}

	for {
		line, err := s.Text.ReadLine()
		if err != nil {
			return err
		}

		err = s.command(line)
		if err != nil {
			return err
		}
	}

	return nil
}

// comand parses and responds to an smtp command
func (s *Session) command(line string) error {
	line = strings.TrimRight(line, "\r\n")

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	switch strings.ToUpper(parts[0]) {
	case "EHLO":
		return s.ehlo(parts[1:])
	case "HELO":
		return s.helo(parts[1:])
	case "MAIL":
		return s.mail(parts[1:])
	case "RCPT":
		return s.rcpt(parts[1:])
	case "DATA":
		return s.data(parts[1:])
	case "QUIT":
		return s.quit(parts[1:])
	case "NOOP":
		return s.Text.PrintfLine("250 noop ok")
	}

	return s.Text.PrintfLine("500 invalid command")
}

// ehlo replies with host name and supported extensions
// sender, recipients, and data buffers are cleared
func (s *Session) ehlo(args []string) error {
	s.Sender = nil
	s.Rcpts = nil
	s.Data = nil

	var builder strings.Builder
	fmt.Fprintf(&builder, "250-yondero.co\r\n")
	fmt.Fprintf(&builder, "250-STARTTLS\r\n")
	fmt.Fprintf(&builder, "250 SIZE 10000000")

	return s.Text.PrintfLine(builder.String())
}

// helo replies with host name
// sender, recipients, and data buffers are cleared
func (s *Session) helo(args []string) error {
	s.Sender = nil
	s.Rcpts = nil
	s.Data = nil

	return s.Text.PrintfLine("250 yondero.co")
}

// mail sets the sender address
// it is possible for the sender address to be nil
func (s *Session) mail(args []string) error {
	if len(args) == 0 {
		return s.Text.PrintfLine("500 missing from")
	}

	from := strings.SplitN(args[0], ":", 2)
	if len(from) != 2 || !strings.EqualFold(from[0], "FROM") {
		return s.Text.PrintfLine("500 invalid from")
	}

	address, err := mail.ParseAddress(from[1])
	if err != nil && from[1] != "<>" {
		return s.Text.PrintfLine("500 invalid address")
	}

	s.Sender = address
	s.Rcpts = nil
	s.Data = nil

	return s.Text.PrintfLine("250 sender ok")
}

// rcpt adds a recipient address
func (s *Session) rcpt(args []string) error {
	if len(args) == 0 {
		return s.Text.PrintfLine("500 missing to")
	}

	to := strings.SplitN(args[0], ":", 2)
	if len(to) != 2 || !strings.EqualFold(to[0], "TO") {
		return s.Text.PrintfLine("500 invalid to")
	}

	address, err := mail.ParseAddress(to[1])
	if err != nil {
		return s.Text.PrintfLine("500 invalid address")
	}

	s.Rcpts = append(s.Rcpts, address)
	return s.Text.PrintfLine("250 rcpt ok")
}

// data receives dot encoded lines
func (s *Session) data(args []string) error {
	if len(s.Rcpts) == 0 {
		return s.Text.PrintfLine("500 no recipients")
	}

	err := s.Text.PrintfLine("354 ready to receive data")
	if err != nil {
		return err
	}

	data, err := s.Text.ReadDotLines()
	if err != nil {
		return err
	}

	s.Data = data
	return s.Text.PrintfLine("250 data ok")
}

// quit closes the connection
func (s *Session) quit(args []string) error {
	err := s.Text.PrintfLine("221 quit ok")
	if err != nil {
		return err
	}

	return s.Text.Close()
}
