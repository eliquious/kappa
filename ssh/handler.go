package ssh

import (
    "net"

    log "github.com/mgutz/logxi/v1"
    "golang.org/x/crypto/ssh"
)

// AuthConnectionHandler validates connections against user accounts
type AuthConnectionHandler func(*ssh.ServerConn) bool

func handleTCPConnection(logger log.Logger, tcpConn net.Conn, sshConfig *ssh.ServerConfig, auth AuthConnectionHandler) {

    // Open SSH connection
    sshConn, channels, requests, err := ssh.NewServerConn(tcpConn, sshConfig)
    if err != nil {
        logger.Warn("SSH handshake failed")
        return
    } else if !auth(sshConn) {
        logger.Warn("Unauthenticated client. Closing connection.")
        return
    }
    logger.Debug("Handshake successful")
    defer sshConn.Conn.Close()

    // Discard requests
    go ssh.DiscardRequests(requests)

    for ch := range channels {
        t := ch.ChannelType()
        if t != "kappa-client" {
            ch.Reject(ssh.UnknownChannelType, t)
            continue
        }

        // Accept channel
        channel, requests, err := ch.Accept()
        if err != nil {
            logger.Warn("Invalid channel")
            continue
        }

        go handleChannelRequests(channel, requests)
    }
}

func handleChannelRequests(channel ssh.Channel, requests <-chan *ssh.Request) {
    defer channel.Close()

    for req := range requests {
        if req.Type == "query" {

            req.Reply(true, nil)
            // pubkey := []byte(sshConn.Permissions.Extensions["pubkey"])
            // url, err := handler(pubkey)

            // Failed to generate URL
            // if err != nil {
            //     logger.Warn("Failed to generate URL", "error", err)
            //     channel.Stderr().Write([]byte("Oh No! Something went wrong!"))
            // } else {

            //     // We're not loggin who logged in on purpose
            //     logger.Info("Successful login via SSH")
            //     fmt.Fprintln(channel, fmt.Sprintf("URL:\n%s\n", url))
            // }
            // break
        } else {
            if req.WantReply {
                req.Reply(false, nil)
            }
        }
    }
}
