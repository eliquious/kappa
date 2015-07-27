package skl

// Statement is the base interface for all SKL statements
type Statement interface {
	String() string
	RequiredPermissions() []string
}
