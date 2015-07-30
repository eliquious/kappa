package skl

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/eliquious/lexer"
)

const (
	// DateFormat represents the format for date literals.
	DateFormat = "2006-01-02"

	// DateTimeFormat represents the format for date time literals.
	DateTimeFormat = "2006-01-02 15:04:05.999999"
)

// Parser represents an InfluxQL parser.
type Parser struct {
	s *bufScanner
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: newBufScanner(r)}
}

// ParseStatement parses a statement string and returns its AST representation.
func ParseStatement(s string) (Statement, error) {
	return NewParser(strings.NewReader(s)).ParseStatement()
}

// ParseStatement parses a string and returns a Statement AST object.
func (p *Parser) ParseStatement() (Statement, error) {

	// Inspect the first token.
	tok, pos, lit := p.scanIgnoreWhitespace()
	switch tok {
	case USE:
		return p.parseUseStatement()
	case CREATE:
		return p.parseCreateStatement()
	case DROP:
		return p.parseDropStatement()
	default:
		return nil, newParseError(tokstr(tok, lit), []string{"USE", "CREATE", "DROP"}, pos)
	}
}

// parseUseStatement parses a string and returns a UseStatement.
// This function assumes the "USE" token has already been consumed.
func (p *Parser) parseUseStatement() (*UseStatement, error) {
	stmt := &UseStatement{}

	// Parse the name of the namespace to be used
	lit, err := p.parseNamespace()
	if err != nil {
		return nil, err
	}
	stmt.name = lit

	return stmt, nil
}

// parseCreateStatement parses a string and returns a Statement AST object.
// This function assumes the "CREATE" token has already been consumed.
func (p *Parser) parseCreateStatement() (Statement, error) {

	// Inspect the first token.
	tok, pos, lit := p.scanIgnoreWhitespace()
	switch tok {
	case NAMESPACE:
		return p.parseCreateNamespaceStatement()
	default:
		return nil, newParseError(tokstr(tok, lit), []string{"NAMESPACE"}, pos)
	}
}

// parseCreateNamespaceStat5ement parses a string and returns a CreateNamespaceStatement.
// This function assumes the "CREATE" token has already been consumed.
func (p *Parser) parseCreateNamespaceStatement() (*CreateNamespaceStatement, error) {
	stmt := &CreateNamespaceStatement{}

	// Parse the name of the namespace to be used
	lit, err := p.parseNamespace()
	if err != nil {
		return nil, err
	}
	stmt.name = lit

	return stmt, nil
}

// parseDropStatement parses a string and returns a Statement AST object.
// This function assumes the "DROP" token has already been consumed.
func (p *Parser) parseDropStatement() (Statement, error) {

	// Inspect the first token.
	tok, pos, lit := p.scanIgnoreWhitespace()
	switch tok {
	case NAMESPACE:
		return p.parseDropNamespaceStatement()
	default:
		return nil, newParseError(tokstr(tok, lit), []string{"NAMESPACE"}, pos)
	}
}

// parseDropNamespaceStat5ement parses a string and returns a DropNamespaceStatement.
// This function assumes the "DROP" token has already been consumed.
func (p *Parser) parseDropNamespaceStatement() (*DropNamespaceStatement, error) {
	stmt := &DropNamespaceStatement{}

	// Parse the name of the namespace to be used
	lit, err := p.parseNamespace()
	if err != nil {
		return nil, err
	}
	stmt.name = lit

	return stmt, nil
}

// parseNamespace returns a namespace title or an error
func (p *Parser) parseNamespace() (string, error) {
	var namespace string
	tok, pos, lit := p.scanIgnoreWhitespace()
	if tok != lexer.IDENT {
		return "", newParseError(tokstr(tok, lit), []string{"namespace"}, pos)
	}
	namespace = lit

	// Scan entire namespace
	// Namespaces are a period delimited list of identifiers
	var endPeriod bool
	for {
		tok, pos, lit = p.scan()
		if tok == lexer.DOT {
			namespace += "."
			endPeriod = true
		} else if tok == lexer.IDENT {
			namespace += lit
			endPeriod = false
		} else {
			break
		}
	}

	// remove last token
	p.unscan()

	// Namespaces can't end on a period
	if endPeriod {
		return "", newParseError(tokstr(tok, lit), []string{"identifier"}, pos)
	}
	return namespace, nil
}

// parserString parses a string.
func (p *Parser) parseString() (string, error) {
	tok, pos, lit := p.scanIgnoreWhitespace()
	if tok != STRING {
		return "", newParseError(tokstr(tok, lit), []string{"string"}, pos)
	}
	return lit, nil
}

// parseIdent parses an identifier.
func (p *Parser) parseIdent() (string, error) {
	tok, pos, lit := p.scanIgnoreWhitespace()
	if tok != lexer.IDENT {
		return "", newParseError(tokstr(tok, lit), []string{"identifier"}, pos)
	}
	return lit, nil
}

// parseInt parses a string and returns an integer literal.
func (p *Parser) parseInt(min, max int) (int, error) {
	tok, pos, lit := p.scanIgnoreWhitespace()
	if tok != lexer.NUMBER {
		return 0, newParseError(tokstr(tok, lit), []string{"number"}, pos)
	}

	// Return an error if the number has a fractional part.
	if strings.Contains(lit, ".") {
		return 0, &ParseError{Message: "number must be an integer", Pos: pos}
	}

	// Convert string to int.
	n, err := strconv.Atoi(lit)
	if err != nil {
		return 0, &ParseError{Message: err.Error(), Pos: pos}
	} else if min > n || n > max {
		return 0, &ParseError{
			Message: fmt.Sprintf("invalid value %d: must be %d <= n <= %d", n, min, max),
			Pos:     pos,
		}
	}

	return n, nil
}

// parseUInt32 parses a string and returns a 32-bit unsigned integer literal.
func (p *Parser) parseUInt32() (uint32, error) {
	tok, pos, lit := p.scanIgnoreWhitespace()
	if tok != lexer.NUMBER {
		return 0, newParseError(tokstr(tok, lit), []string{"number"}, pos)
	}

	// Convert string to unsigned 32-bit integer
	n, err := strconv.ParseUint(lit, 10, 32)
	if err != nil {
		return 0, &ParseError{Message: err.Error(), Pos: pos}
	}

	return uint32(n), nil
}

// parseUInt64 parses a string and returns a 64-bit unsigned integer literal.
func (p *Parser) parseUInt64() (uint64, error) {
	tok, pos, lit := p.scanIgnoreWhitespace()
	if tok != lexer.NUMBER {
		return 0, newParseError(tokstr(tok, lit), []string{"number"}, pos)
	}

	// Convert string to unsigned 64-bit integer
	n, err := strconv.ParseUint(lit, 10, 64)
	if err != nil {
		return 0, &ParseError{Message: err.Error(), Pos: pos}
	}

	return uint64(n), nil
}

// scan returns the next token from the underlying scanner.
func (p *Parser) scan() (tok lexer.Token, pos lexer.Pos, lit string) { return p.s.Scan() }

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.s.Unscan() }

// peekRune returns the next rune that would be read by the scanner.
func (p *Parser) peekRune() rune { return p.s.s.Peek() }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok lexer.Token, pos lexer.Pos, lit string) {
	tok, pos, lit = p.scan()
	if tok == lexer.WS {
		tok, pos, lit = p.scan()
	}
	return
}
