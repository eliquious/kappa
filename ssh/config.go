package ssh

import (
	log "github.com/mgutz/logxi/v1"
	"github.com/subsilent/kappa/datamodel"
	"golang.org/x/crypto/ssh"
)

type Config struct {

	// Logger logs errors and debug output for the SSH server
	Logger log.Logger

	// Port specifies the port the SSH server will listen on
	Port int

	// PrivateKey is added to the SSH config as a host key
	PrivateKey ssh.Signer

	// System is the System datamodel
	System datamodel.System
}
