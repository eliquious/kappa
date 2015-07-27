package skl

import (
	"io"

	"github.com/eliquious/lexer"
)

// bufScanner represents a wrapper for scanner to add a buffer.
// It provides a fixed-length circular buffer that can be unread.
type bufScanner struct {
	s   *lexer.Scanner
	i   int // buffer index
	n   int // buffer size
	buf [3]struct {
		tok lexer.Token
		pos lexer.Pos
		lit string
	}
}

// newBufScanner returns a new buffered scanner for a reader.
func newBufScanner(r io.Reader) *bufScanner {
	return &bufScanner{s: lexer.NewScanner(r)}
}

// Scan reads the next token from the scanner.
func (s *bufScanner) Scan() (tok lexer.Token, pos lexer.Pos, lit string) {
	return s.scanFunc(s.s.Scan)
}

// ScanRegex reads a regex token from the scanner.
func (s *bufScanner) ScanRegex() (tok lexer.Token, pos lexer.Pos, lit string) {
	return s.scanFunc(s.s.ScanRegex)
}

// scanFunc uses the provided function to scan the next token.
func (s *bufScanner) scanFunc(scan func() (lexer.Token, lexer.Pos, string)) (tok lexer.Token, pos lexer.Pos, lit string) {
	// If we have unread tokens then read them off the buffer first.
	if s.n > 0 {
		s.n--
		return s.curr()
	}

	// Move buffer position forward and save the token.
	s.i = (s.i + 1) % len(s.buf)
	buf := &s.buf[s.i]
	buf.tok, buf.pos, buf.lit = scan()

	return s.curr()
}

// Unscan pushes the previously token back onto the buffer.
func (s *bufScanner) Unscan() { s.n++ }

// curr returns the last read token.
func (s *bufScanner) curr() (tok lexer.Token, pos lexer.Pos, lit string) {
	buf := &s.buf[(s.i-s.n+len(s.buf))%len(s.buf)]
	return buf.tok, buf.pos, buf.lit
}
