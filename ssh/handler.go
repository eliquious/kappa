package ssh

import (
	"fmt"
	"net"
	"strings"

	log "github.com/mgutz/logxi/v1"
	"github.com/subsilent/kappa/skl"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// AuthConnectionHandler validates connections against user accounts
type AuthConnectionHandler func(*ssh.ServerConn) bool

func handleTCPConnection(logger log.Logger, conn net.Conn, sshConfig *ssh.ServerConfig) {

	// Open SSH connection
	sshConn, channels, requests, err := ssh.NewServerConn(conn, sshConfig)
	if err != nil {
		logger.Warn("SSH handshake failed")
		return
	}

	logger.Debug("Handshake successful")
	defer sshConn.Conn.Close()

	// Discard requests
	go ssh.DiscardRequests(requests)

	for ch := range channels {
		t := ch.ChannelType()

		if t != "session" && t != "kappa-client" {
			logger.Info("UnknownChannelType", "type", t)
			ch.Reject(ssh.UnknownChannelType, t)
			break
		}

		// Accept channel
		channel, requests, err := ch.Accept()
		if err != nil {
			logger.Warn("Error creating channel")
			continue
		}

		if t == "session" {
			go handleSessionRequests(logger, channel, requests)
		} else if t == "kappa-client" {
			go handleChannelRequests(logger, channel, requests)
		}
	}
}

func handleChannelRequests(logger log.Logger, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		if req.Type == "skl" {
			logger.Info("SKL request", "request", string(req.Payload))
			req.Reply(true, nil)
		} else {
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

func handleSessionRequests(logger log.Logger, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	// Sessions have out-of-band requests such as "shell",
	// "pty-req" and "env".  Here we handle only the
	// "shell" request.
	for req := range requests {

		ok := false
		switch req.Type {
		case "shell":
			ok = true

			if len(req.Payload) > 0 {
				fmt.Println(string(req.Payload))
				// We don't accept any
				// commands, only the
				// default shell.
				ok = false
			}

		case "pty-req":
			// Responding 'ok' here will let the client
			// know we have a pty ready for input
			ok = true

			go startTerminal(logger, channel)
		default:
			// fmt.Println("default req: ", req)
		}

		req.Reply(ok, nil)
	}
}

func startTerminal(logger log.Logger, channel ssh.Channel) {
	defer channel.Close()
	term := terminal.NewTerminal(channel, "kappa > ")

	// // Try to make the terminal raw
	// oldState, err := terminal.MakeRaw(0)
	// if err != nil {
	//     logger.Warn("Error making terminal raw: ", err.Error())
	// }
	// defer terminal.Restore(0, oldState)

	for _, line := range ASCII {
		term.Write([]byte(line))
		term.Write([]byte("\r\n"))
	}
	term.Write([]byte("\r\nWelcome to Kappa DB!\r\n"))

	for {
		input, err := term.ReadLine()
		if err != nil {
			fmt.Errorf("Readline() error")
			break
		}

		// Process line
		line := strings.TrimSpace(input)
		if len(line) > 0 {

			// Log input and handle exit requests
			if line == "exit" || line == "quit" {
				logger.Info("Closing connection")
				break
			}

			// Parse statement
			stmt, err := skl.ParseStatement(line)

			// Return parse error in red
			if err != nil {
				logger.Warn("Bad Statement", "statement", line, "error", err)
				channel.Write(term.Escape.Red)
				channel.Write([]byte(err.Error()))
				channel.Write([]byte("\r\n"))
				channel.Write(term.Escape.Reset)
				continue
			}

			// For now, just echo the successful command in green
			logger.Info("Successful Statement", "statement", line)
			channel.Write(term.Escape.Green)
			channel.Write([]byte(stmt.String()))
			channel.Write([]byte("\r\n"))
			channel.Write(term.Escape.Reset)
		}
	}
}
