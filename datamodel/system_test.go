package datamodel

import (
	"io/ioutil"
	"os"
	"path"

	"testing"

	"github.com/subsilent/kappa/Godeps/_workspace/src/github.com/eliquious/leaf"
	"github.com/subsilent/kappa/Godeps/_workspace/src/github.com/stretchr/testify/suite"
)

// TestSystemTestSuite runs the SystemTestSuite
func TestSystemTestSuite(t *testing.T) {
	suite.Run(t, new(SystemTestSuite))
}

// SystemTestSuite tests all the System level database functionss
type SystemTestSuite struct {
	suite.Suite
	Dir    string
	System BoltSystemStore
	DB     leaf.KeyValueDatabase
}

// SetupSuite prepares the suite before any tests are ran
func (suite *SystemTestSuite) SetupSuite() {

	// Create temp directory
	suite.Dir, _ = ioutil.TempDir("", "datamodel.test")

	db, err := leaf.NewLeaf(path.Join(suite.Dir, "test.db"))
	if err != nil {
		suite.T().Log("Error creating database")
		suite.T().FailNow()
	}
	suite.DB = db
	suite.System = BoltSystemStore{db}
}

// TearDownSuite cleans up suite state after all the tests have completed
func (suite *SystemTestSuite) TearDownSuite() {

	// Close database
	suite.System.Close()

	// Clear test directory
	os.RemoveAll(suite.Dir)
}

// SetupTest prepares each test before execution
func (suite *SystemTestSuite) SetupTest() {
}

// TearDownTest cleans up after each test
func (suite *SystemTestSuite) TearDownTest() {
}

func (suite *SystemTestSuite) TestCreateKeyspaceError() {
	nss, err := suite.System.Namespaces()
	suite.Nil(err)
	suite.NotNil(nss)
}

func (suite *SystemTestSuite) TestGetUserStore() {
	nss, err := suite.System.Users()
	suite.Nil(err)
	suite.NotNil(nss)
}
