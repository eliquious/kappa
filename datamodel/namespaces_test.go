package datamodel

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"testing"

	"github.com/subsilent/kappa/Godeps/_workspace/src/github.com/boltdb/bolt"
	"github.com/subsilent/kappa/Godeps/_workspace/src/github.com/eliquious/leaf"
	"github.com/subsilent/kappa/Godeps/_workspace/src/github.com/stretchr/testify/suite"
)

// TestNamespaceTestSuite runs the NamespaceTestSuite
func TestNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceTestSuite))
}

// NamespaceTestSuite tests all the auth routes
type NamespaceTestSuite struct {
	suite.Suite
	Dir string
	DB  leaf.KeyValueDatabase
	NS  NamespaceStore
	KS  leaf.Keyspace
}

// SetupSuite prepares the suite before any tests are ran
func (suite *NamespaceTestSuite) SetupSuite() {

	// Create temp directory
	suite.Dir, _ = ioutil.TempDir("", "datamodel.test")

	// Connect to database
	db, err := leaf.NewLeaf(path.Join(suite.Dir, "test.db"))
	if err != nil {
		suite.T().Log("Error creating database")
		suite.T().FailNow()
	}
	suite.DB = db

	// Create keyspace
	ks, err := db.GetOrCreateKeyspace(Namespaces)
	suite.Nil(err)
	suite.KS = ks

	// Create namespace store
	suite.NS = NewBoltNamespaceStore(ks)
}

// TearDownSuite cleans up suite state after all the tests have completed
func (suite *NamespaceTestSuite) TearDownSuite() {

	// Close database
	suite.DB.Close()

	// Clear test directory
	os.RemoveAll(suite.Dir)
}

// SetupTest prepares each test before execution
func (suite *NamespaceTestSuite) SetupTest() {
}

// TearDownTest cleans up after each test
func (suite *NamespaceTestSuite) TearDownTest() {
}

// TestCreateKeyspaceError ensures a bolt db error bubbles up. Such as an empty namespace name;
func (suite *NamespaceTestSuite) TestCreateKeyspaceError() {
	ns, err := suite.NS.Create("")
	suite.Nil(ns)
	suite.NotNil(err)
}

// TestCreateKeyspace ensures a namespace can be created
func (suite *NamespaceTestSuite) TestCreateKeyspace() {
	ns, err := suite.NS.Create("acme")
	suite.Nil(err)
	suite.NotNil(ns)

	// Test that the namespace was created
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		b := bkt.Bucket([]byte("acme"))
		suite.NotNil(b)
	})
}

// TestGetKeyspace ensures a namespace can be created
func (suite *NamespaceTestSuite) TestGetKeyspace() {

	// Test that the namespace does not exist prior
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		b := bkt.Bucket([]byte("acme.none"))
		suite.Nil(b)
	})

	// Get the namespace
	ns, err := suite.NS.Get("acme.none")
	suite.Nil(err)
	suite.NotNil(ns)

	// Test that the namespace was created
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		b := bkt.Bucket([]byte("acme.none"))
		suite.NotNil(b)
	})
}

// TestDeleteKeyspace ensures a namespace can be deleted
func (suite *NamespaceTestSuite) TestDeleteKeyspace() {
	ns, err := suite.NS.Create("acme.delete")
	suite.Nil(err)
	suite.NotNil(ns)

	// Test that the namespace was created
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		b := bkt.Bucket([]byte("acme.delete"))
		suite.NotNil(b)
	})

	// Delete the namespace
	err = suite.NS.Delete("acme.delete")
	suite.Nil(err)

	// Test that the namespace was deleted
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		b := bkt.Bucket([]byte("acme.delete"))
		suite.Nil(b)
	})
}

func (suite *NamespaceTestSuite) verifyNamespaceExists(name string) (exists bool) {

	// Test that the namespace was created
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {
		b := bkt.Bucket([]byte(name))
		exists = b != nil
	})
	return
}

func (suite *NamespaceTestSuite) createNamespace(name string) (Namespace, error) {
	ns, err := suite.NS.Create(name)
	suite.Nil(err)
	suite.NotNil(ns)
	return ns, err
}

