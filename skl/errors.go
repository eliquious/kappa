package skl

import (
	"fmt"
	"strings"

	"github.com/eliquious/lexer"
)

// ParseError represents an error that occurred during parsing.
type ParseError struct {
	Message  string
	Found    string
	Expected []string
	Pos      lexer.Pos
}

// newParseError returns a new instance of ParseError.
func newParseError(found string, expected []string, pos lexer.Pos) *ParseError {
	return &ParseError{Found: found, Expected: expected, Pos: pos}
}

// Error returns the string representation of the error.
func (e *ParseError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s at line %d, char %d", e.Message, e.Pos.Line+1, e.Pos.Char+1)
	}
	return fmt.Sprintf("found %s, expected %s at line %d, char %d", e.Found, strings.Join(e.Expected, ", "), e.Pos.Line+1, e.Pos.Char+1)
}
