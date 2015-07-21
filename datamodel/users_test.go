package datamodel

import (
    "io/ioutil"
    "os"

    "testing"

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
}

// SetupSuite prepares the suite before any tests are ran
func (suite *UserTestSuite) SetupSuite() {

    // Create temp directory
    suite.Dir, _ = ioutil.TempDir("", "datamodel.test")

}

// TearDownSuite cleans up suite state after all the tests have completed
func (suite *UserTestSuite) TearDownSuite() {
    os.RemoveAll(suite.Dir)

}

// SetupTest prepares each test before execution
func (suite *UserTestSuite) SetupTest() {
}

// TearDownTest cleans up after each test
func (suite *UserTestSuite) TearDownTest() {
}