func (suite *NamespaceTestSuite) deleteNamespace(name string) {
	err := suite.NS.Delete(name)
	suite.Nil(err)
}

func (suite *NamespaceTestSuite) TestHasAccessEmpty() {
	name := "acme.has.access"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// User should not have access as no users have been assigned to the namespace
	access := ns.HasAccess("user")
	suite.Equal(false, access)
}

func (suite *NamespaceTestSuite) TestHasAccessMultiple() {
	name := "acme.has.access.multiple"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	usernames := []string{"bugs.bunny", "sylvester"}
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))
		if ns == nil {
			return
		}

		// Add test users
		ns.Put([]byte("users"), []byte(strings.Join(usernames, ",")))
		return
	})

	// User should not have access as no users have been assigned to the namespace
	access := ns.HasAccess("bugs.bunny")
	suite.Equal(true, access)

	access = ns.HasAccess("sylvester")
	suite.Equal(true, access)

	access = ns.HasAccess("elmyra")
	suite.Equal(false, access)
}

func (suite *NamespaceTestSuite) TestHasAccessInvalidNamespace() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// User should not have access as the namespace does not exist
	access := ns.HasAccess("user")
	suite.Equal(false, access)
}

func (suite *NamespaceTestSuite) TestUsersEmpty() {
	name := "acme.users"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// User list should be empty
	users := ns.Users()
	suite.Equal(0, len(users))
}

func (suite *NamespaceTestSuite) TestUsersInvalidNamespace() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// User should not have access as the namespace does not exist
	users := ns.Users()
	suite.Equal(0, len(users))
}

func (suite *NamespaceTestSuite) TestUsersMultiple() {
	name := "acme.users"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	usernames := []string{"bugs.bunny", "sylvester"}
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))
		if ns == nil {
			return
		}

		// Add test users
		ns.Put([]byte("users"), []byte(strings.Join(usernames, ",")))
		return
	})

	// User list should contain bugs.bunny and sylvester
	users := ns.Users()
	suite.Equal(2, len(users))
	suite.Equal("bugs.bunny", users[0])
	suite.Equal("sylvester", users[1])
}

func (suite *NamespaceTestSuite) TestAddUserInvalidNamespace() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// User should not be added as the namespace does not exist
	err := ns.AddUser("wiley.coyote")
	suite.NotNil(err)
}

func (suite *NamespaceTestSuite) TestAddUserMultiple() {
	name := "acme.add.user"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	err := ns.AddUser("bugs.bunny")
	suite.Nil(err)

	err = ns.AddUser("sylvester")
	suite.Nil(err)

	usernames := []string{"bugs.bunny", "sylvester"}
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Add test users
		users := ns.Get([]byte("users"))
		suite.Equal([]byte(strings.Join(usernames, ",")), users)
		return
	})
}

func (suite *NamespaceTestSuite) TestAddUserEmpty() {
	name := "acme.add.user.multiple"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	err := ns.AddUser("bugs.bunny")
	suite.Nil(err)

	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Add test users
		users := ns.Get([]byte("users"))
		suite.Equal([]byte("bugs.bunny"), users)
		return
	})
}

func (suite *NamespaceTestSuite) TestRemoveUserInvalid() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// User cannot be removed as the namespace does not exist
	err := ns.RemoveUser("wiley.coyote")
	suite.NotNil(err)
}

func (suite *NamespaceTestSuite) TestRemoveUserMultiple() {
	name := "acme.add.user.multiple"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Add test users
	usernames := []string{"bugs.bunny", "sylvester"}
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Add test users
		ns.Put([]byte("users"), []byte(strings.Join(usernames, ",")))
		return
	})

	// Remove user
	err := ns.RemoveUser("sylvester")
	suite.Nil(err)

	// Verify sylvester was removed
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Add test users
		users := ns.Get([]byte("users"))
		suite.Equal([]byte("bugs.bunny"), users)
		return
	})
}

