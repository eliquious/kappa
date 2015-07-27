package skl

import (
	"reflect"
	"strings"
	"testing"
)

// Ensure the parser can parse strings into Statement ASTs.
func TestParser_ParseStatement(t *testing.T) {

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
		{s: `USE acme.example.`, err: `found EOF, expected identifier at line 1, char 18`},
		{s: `USE acme.example. `, err: `found WS, expected identifier at line 1, char 18`},
	}

	for i, tt := range tests {
		if tt.skip {
			t.Logf("skipping test of '%s'", tt.s)
			continue
		}
		stmt, err := NewParser(strings.NewReader(tt.s)).ParseStatement()

		if !reflect.DeepEqual(tt.err, errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.stmt, stmt) {
			// t.Logf("\nexp=%s\ngot=%s\n", mustMarshalJSON(tt.stmt), mustMarshalJSON(stmt))
			t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
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
