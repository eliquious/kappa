package ssh

type StatusCode int

// Success codes
const (
<<<<<<< Updated upstream
	OK StatusCode = 2000
=======
	OK StatusCode = iota + 2000
	NamespaceAlreadyExists
	UserAlreadyExists
>>>>>>> Stashed changes
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
<<<<<<< Updated upstream
	NamespaceAlreadyExists
	UserAlreadyExists
=======
	CreateNamespaceError
>>>>>>> Stashed changes
)

var statusCodes = map[StatusCode]string{

	// Success
	OK: "OK",
<<<<<<< Updated upstream
=======
	NamespaceAlreadyExists: "NamespaceAlreadyExists",
	UserAlreadyExists:      "UserAlreadyExists",
>>>>>>> Stashed changes

	// Security errors
	Unauthorized: "Unauthorized",

	// General errors
<<<<<<< Updated upstream
	InternalServerError:    "InternalServerError",
	InvalidStatementType:   "InvalidStatementType",
	NamespaceDoesNotExist:  "NamespaceDoesNotExist",
	UserDoesNotExist:       "UserDoesNotExist",
	NamespaceAlreadyExists: "NamespaceAlreadyExists",
	UserAlreadyExists:      "UserAlreadyExists",
=======
	InternalServerError:   "InternalServerError",
	InvalidStatementType:  "InvalidStatementType",
	NamespaceDoesNotExist: "NamespaceDoesNotExist",
	UserDoesNotExist:      "UserDoesNotExist",
	CreateNamespaceError:  "CreateNamespaceError",
>>>>>>> Stashed changes
}
