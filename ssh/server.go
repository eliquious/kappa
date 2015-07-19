package ssh

import (
    "crypto/rand"
    "crypto/x509"
    "fmt"
    "net"
    "time"

    log "github.com/mgutz/logxi/v1"
    "github.com/spf13/viper"

    "github.com/subsilent/kappa/datamodel"
    "github.com/subsilent/zbase32"
    "golang.org/x/crypto/ssh"
)

const (
    SignInIdLength      = 4
    SignInSecretLength  = 16
    SessionIdLength     = 8
    SessionSecretLength = 40
    CsrfTokenLength     = 40
)

func NewSSHServer(logger log.Logger, sys datamodel.System, privateKey ssh.Signer, roots *x509.CertPool) (server SSHServer, err error) {

    // Get user store
    users := sys.Users()

    // Create server config
    sshConfig := &ssh.ServerConfig{
        NoClientAuth: false,
        PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (perm *ssh.Permissions, err error) {

            // Get user if exists, otherwise return error
            user, err := users.Get(conn.User())
            if err != nil {
                return
            }

            // Check keyring for public key
            if keyring := user.KeyRing(); !keyring.Contains(key.Marshal()) {
                err = fmt.Errorf("invalid public key")
                return
            }

            // Add pubkey and username to permissions
            perm = &ssh.Permissions{
                Extensions: map[string]string{
                    "pubkey":   string(key.Marshal()),
                    "username": conn.User(),
                },
            }
            return
        },
        AuthLogCallback: func(conn ssh.ConnMetadata, method string, err error) {
            logger.Info("Login attempt", "user", conn.User(), "method", method, "error", err)
        },
    }
    sshConfig.AddHostKey(privateKey)

    // Get ssh bind addr
    bind := viper.GetString("SSHListen")
    if bind == "" {
        err = fmt.Errorf("Empty SSH bind address")
        return
    }

    // Open SSH socket
    logger.Info("Starting SSH server", "addr", bind)
    sshAddr, err := net.ResolveTCPAddr("tcp", bind)
    if err != nil {
        err = fmt.Errorf("Invalid tcp address")
        return
    }

    // Create listener
    listener, err := net.ListenTCP("tcp", sshAddr)
    if err != nil {
        return
    }

    server.logger = logger
    server.sshConfig = sshConfig
    server.listener = listener
    return
}

type SSHServer struct {
    logger    log.Logger
    sshConfig *ssh.ServerConfig
    listener  *net.TCPListener
    done      chan bool
}

func (s *SSHServer) Run(logger log.Logger, closer chan<- bool) {
    logger.Info("Starting SSH server", "addr", viper.GetString("SSHListen"))
    s.done = make(chan bool)

    // Start server
    go func(l log.Logger, sock *net.TCPListener, config *ssh.ServerConfig, c <-chan bool, complete chan<- bool) {
        defer sock.Close()
        for {

            // Accepts will only block for 1s
            sock.SetDeadline(time.Now().Add(time.Second))

            select {

            // Stop server on channel recieve
            case <-c:
                l.Info("Stopping SSH server")
                complete <- true
                return
            default:

                // Accept new connection
                tcpConn, err := sock.Accept()
                if err != nil {
                    if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
                        // l.Debug("Connection timeout...")
                    } else {
                        l.Warn("Connection failed", "error", err)
                    }
                    continue
                }

                // Handle connection
                l.Debug("Successful SSH connection")
                go handleTCPConnection(l, tcpConn, config)
            }
        }
    }(logger, s.listener, s.sshConfig, s.done, closer)
}

func (s *SSHServer) Wait() {
    s.done <- true
}

func generateToken(length int) (token string, err error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }

    // stringify token
    token, err = zbase32.EncodeAll(bytes)
    return
}
