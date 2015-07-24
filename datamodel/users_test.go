package datamodel

import (
    "crypto/sha256"
    "io/ioutil"
    "os"
    "path"

    "testing"

    "github.com/boltdb/bolt"
    "github.com/eliquious/leaf"
    "github.com/stretchr/testify/suite"
)

// TestUserTestSuite runs the UserTestSuite
func TestUserTestSuite(t *testing.T) {
    suite.Run(t, new(UserTestSuite))
}

// UserTestSuite tests all the auth routes
type UserTestSuite struct {
    suite.Suite
    Dir string
    DB  leaf.KeyValueDatabase
    US  UserStore
    KS  leaf.Keyspace
}

// SetupSuite prepares the suite before any tests are ran
func (suite *UserTestSuite) SetupSuite() {

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
    ks, err := db.GetOrCreateKeyspace(Users)
    suite.Nil(err)
    suite.KS = ks

    // Create user store
    suite.US = NewBoltUserStore(ks)
}

// TearDownSuite cleans up suite state after all the tests have completed
func (suite *UserTestSuite) TearDownSuite() {
    os.RemoveAll(suite.Dir)

    // Close database
    suite.DB.Close()
}

// SetupTest prepares each test before execution
func (suite *UserTestSuite) SetupTest() {
}

// TearDownTest cleans up after each test
func (suite *UserTestSuite) TearDownTest() {
}

// TestCreateUserError ensures a bolt db error bubbles up. Such as an empty user name;
func (suite *UserTestSuite) TestCreateUserError() {
    ns, err := suite.US.Create("")
    suite.Nil(ns)
    suite.NotNil(err)
}

// TestCreateUser ensures a user can be created
func (suite *UserTestSuite) TestCreateUser() {
    ns, err := suite.US.Create("acme")
    suite.Nil(err)
    suite.NotNil(ns)

    // Test that the user was created
    suite.KS.ReadTx(func(bkt *bolt.Bucket) {

        b := bkt.Bucket([]byte("acme"))
        suite.NotNil(b)
    })
}

// TestGetUser ensures a user can be created
func (suite *UserTestSuite) TestGetUser() {

    // Test that the user does not exist prior
    suite.KS.ReadTx(func(bkt *bolt.Bucket) {

        b := bkt.Bucket([]byte("acme.none"))
        suite.Nil(b)
    })

    // Get the user
    ns, err := suite.US.Get("acme.none")
    suite.Nil(err)
    suite.NotNil(ns)

    // Test that the user was created
    suite.KS.ReadTx(func(bkt *bolt.Bucket) {

        b := bkt.Bucket([]byte("acme.none"))
        suite.NotNil(b)
    })
}

// TestDeleteUser ensures a user can be deleted
func (suite *UserTestSuite) TestDeleteUser() {
    ns, err := suite.US.Create("acme.delete")
    suite.Nil(err)
    suite.NotNil(ns)

    // Test that the user was created
    suite.KS.ReadTx(func(bkt *bolt.Bucket) {

        b := bkt.Bucket([]byte("acme.delete"))
        suite.NotNil(b)
    })

    // Delete the user
    err = suite.US.Delete("acme.delete")
    suite.Nil(err)

    // Test that the user was deleted
    suite.KS.ReadTx(func(bkt *bolt.Bucket) {

        b := bkt.Bucket([]byte("acme.delete"))
        suite.Nil(b)
    })
}

func (suite *UserTestSuite) verifyUserExists(name string) (exists bool) {

    // Test that the user was created
    suite.KS.ReadTx(func(bkt *bolt.Bucket) {
        b := bkt.Bucket([]byte(name))
        exists = b != nil
    })
    return
}

func (suite *UserTestSuite) createUser(name string) (User, error) {
    ns, err := suite.US.Create(name)
    suite.Nil(err)
    suite.NotNil(ns)
    return ns, err
}

func (suite *UserTestSuite) deleteUser(name string) {
    err := suite.US.Delete(name)
    suite.Nil(err)
}

// TestValidatePasswordNoUser
func (suite *UserTestSuite) TestValidatePasswordNoUser() {
    // user, err := suite.US.Create("acme.validate.password")
    // suite.Nil(err)
    // suite.NotNil(us)

    user := boltUser{[]byte("blahblahblah"), suite.KS}

    match := user.ValidatePassword("password")
    suite.False(match)
}

// TestValidatePasswordNoSalt
func (suite *UserTestSuite) TestValidatePasswordNoSalt() {
    user, err := suite.US.Create("acme.validate.password.no.salt")
    suite.Nil(err)
    suite.NotNil(user)

    match := user.ValidatePassword("password")
    suite.False(match)
}

// TestValidatePasswordNoSaltedPassword
func (suite *UserTestSuite) TestValidatePasswordNoSaltedPassword() {
    name := "acme.validate.password.no.saltedpw"
    user, err := suite.US.Create(name)
    suite.Nil(err)
    suite.NotNil(user)

    // Write salt
    suite.KS.WriteTx(func(bkt *bolt.Bucket) {
        userBucket := bkt.Bucket([]byte(name))

        // Generate salt
        salt, _, _ := GenerateSalt([]byte("password"))
        userBucket.Put([]byte("salt"), salt)
    })

    match := user.ValidatePassword("password")
    suite.False(match)
}

