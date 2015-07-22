package datamodel

import (
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
    NS  UserStore
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
    suite.NS = NewBoltUserStore(ks)
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
    ns, err := suite.NS.Create("")
    suite.Nil(ns)
    suite.NotNil(err)
}

// TestCreateUser ensures a user can be created
func (suite *UserTestSuite) TestCreateUser() {
    ns, err := suite.NS.Create("acme")
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
    ns, err := suite.NS.Get("acme.none")
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
    ns, err := suite.NS.Create("acme.delete")
    suite.Nil(err)
    suite.NotNil(ns)

    // Test that the user was created
    suite.KS.ReadTx(func(bkt *bolt.Bucket) {

        b := bkt.Bucket([]byte("acme.delete"))
        suite.NotNil(b)
    })

    // Delete the user
    err = suite.NS.Delete("acme.delete")
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
    ns, err := suite.NS.Create(name)
    suite.Nil(err)
    suite.NotNil(ns)
    return ns, err
}

func (suite *UserTestSuite) deleteUser(name string) {
    err := suite.NS.Delete(name)
    suite.Nil(err)
}