func (suite *NamespaceTestSuite) TestRemoveUserSingle() {
	name := "acme.add.user.multiple"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Add test users
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Add test users
		ns.Put([]byte("users"), []byte("bugs.bunny"))
		return
	})

	// Remove user
	err := ns.RemoveUser("bugs.bunny")
	suite.Nil(err)

	// Verify sylvester was removed
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Add test users
		users := ns.Get([]byte("users"))
		suite.Equal([]byte{}, users)
		suite.Equal(0, len(users))
		return
	})
}

func (suite *NamespaceTestSuite) TestAddRoleInvalidNamespace() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// Role cannot be added as the namespace does not exist
	err := ns.AddRole("guest")
	suite.NotNil(err)
}

func (suite *NamespaceTestSuite) TestAddRoleExists() {
	name := "acme.add.old.role"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Add test users
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
		suite.Nil(err)

		// Role should not exist
		err = roles.Put([]byte("guest"), []byte("read"))
		suite.Nil(err)
		return
	})

	// Add roles
	err := ns.AddRole("guest")
	suite.Nil(err)

	// Verify guest role was added
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
		suite.Nil(err)

		// Role should not exist
		suite.Equal([]byte("read"), roles.Get([]byte("guest")))
		return
	})
}

func (suite *NamespaceTestSuite) TestAddRoleDoesNotExist() {
	name := "acme.add.new.role"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Add test users
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
		suite.Nil(err)

		// Role should not exist
		suite.Nil(roles.Get([]byte("guest")))
		return
	})

	// Add roles
	err := ns.AddRole("guest")
	suite.Nil(err)

	// Verify guest role was added
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
		suite.Nil(err)

		// Role should not exist
		suite.NotNil(roles.Get([]byte("guest")))
		return
	})
}

func (suite *NamespaceTestSuite) TestRemoveRoleInvalidNamespace() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// Role cannot be removed as the namespace does not exist
	err := ns.RemoveRole("guest")
	suite.NotNil(err)
}

func (suite *NamespaceTestSuite) TestRemoveRoleDoesNotExist() {
	name := "acme.remove.new.role"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Remove role
	err := ns.RemoveRole("guest")
	suite.Nil(err)

	// Verify remove role does not exist
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
		suite.Nil(err)

		// Role should not exist
		suite.Nil(roles.Get([]byte("guest")))
		return
	})
}

func (suite *NamespaceTestSuite) TestRemoveRole() {
	name := "acme.remove.old.role"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Add role
	err := ns.AddRole("guest")
	suite.Nil(err)

	// Remove role
	err = ns.RemoveRole("guest")
	suite.Nil(err)

	// Verify remove role does not exist
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles, err := ns.CreateBucketIfNotExists([]byte("roles"))
		suite.Nil(err)

		// Role should not exist
		suite.Nil(roles.Get([]byte("guest")))
		return
	})
}

func (suite *NamespaceTestSuite) TestRolesEmpty() {
	name := "acme.roles.empty"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Get roles
	roles := ns.Roles()
	suite.Equal(0, len(roles))
}

func (suite *NamespaceTestSuite) TestRoles() {
	name := "acme.roles"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Add role
	err := ns.AddRole("guest")
	suite.Nil(err)

	// Add role
	err = ns.AddRole("admin")
	suite.Nil(err)

	// Get roles
	roles := ns.Roles()
	suite.Equal(2, len(roles))
	suite.Equal("admin", roles[0])
	suite.Equal("guest", roles[1])
}

func (suite *NamespaceTestSuite) TestRolesInvalidNamespace() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// Role cannot be removed as the namespace does not exist
	roles := ns.Roles()
	suite.Equal(0, len(roles))
}

func (suite *NamespaceTestSuite) TestGrantPermissionsInvalidNamespace() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// Permissions cannot be added because namespace does not exist
	err := ns.GrantPermissions("admin", "users.list", "users.add")
	suite.NotNil(err)
}