// TestValidatePasswordInvalid
func (suite *UserTestSuite) TestValidatePasswordInvalid() {
    name := "acme.validate.password.invalid"
    user, err := suite.US.Create(name)
    suite.Nil(err)
    suite.NotNil(user)

    // Write salt
    suite.KS.WriteTx(func(bkt *bolt.Bucket) {
        userBucket := bkt.Bucket([]byte(name))

        // Generate salt
        salt, saltedpw, _ := GenerateSalt([]byte("password"))
        userBucket.Put([]byte("salt"), salt)
        userBucket.Put([]byte("salted_password"), saltedpw)
    })

    // Test match
    match := user.ValidatePassword("shaken, not stirred")
    suite.False(match)
}

// TestValidatePassword
func (suite *UserTestSuite) TestValidatePassword() {
    name := "acme.validate.password"
    user, err := suite.US.Create(name)
    suite.Nil(err)
    suite.NotNil(user)

    // Write salt
    suite.KS.WriteTx(func(bkt *bolt.Bucket) {
        userBucket := bkt.Bucket([]byte(name))

        // Generate salt
        salt, saltedpw, _ := GenerateSalt([]byte("password"))
        userBucket.Put([]byte("salt"), salt)
        userBucket.Put([]byte("salted_password"), saltedpw)
    })

    // Test match
    match := user.ValidatePassword("password")
    suite.True(match)
}

func (suite *UserTestSuite) TestUpdatePasswordInvalidUser() {
    user := boltUser{[]byte("blahblahblah"), suite.KS}

    err := user.UpdatePassword("password")
    suite.NotNil(err)
}

func (suite *UserTestSuite) TestUpdatePassword() {
    name := "acme.update.password"

    // Create user
    user, err := suite.US.Create(name)
    suite.Nil(err)
    suite.NotNil(user)

    // Update password
    err = user.UpdatePassword("password")
    suite.Nil(err)

    // Validate salt + salted_password
    suite.KS.WriteTx(func(bkt *bolt.Bucket) {
        userBucket := bkt.Bucket([]byte(name))

        salt := userBucket.Get([]byte("salt"))
        suite.NotNil(salt)

        salted_password := userBucket.Get([]byte("salted_password"))
        suite.NotNil(salted_password)

        // Salt password
        hash := sha256.New()
        hash.Write(salt)
        hash.Write([]byte("password"))

        match := SecureCompare(hash.Sum(nil), salted_password)
        suite.True(match)
    })
}

func (suite *UserTestSuite) TestUserNamespacesInvalidUser() {
    user := boltUser{[]byte("blahblahblah"), suite.KS}

    // Get namespaces
    nss := user.Namespaces()
    suite.Nil(nss)
}

func (suite *UserTestSuite) TestUserNamespaces() {
    name := "acme.user.namespaces"

    // Create user
    user, err := suite.US.Create(name)
    suite.Nil(err)
    suite.NotNil(user)

    // Validate salt + salted_password
    suite.KS.WriteTx(func(bkt *bolt.Bucket) {
        userBucket := bkt.Bucket([]byte(name))

        ns, err := userBucket.CreateBucketIfNotExists([]byte("namespaces"))
        suite.Nil(err)

        ns.Put([]byte("acme.users"), []byte("guest"))
        ns.Put([]byte("acme.trending"), []byte("guest"))
    })

    // Get namespaces
    nss := user.Namespaces()
    suite.NotNil(nss)
    suite.Equal(len(nss), 2)
    suite.Equal(nss[0], "acme.trending")
    suite.Equal(nss[1], "acme.users")
}

func (suite *UserTestSuite) TestUserRolesInvalidUser() {
    user := boltUser{[]byte("blahblahblah"), suite.KS}

    // Get roles
    nss := user.Roles("acme.namespace")
    suite.Nil(nss)
}

func (suite *UserTestSuite) TestUserRoles() {
    name := "acme.user.roles"

    // Create user
    user, err := suite.US.Create(name)
    suite.Nil(err)
    suite.NotNil(user)

    // Validate salt + salted_password
    suite.KS.WriteTx(func(bkt *bolt.Bucket) {
        userBucket := bkt.Bucket([]byte(name))

        ns, err := userBucket.CreateBucketIfNotExists([]byte("namespaces"))
        suite.Nil(err)

        ns.Put([]byte("acme.users"), []byte("guest,admin"))
        ns.Put([]byte("acme.trending"), []byte("admin"))
    })

    // Get Roles for invalid namespace
    roles := user.Roles("invalid.namespace")
    suite.Equal(len(roles), 0)
    suite.Nil(roles)

    // Get roles for acme.users (multiple roles)
    roles = user.Roles("acme.users")
    suite.Equal(len(roles), 2)
    suite.NotNil(roles)
    suite.Equal("guest", roles[0])
    suite.Equal("admin", roles[1])

    // Get roles for acme.trending (single role)
    roles = user.Roles("acme.trending")
    suite.Equal(len(roles), 1)
    suite.NotNil(roles)
    suite.Equal(roles[0], "admin")
}
