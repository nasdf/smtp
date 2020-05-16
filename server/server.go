package server

import (
	"net"
	"net/mail"
	"net/textproto"
	"strings"
)

// Server contains server info.
type Server struct {
	Host string
	Exts []string
}

// NewServer returns a new Server using the host name as an identifier.
func NewServer(host string, exts ...string) *Server {
	return &Server{Host: host, Exts: exts}
}

// Listen listens for new connections.
func (s *Server) Listen() error {
	l, err := net.Listen("tcp", ":smtp")
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go NewSession(s, conn).Loop()
	}
}

// Session contains info about client connections
type Session struct {
	Conn   net.Conn
	Text   *textproto.Conn
	Server *Server
	Sender *mail.Address
	Rcpts  []*mail.Address
	Data   []string
}

// NewSession creates a new session
func NewSession(server *Server, conn net.Conn) *Session {
	text := textproto.NewConn(conn)
	return &Session{Conn: conn, Text: text, Server: server}
}

// loop handles session commands
func (s *Session) Loop() error {
	err := s.Text.PrintfLine("220 %s ESMTP service ready", s.Server.Host)
	if err != nil {
		return err
	}

	for {
		line, err := s.Text.ReadLine()
		if err != nil {
			return err
		}

		s.command(line)
	}

	return s.Text.Close()
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
		return s.reply(250, "noop ok")
	default:
		return s.reply(500, "invalid command")
	}

	return nil
}

// reply sends a multi line reply using the code and lines
func (s *Session) reply(code int, lines ...string) error {
	for i, line := range lines {
		var err error

		if i == len(lines)-1 {
			err = s.Text.PrintfLine("%d %s", code, line)
		} else {
			err = s.Text.PrintfLine("%d-%s", code, line)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// ehlo replies with host name and supported extensions
// sender, recipients, and data buffers are cleared
func (s *Session) ehlo(args []string) error {
	s.Sender = nil
	s.Rcpts = nil
	s.Data = nil

	lines := append([]string{s.Server.Host}, s.Server.Exts...)

	return s.reply(250, lines...)
}

// helo replies with host name
// sender, recipients, and data buffers are cleared
func (s *Session) helo(args []string) error {
	s.Sender = nil
	s.Rcpts = nil
	s.Data = nil

	return s.reply(250, s.Server.Host)
}

// mail sets the sender address
// it is possible for the sender address to be nil
func (s *Session) mail(args []string) error {
	if len(args) == 0 {
		return s.reply(500, "missing from")
	}

	from := strings.SplitN(args[0], ":", 2)
	if len(from) != 2 || !strings.EqualFold(from[0], "FROM") {
		return s.reply(500, "invalid from")
	}

	address, err := mail.ParseAddress(from[1])
	if err != nil && from[1] != "<>" {
		return s.reply(500, "invalid address")
	}

	s.Sender = address
	return s.reply(250, "sender ok")
}

// rcpt adds a recipient address
func (s *Session) rcpt(args []string) error {
	if len(args) == 0 {
		return s.reply(500, "missing to")
	}

	to := strings.SplitN(args[0], ":", 2)
	if len(to) != 2 || !strings.EqualFold(to[0], "TO") {
		return s.reply(500, "invalid to")
	}

	address, err := mail.ParseAddress(to[1])
	if err != nil {
		return s.reply(500, "invalid address")
	}

	s.Rcpts = append(s.Rcpts, address)
	return s.reply(250, "rcpt ok")
}

// data receives dot encoded lines
func (s *Session) data(args []string) error {
	if len(s.Rcpts) == 0 {
		return s.reply(500, "no recipients")
	}

	err := s.reply(354, "ready to receive data")
	if err != nil {
		return err
	}

	data, err := s.Text.ReadDotLines()
	if err != nil {
		return err
	}

	s.Data = data
	return s.reply(250, "data ok")
}

// quit closes the connection
func (s *Session) quit(args []string) error {
	err := s.reply(221, "quit ok")
	if err != nil {
		return err
	}

	return s.Text.Close()
}