func (suite *NamespaceTestSuite) TestGrantPermissionsRoleDoesNotExist() {
	name := "acme.grant.permissions.new.role"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Grant permissions
	err := ns.GrantPermissions("guest", "subscribe", "select")
	suite.Nil(err)

	// Verify permissions were granted
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles := ns.Bucket([]byte("roles"))
		suite.NotNil(roles)

		// Role should exist with permissions
		permissions := strings.Split(string(roles.Get([]byte("guest"))), ",")
		suite.Equal(2, len(permissions))
		suite.Equal("subscribe", permissions[0])
		suite.Equal("select", permissions[1])
		return
	})
}

func (suite *NamespaceTestSuite) TestGrantPermissions() {
	name := "acme.grant.permissions"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Grant permissions
	err := ns.GrantPermissions("guest", "subscribe")
	suite.Nil(err)

	// Grant permissions
	err = ns.GrantPermissions("guest", "select")
	suite.Nil(err)

	// Verify permissions were granted
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles := ns.Bucket([]byte("roles"))
		suite.NotNil(roles)

		// Role should exist with permissions
		permissions := strings.Split(string(roles.Get([]byte("guest"))), ",")
		suite.Equal(2, len(permissions))
		suite.Equal("subscribe", permissions[0])
		suite.Equal("select", permissions[1])
		return
	})
}

// func (suite *NamespaceTestSuite) TestRevokePermission() {
// }

func (suite *NamespaceTestSuite) TestRevokePermissionsInvalidNamespace() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// Permissions cannot be removed because namespace does not exist
	err := ns.RevokePermission("admin", "users.list")
	suite.NotNil(err)
}

func (suite *NamespaceTestSuite) TestRevokePermissions() {
	name := "acme.revoke.permissions.exists"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Grant permissions
	err := ns.GrantPermissions("guest", "subscribe", "select")
	suite.Nil(err)

	err = ns.RevokePermission("guest", "select")
	suite.Nil(err)

	// Verify permissions were granted
	suite.KS.WriteTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles := ns.Bucket([]byte("roles"))
		suite.NotNil(roles)

		// Role should exist with permissions
		permissions := strings.Split(string(roles.Get([]byte("guest"))), ",")
		suite.Equal(1, len(permissions))
		suite.Equal("subscribe", permissions[0])
		return
	})
}

func (suite *NamespaceTestSuite) TestRevokePermissionsRoleDoesNotExist() {
	name := "acme.revoke.permissions"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Revoke permissions
	err := ns.RevokePermission("guest", "select")
	suite.Nil(err)

	// Verify permissions were granted
	suite.KS.ReadTx(func(bkt *bolt.Bucket) {

		// Get namespace bucket
		ns := bkt.Bucket([]byte(name))

		// Get roles bucket
		roles := ns.Bucket([]byte("roles"))
		suite.NotNil(roles)

		// Role should exist with permissions
		permissions := roles.Get([]byte("guest"))
		suite.Equal(0, len(permissions))
		return
	})
}

func (suite *NamespaceTestSuite) TestHasPermissionInvalidNamespace() {
	name := "acme.fake"

	// Create namespace
	ns := boltNamespace{[]byte(name), suite.KS}

	// Role should not have permissions as namespace does not exist
	allow := ns.HasPermission("admin", "users.list")
	suite.False(allow)
}

func (suite *NamespaceTestSuite) TestHasPermissions() {
	name := "acme.has.permissions"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Grant permissions
	err := ns.GrantPermissions("guest", "subscribe", "select")
	suite.Nil(err)

	// Test permissions
	allow := ns.HasPermission("guest", "create.view")
	suite.False(allow)

	allow = ns.HasPermission("guest", "subscribe")
	suite.True(allow)

	allow = ns.HasPermission("guest", "select")
	suite.True(allow)
}

func (suite *NamespaceTestSuite) TestHasPermissionsRoleDoesNotExist() {
	name := "acme.revoke.permissions"

	// Create namespace
	ns, _ := suite.createNamespace(name)

	// Test that the namespace was created
	suite.verifyNamespaceExists(name)

	// Test permissions
	allow := ns.HasPermission("guest", "select")
	suite.False(allow)
}
