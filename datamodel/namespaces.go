package datamodel

import "github.com/mikespook/gorbac"

type Namespace interface {

    // Roles returns a list of roles for user permissions
    Permissions() *gorbac.Rbac

    // AddUser registers a user with the namespace
    AddUser(username string)

    // RemoveUser unregisters a user with the namespace
    RemoveUser(username string)
}

type NamespaceStore interface {

    // Get returns a Namespace by name
    Get(name string) (Namespace, error)

    // Create inserts a new namespace
    Create(name string) (Namespace, error)

    // Delete removes a namespace
    Delete(name string) error
}
