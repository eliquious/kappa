package skl

import "github.com/eliquious/lexer"

func init() {
	lexer.LoadTokenMap(testKeywords)
}

// Token enums
// Built-in Types
const (
	startTypes lexer.Token = iota + 1000
	STRING
	UINT8
	INT8
	UINT16
	INT16
	UINT32
	INT32
	UINT64
	INT64
	FLOAT32
	FLOAT64
	TIMESTAMP
	BOOLEAN
	endTypes

	// Keywords

	startKeywords
	ADD
	BY
	CLUSTERED
	CREATE
	DESCRIBE
	FOR
	FROM
	INSERT
	LIMIT
	LOG
	NAMESPACE
	OFFSET
	ON
	OPTIONAL
	OPTIONS
	PASSWORD
	PERMISSION
	REMOVE
	REQUIRED
	ROLE
	SELECT
	SET
	SHOW
	SUBSCRIBE
	TO
	TYPE
	UNSUBSCRIBE
	UPDATE
	USE
	USER
	USING
	VIEW
	WHERE
	WITH
	endKeywords
)

var testKeywords = map[lexer.Token]string{

	STRING:    "string",
	UINT8:     "uint8",
	INT8:      "int8",
	UINT16:    "uint16",
	INT16:     "int16",
	UINT32:    "uint32",
	INT32:     "int32",
	UINT64:    "uint64",
	INT64:     "int64",
	FLOAT32:   "float32",
	FLOAT64:   "float64",
	TIMESTAMP: "timestamp",
	BOOLEAN:   "boolean",

	ADD:         "ADD",
	BY:          "BY",
	CLUSTERED:   "CLUSTERED",
	CREATE:      "CREATE",
	DESCRIBE:    "DESCRIBE",
	FOR:         "FOR",
	FROM:        "FROM",
	INSERT:      "INSERT",
	LIMIT:       "LIMIT",
	LOG:         "LOG",
	NAMESPACE:   "NAMESPACE",
	OFFSET:      "OFFSET",
	ON:          "ON",
	OPTIONAL:    "OPTIONAL",
	OPTIONS:     "OPTIONS",
	PASSWORD:    "PASSWORD",
	PERMISSION:  "PERMISSION",
	REMOVE:      "REMOVE",
	REQUIRED:    "REQUIRED",
	ROLE:        "ROLE",
	SELECT:      "SELECT",
	SET:         "SET",
	SHOW:        "SHOW",
	SUBSCRIBE:   "SUBSCRIBE",
	TO:          "TO",
	TYPE:        "TYPE",
	UNSUBSCRIBE: "UNSUBSCRIBE",
	UPDATE:      "UPDATE",
	USE:         "USE",
	USER:        "USER",
	USING:       "USING",
	VIEW:        "VIEW",
	WHERE:       "WHERE",
	WITH:        "WITH",
}

// tokstr returns a literal if provided, otherwise returns the token string.
func tokstr(tok lexer.Token, lit string) string {
	if lit != "" && tok != lexer.WS {
		return lit
	}
	return tok.String()
}
