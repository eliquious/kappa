package ssh

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
	tomb "gopkg.in/tomb.v2"
)

const (
	SignInIdLength      = 4
	SignInSecretLength  = 16
	SessionIdLength     = 8
	SessionSecretLength = 40
	CsrfTokenLength     = 40
)

func NewSSHServer(cfg *Config) (server SSHServer, err error) {

	// Create ssh config for server
	sshConfig, e := cfg.SSHConfig()
	if e != nil {
		err = e
		return
	}
	cfg.sshConfig = sshConfig

	// Validate the ssh bind addr
	if cfg.Bind == "" {
		err = fmt.Errorf("ssh server: Empty SSH bind address")
		return
	}

	// Open SSH socket
	sshAddr, e := net.ResolveTCPAddr("tcp", cfg.Bind)
	if e != nil {
		err = fmt.Errorf("ssh server: Invalid tcp address")
		return
	}

	// Create listener
	listener, e := net.ListenTCP("tcp", sshAddr)
	if e != nil {
		err = e
		return
	}
	server.listener = listener
	return
}

type SSHServer struct {
	config   *Config
	listener *net.TCPListener
	t        tomb.Tomb
}

func (s *SSHServer) Start() {
	s.config.Logger.Info("Starting SSH server", "addr", s.config.Bind)
	s.t.Go(s.listen)
}

func (s *SSHServer) Stop() error {
	s.t.Kill(nil)
	s.config.Logger.Info("Shutting down SSH server...")
	return s.t.Wait()
}

func (s *SSHServer) listen() error {
	defer s.listener.Close()

	// Create tomb for connection goroutines
	var t tomb.Tomb

	for {

		// Accepts will only block for 1s
		s.listener.SetDeadline(time.Now().Add(s.config.Deadline))

		select {

		// Stop server on channel receive
		case <-s.t.Dying():
			t.Kill(nil)
			t.Wait()
			return nil
		default:

			// Accept new connection
			tcpConn, err := s.listener.Accept()
			if err != nil {
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					s.config.Logger.Trace("Connection timeout...")
				} else {
					s.config.Logger.Warn("Connection failed", "error", err)
				}
				continue
			}

			// Handle connection
			s.config.Logger.Debug("Successful TCP connection: %s", tcpConn.RemoteAddr())
			t.Go(func() error {
				return s.handleTCPConnection(t, tcpConn)
			})
		}
	}
	return nil
}

func (s *SSHServer) handleTCPConnection(parentTomb tomb.Tomb, conn net.Conn) error {

	// Convert to SSH connection
	sshConn, channels, requests, err := ssh.NewServerConn(conn, s.config.sshConfig)
	if err != nil {
		s.config.Logger.Warn("SSH handshake failed: %s", conn.RemoteAddr())
		return err
	}

	// Close connection on exit
	s.config.Logger.Debug("Handshake successful")
	defer sshConn.Conn.Close()

	// Discard requests
	go ssh.DiscardRequests(requests)

	// Create new tomb stone
	var t tomb.Tomb

	for {
		select {
		case ch := <-channels:
			chType := ch.ChannelType()

			// Determine if channel is acceptable (has a registered handler)
			handler, ok := s.config.Handler(chType)
			if !ok {
				s.config.Logger.Info("UnknownChannelType", "type", chType)
				ch.Reject(ssh.UnknownChannelType, chType)
				break
			}

			// Accept channel
			channel, requests, err := ch.Accept()
			if err != nil {
				s.config.Logger.Warn("Error creating channel")
				continue
			}

			t.Go(func() error {
				return handler.Handle(t, sshConn, channel, requests)
			})
		case <-parentTomb.Dying():
			t.Kill(nil)
			if err := t.Wait(); err != nil {
				s.config.Logger.Warn("ssh handler error: %s", err)
			}
			sshConn.Close()
			return sshConn.Wait()
		}
	}
	return nil
}
