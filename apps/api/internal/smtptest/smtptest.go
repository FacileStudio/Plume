// Package smtptest is a one-connection-at-a-time SMTP server on the loopback
// interface that accepts every message and remembers the AUTH PLAIN password
// and the body of each. Tests point an owner's SMTP config at it.
package smtptest

import (
	"bufio"
	"encoding/base64"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// Server is the fake SMTP server.
type Server struct {
	listener  net.Listener
	mu        sync.Mutex
	passwords []string
	bodies    []string
}

// Start listens on a free loopback port and stops when the test ends.
func Start(t testing.TB) *Server {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("smtptest: listen: %v", err)
	}
	server := &Server{listener: listener}
	t.Cleanup(func() { listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.serve(conn)
		}
	}()
	return server
}

// Host is the loopback address to put in an SMTP config.
func (s *Server) Host() string { return "127.0.0.1" }

// Port is the port to put in an SMTP config.
func (s *Server) Port() int { return s.listener.Addr().(*net.TCPAddr).Port }

// Passwords are the AUTH PLAIN passwords received, in order.
func (s *Server) Passwords() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.passwords...)
}

// WaitBodies blocks until n messages arrived and returns their bodies.
func (s *Server) WaitBodies(t testing.TB, n int) []string {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		if len(s.bodies) >= n {
			bodies := append([]string(nil), s.bodies...)
			s.mu.Unlock()
			return bodies
		}
		s.mu.Unlock()
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("smtptest: %d message(s) did not arrive", n)
	return nil
}

func (s *Server) serve(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	reply := func(line string) { conn.Write([]byte(line + "\r\n")) }
	reply("220 smtptest ready")
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		command := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(command, "EHLO"):
			reply("250-smtptest")
			reply("250 AUTH PLAIN")
		case strings.HasPrefix(command, "AUTH PLAIN"):
			s.recordPassword(strings.TrimSpace(line))
			reply("235 ok")
		case command == "DATA":
			reply("354 go ahead")
			s.recordBody(reader)
			reply("250 queued")
		case command == "QUIT":
			reply("221 bye")
			return
		default:
			reply("250 ok")
		}
	}
}

func (s *Server) recordPassword(line string) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return
	}
	raw, err := base64.StdEncoding.DecodeString(fields[2])
	if err != nil {
		return
	}
	parts := strings.Split(string(raw), "\x00")
	s.mu.Lock()
	s.passwords = append(s.passwords, parts[len(parts)-1])
	s.mu.Unlock()
}

func (s *Server) recordBody(reader *bufio.Reader) {
	var body strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil || strings.TrimRight(line, "\r\n") == "." {
			break
		}
		body.WriteString(line)
	}
	s.mu.Lock()
	s.bodies = append(s.bodies, body.String())
	s.mu.Unlock()
}
