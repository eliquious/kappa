package datamodel

import "github.com/subsilent/kappa/Godeps/_workspace/src/github.com/eliquious/leaf"

const (

	// Users is the name of the user keyspace
	Users = "users"

	// Namespaces is the name of the namespace keyspace
	Namespaces = "namespaces"
)

// System provides an interface for accessing information about the database.
type System interface {
	Users() (UserStore, error)
	Namespaces() (NamespaceStore, error)

	Close()
}

// NewSystem creates a database connection to access system metadata
func NewSystem(filename string) (System, error) {
	leaf, err := leaf.NewLeaf(filename)
	if err != nil {
		return nil, err
	}
	return &BoltSystemStore{leaf}, nil
}

// BoltSystemStore implements the System interface on top of a boltdb connection
type BoltSystemStore struct {
	db leaf.KeyValueDatabase
}

// Users returns a UserStore
func (s BoltSystemStore) Users() (UserStore, error) {
	ks, err := s.db.GetOrCreateKeyspace(Users)
	if err != nil {
		return nil, err
	}
	return NewBoltUserStore(ks), nil
}

// Namespaces returns a NamespaceStore
func (s BoltSystemStore) Namespaces() (NamespaceStore, error) {
	ks, err := s.db.GetOrCreateKeyspace(Namespaces)
	if err != nil {
		return nil, err
	}
	return NewBoltNamespaceStore(ks), nil
}

// Close closes the database connection
func (s BoltSystemStore) Close() {
	s.db.Close()
}
