package datamodel

import (
    "fmt"
    "strings"

    "github.com/boltdb/bolt"
    "github.com/eliquious/leaf"
)

var (

    // ErrNamespaceDoesNotExist is returned if a namespace does not exist when an operation is attempted to be performed on it
    ErrNamespaceDoesNotExist = fmt.Errorf("namespace does not exist")
)

// Namespace represents a namespace in the database. Each Namespace has users, logs and views.
type Namespace interface {

    // AddRole adds a new role to the namespace
    AddRole(name string) error

    // RemoveRole deletes a roel from the namespace
    RemoveRole(name string) error

    // Roles returns a list of roles for user permissions
    Roles() []string

    // GrantPermissions appends permissions for the given role
    GrantPermissions(role string, permissions ...string) error

    // RevokePermission removes a permission from the given role
    RevokePermission(role string, permission string) error

    // HasPermission detmines if the given role has a certain permission
    HasPermission(role string, permission string) bool

    // AddUser registers a user with the namespace
    AddUser(username string) error

    // RemoveUser unregisters a user with the namespace
    RemoveUser(username string) error

    // HasAccess determines if the namespace grants access to the given user
    HasAccess(username string) bool

    // Users returns a list of authorized users
    Users() []string
}

// NamespaceStore contains namespace information
type NamespaceStore interface {

    // Get returns a Namespace by name
    Get(name string) (Namespace, error)

    // Create inserts a new namespace
    Create(name string) (Namespace, error)

    // Delete removes a namespace
    Delete(name string) error
}

// NewBoltNamespaceStore creates a new NamespaceStore using the given keyspace
func NewBoltNamespaceStore(ks leaf.Keyspace) NamespaceStore {
    return &boltNamespaceStore{ks}
}

type boltNamespaceStore struct {
    ks leaf.Keyspace
}

// Create adds a namespace to the database
func (b boltNamespaceStore) Create(name string) (ns Namespace, err error) {
    b.ks.WriteTx(func(bkt *bolt.Bucket) {

        // Create bucket
        if _, err = bkt.CreateBucketIfNotExists([]byte(name)); err == nil {
            ns = boltNamespace{[]byte(name), b.ks}
        }
        return
    })
    return
}

// Get returns a Namespace, creating it if doesn't exist
func (b boltNamespaceStore) Get(name string) (Namespace, error) {
    return b.Create(name)
}

// Delete removes a namespace from the database
func (b boltNamespaceStore) Delete(name string) (err error) {
    b.ks.WriteTx(func(bkt *bolt.Bucket) {

        // Delete bucket
        err = bkt.DeleteBucket([]byte(name))
        return
    })
    return
}

// boltNamespace implements the Namespace interface on top of boltdb
//
// Each namespace has a bucket in the keyspace. Inside each bucket, there is a key for users and another bucket for roles. The user key contains a comma dlimited array of usernames. The interior roles bucket contains keys for each role and a comma delimited list of permissions.
type boltNamespace struct {
    name       []byte
    namespaces leaf.Keyspace
}

func (b boltNamespace) HasAccess(username string) (access bool) {
    b.namespaces.ReadTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            return
        }

        // Get existing users.
        u := ns.Get([]byte("users"))
        if len(u) > 0 {
            users := strings.Split(string(u), ",")

            for _, user := range users {
                if user == username {
                    access = true
                    break
                }
            }
        }
        return
    })
    return
}

// Users returns a list of users
func (b boltNamespace) Users() (users []string) {
    b.namespaces.ReadTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            return
        }

        // Get existing users.
        u := ns.Get([]byte("users"))
        if len(u) > 0 {
            users = strings.Split(string(u), ",")
        }

        return
    })
    return
}

// AddUser registers a user with the namespace
func (b boltNamespace) AddUser(username string) (err error) {
    b.namespaces.WriteTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            err = ErrNamespaceDoesNotExist
            return
        }

        // Get existing users. If users already exists, join the new user at the end, otherwise, simply add the new user.
        users := ns.Get([]byte("users"))
        if len(users) > 0 {

            list := []string{string(users), username}
            ns.Put([]byte("users"), []byte(strings.Join(list, ",")))
        } else {
            ns.Put([]byte("users"), []byte(username))
        }

        return
    })
    return
}

