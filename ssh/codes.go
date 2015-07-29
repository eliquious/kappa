package ssh

type StatusCode int

// Success codes
const (
	OK StatusCode = 2000
)

// Authentication related error codes
const (
	Unauthorized StatusCode = iota + 4000
)

// Error codes
const (
	InternalServerError StatusCode = iota + 5000
	InvalidStatementType
	NamespaceDoesNotExist
	UserDoesNotExist
	NamespaceAlreadyExists
	UserAlreadyExists
)

var statusCodes = map[StatusCode]string{

	// Success
	OK: "OK",

	// Security errors
	Unauthorized: "Unauthorized",

	// General errors
	InternalServerError:    "InternalServerError",
	InvalidStatementType:   "InvalidStatementType",
	NamespaceDoesNotExist:  "NamespaceDoesNotExist",
	UserDoesNotExist:       "UserDoesNotExist",
	NamespaceAlreadyExists: "NamespaceAlreadyExists",
	UserAlreadyExists:      "UserAlreadyExists",
}
