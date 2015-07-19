package datamodel

import "github.com/mikespook/gorbac"

// PublicKey wraps an ssh.PublicKey and simply provides methods for validation.
type PublicKey interface {

    // Fingerprint provides a string hash representing a PublicKey
    Fingerprint() string

    // Equals determines the equivalence of two PublicKeys
    Equals([]byte) bool
}

// PublicKeyRing provides an interface for interacting with a user's public keys
type PublicKeyRing interface {

    // AddPublicKey simply adds a public key to the user's key ring
    AddPublicKey(pemBytes []byte) error

    // RemovePublicKey will remove a public key from a user's key ring
    RemovePublicKey(fingerprint string) error

    // ListPublicKey returns all of a user's public keys
    ListPublicKeys() []PublicKey

    // Contains determines if a key exists in the ring. The provided bytes should be the output of ssh.PublicKey.Marshal.
    Contains(key []byte) bool
}

// User represents a database user
type User interface {

    // ValidatePassword determines the validity of a password.
    ValidatePassword(password string) bool

    // UpdatePassword updates a user's password. This password is only used to log into the web ui.
    UpdatePassword(password string) error

    // KeyRing returns a PublicKeyRing containing all of a user's public keys
    KeyRing() PublicKeyRing

    // Permissions returns an interface for determining a user's accessibility for the given namespace
    Permissions(namespace string) *gorbac.Rbac

    // Namespaces returns a list of namespaces for which the user has access
    Namespaces() []string
}

// UserStore stores all user information
type UserStore interface {

    // Get returns a User by username
    Get(username string) (User, error)

    // Create inserts a new user
    Create(username string) (User, error)

    // Delete removes a user account from a namespace
    Delete(username string) error
}
