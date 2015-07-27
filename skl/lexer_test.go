package skl

import (
	"reflect"
	"strings"
	"testing"

	"github.com/eliquious/lexer"
)

// Ensure the scanner can scan tokens correctly.
func TestScanner_Scan(t *testing.T) {
	var tests = []struct {
		s   string
		tok lexer.Token
		lit string
		pos lexer.Pos
	}{

		// Types
		{s: `STRING`, tok: STRING},
		{s: `UINT8`, tok: UINT8},
		{s: `INT8`, tok: INT8},
		{s: `UINT16`, tok: UINT16},
		{s: `INT16`, tok: INT16},
		{s: `UINT32`, tok: UINT32},
		{s: `INT32`, tok: INT32},
		{s: `UINT64`, tok: UINT64},
		{s: `INT64`, tok: INT64},
		{s: `FLOAT32`, tok: FLOAT32},
		{s: `FLOAT64`, tok: FLOAT64},
		{s: `TIMESTAMP`, tok: TIMESTAMP},
		{s: `BOOLEAN`, tok: BOOLEAN},

		// Keywords
		{s: `ADD`, tok: ADD},
		{s: `BY`, tok: BY},
		{s: `CLUSTERED`, tok: CLUSTERED},
		{s: `CREATE`, tok: CREATE},
		{s: `DESCRIBE`, tok: DESCRIBE},
		{s: `FOR`, tok: FOR},
		{s: `FROM`, tok: FROM},
		{s: `INSERT`, tok: INSERT},
		{s: `LIMIT`, tok: LIMIT},
		{s: `LOG`, tok: LOG},
		{s: `NAMESPACE`, tok: NAMESPACE},
		{s: `OFFSET`, tok: OFFSET},
		{s: `ON`, tok: ON},
		{s: `OPTIONAL`, tok: OPTIONAL},
		{s: `OPTIONS`, tok: OPTIONS},
		{s: `PASSWORD`, tok: PASSWORD},
		{s: `PERMISSION`, tok: PERMISSION},
		{s: `REMOVE`, tok: REMOVE},
		{s: `REQUIRED`, tok: REQUIRED},
		{s: `ROLE`, tok: ROLE},
		{s: `SELECT`, tok: SELECT},
		{s: `SET`, tok: SET},
		{s: `SHOW`, tok: SHOW},
		{s: `SUBSCRIBE`, tok: SUBSCRIBE},
		{s: `TO`, tok: TO},
		{s: `TYPE`, tok: TYPE},
		{s: `UNSUBSCRIBE`, tok: UNSUBSCRIBE},
		{s: `UPDATE`, tok: UPDATE},
		{s: `USE`, tok: USE},
		{s: `USER`, tok: USER},
		{s: `USING`, tok: USING},
		{s: `VIEW`, tok: VIEW},
		{s: `WHERE`, tok: WHERE},
		{s: `WITH`, tok: WITH},
	}

	for i, tt := range tests {
		s := lexer.NewScanner(strings.NewReader(tt.s))
		tok, pos, lit := s.Scan()
		if tt.tok != tok {
			t.Errorf("%d. %q token mismatch: exp=%q got=%q <%q>", i, tt.s, tt.tok, tok, lit)
		} else if tt.pos.Line != pos.Line || tt.pos.Char != pos.Char {
			t.Errorf("%d. %q pos mismatch: exp=%#v got=%#v", i, tt.s, tt.pos, pos)
		} else if tt.lit != lit {
			t.Errorf("%d. %q literal mismatch: exp=%q got=%q", i, tt.s, tt.lit, lit)
		}
	}
}

// Ensure the scanner can scan a series of tokens correctly.
func TestScanner_Scan_Multi(t *testing.T) {
	type result struct {
		tok lexer.Token
		pos lexer.Pos
		lit string
	}
	exp := []result{
		{tok: SELECT, pos: lexer.Pos{Line: 0, Char: 0}, lit: ""},
		{tok: lexer.WS, pos: lexer.Pos{Line: 0, Char: 6}, lit: " "},
		{tok: lexer.IDENT, pos: lexer.Pos{Line: 0, Char: 7}, lit: "value"},
		{tok: lexer.WS, pos: lexer.Pos{Line: 0, Char: 12}, lit: " "},
		{tok: FROM, pos: lexer.Pos{Line: 0, Char: 13}, lit: ""},
		{tok: lexer.WS, pos: lexer.Pos{Line: 0, Char: 17}, lit: " "},
		{tok: lexer.IDENT, pos: lexer.Pos{Line: 0, Char: 18}, lit: "myseries"},
		{tok: lexer.WS, pos: lexer.Pos{Line: 0, Char: 26}, lit: " "},
		{tok: WHERE, pos: lexer.Pos{Line: 0, Char: 27}, lit: ""},
		{tok: lexer.WS, pos: lexer.Pos{Line: 0, Char: 32}, lit: " "},
		{tok: lexer.IDENT, pos: lexer.Pos{Line: 0, Char: 33}, lit: "a"},
		{tok: lexer.WS, pos: lexer.Pos{Line: 0, Char: 34}, lit: " "},
		{tok: lexer.EQ, pos: lexer.Pos{Line: 0, Char: 35}, lit: ""},
		{tok: lexer.WS, pos: lexer.Pos{Line: 0, Char: 36}, lit: " "},
		{tok: lexer.STRING, pos: lexer.Pos{Line: 0, Char: 36}, lit: "b"},
		{tok: lexer.EOF, pos: lexer.Pos{Line: 0, Char: 40}, lit: ""},
	}

	// Create a scanner.
	v := `SELECT value from myseries WHERE a = 'b'`
	s := lexer.NewScanner(strings.NewReader(v))

	// Continually scan until we reach the end.
	var act []result
	for {
		tok, pos, lit := s.Scan()
		act = append(act, result{tok, pos, lit})
		if tok == lexer.EOF {
			break
		}
	}

	// Verify the token counts match.
	if len(exp) != len(act) {
		t.Fatalf("token count mismatch: exp=%d, got=%d", len(exp), len(act))
	}

	// Verify each token matches.
	for i := range exp {
		if !reflect.DeepEqual(exp[i], act[i]) {
			t.Fatalf("%d. token mismatch:\n\nexp=%#v\n\ngot=%#v", i, exp[i], act[i])
		}
	}
}
