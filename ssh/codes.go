package ssh

type StatusCode int

const (
	OK                    StatusCode = 2000
	Unauthorized          StatusCode = 4001
	InternalServerError   StatusCode = 5000
	InvalidStatementType  StatusCode = 5001
	NamespaceDoesNotExist StatusCode = 5002
	UserDoesNotExist      StatusCode = 5003
)

var statusCodes = map[StatusCode]string{
	OK:                    "OK",
	Unauthorized:          "Unauthorized",
	InternalServerError:   "InternalServerError",
	InvalidStatementType:  "InvalidStatementType",
	NamespaceDoesNotExist: "NamespaceDoesNotExist",
	UserDoesNotExist:      "UserDoesNotExist",
}
