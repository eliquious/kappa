package skl

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}

// ParserTestSuite executes all the parser tests
type ParserTestSuite struct {
	suite.Suite
}

func (suite *ParserTestSuite) SetupTest() {
}

func (suite *ParserTestSuite) TearDownTest() {
}

// Ensure the parser can parse strings into Statement ASTs.
func (suite *ParserTestSuite) TestStatements() {
	var tests = []struct {
		skip bool
		s    string
		stmt Statement
		err  string
	}{
		{
			s:    `USE acme.example`,
			stmt: &UseStatement{Name: "acme.example"},
		},

		// Errors
		{s: `a bad statement.`, err: `found a, expected USE at line 1, char 1`},
		{s: `USE `, err: `found EOF, expected namespace at line 1, char 6`},
		{s: `USE acme.example.`, err: `found EOF, expected identifier at line 1, char 18`},
		{s: `USE acme.example. `, err: `found WS, expected identifier at line 1, char 18`},
	}

	for i, tt := range tests {
		if tt.skip {
			suite.T().Logf("skipping test of '%s'", tt.s)
			continue
		}
		stmt, err := NewParser(strings.NewReader(tt.s)).ParseStatement()

		if !reflect.DeepEqual(tt.err, errstring(err)) {
			suite.T().Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.stmt, stmt) {
			suite.T().Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
		}
	}
}

// errstring converts an error to its string representation.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