// RemoveUser unregisters a user with the namespace
func (b boltNamespace) RemoveUser(username string) (err error) {
    b.namespaces.WriteTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            err = ErrNamespaceDoesNotExist
            return
        }

        // Get user list
        users := ns.Get([]byte("users"))

        // Iterate over all the users and remove the given username
        if len(users) > 0 {
            var updated []string
            for _, user := range strings.Split(string(users), ",") {
                if user != username {
                    updated = append(updated, user)
                }
            }

            // Update the user list
            ns.Put([]byte("users"), []byte(strings.Join(updated, ",")))
        }

        return
    })
    return
}

// AddRole adds a new role to the namespace
func (b boltNamespace) AddRole(name string) (err error) {
    b.namespaces.WriteTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            err = ErrNamespaceDoesNotExist
            return
        }

        // Get roles bucket
        roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
        if err != nil {
            return
        }

        // If the role does not exist, create it.
        if roles.Get([]byte(name)) == nil {
            err = roles.Put([]byte(name), []byte(""))
        }
        return
    })
    return
}

// RemoveRole deletes a role from the namespace
func (b boltNamespace) RemoveRole(name string) (err error) {
    b.namespaces.WriteTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            err = ErrNamespaceDoesNotExist
            return
        }

        // Get roles bucket
        roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
        if err != nil {
            return
        }

        // If the role does exist, remove it.
        err = roles.Delete([]byte(name))
        return
    })
    return
}

// Roles returns a list of roles for user permissions
func (b boltNamespace) Roles() (list []string) {
    b.namespaces.WriteTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            return
        }

        // Get roles bucket
        roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
        if err != nil {
            return
        }

        // Get roles bucket and iterate over keys
        roles.ForEach(func(k []byte, _ []byte) error {
            list = append(list, string(k))
            return nil
        })
        return
    })
    return
}

// GrantPermissions appends permissions for the given role
func (b boltNamespace) GrantPermissions(role string, permissions ...string) (err error) {
    b.namespaces.WriteTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            err = ErrNamespaceDoesNotExist
            return
        }

        // Get roles bucket
        roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
        if err != nil {
            return
        }

        // Get existing permissions and add new
        perms := roles.Get([]byte(role))
        if len(perms) > 0 {

            list := []string{string(perms), strings.Join(permissions, ",")}
            roles.Put([]byte(role), []byte(strings.Join(list, ",")))
        } else {
            roles.Put([]byte(role), []byte(strings.Join(permissions, ",")))
        }
        return
    })
    return
}

// RevokePermissions removes a permission from the given role
func (b boltNamespace) RevokePermission(role string, permission string) (err error) {
    b.namespaces.WriteTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            err = ErrNamespaceDoesNotExist
            return
        }

        // Get roles bucket
        roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
        if err != nil {
            return
        }

        // Get existing permissions and add new
        perms := roles.Get([]byte(role))
        if len(perms) > 0 {

            var list []string
            for _, p := range strings.Split(string(perms), ",") {
                if p != permission {
                    list = append(list, p)
                }
            }

            // Save permissions
            err = roles.Put([]byte(role), []byte(strings.Join(list, ",")))
        }
        return
    })
    return
}

func (b boltNamespace) HasPermission(role string, permission string) (allow bool) {
    b.namespaces.WriteTx(func(bkt *bolt.Bucket) {

        // Get namespace bucket
        ns := bkt.Bucket(b.name)
        if ns == nil {
            return
        }

        // Get roles bucket
        roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
        if err != nil {
            return
        }

        // Get permissions
        perms := roles.Get([]byte(role))
        if len(perms) > 0 {
            for _, p := range strings.Split(string(perms), ",") {
                if p == permission {
                    allow = true
                    break
                }
            }
        }

        return
    })
    return
}
