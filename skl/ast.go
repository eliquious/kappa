package skl

<<<<<<< Updated upstream
import "bytes"
=======
import (
	"bytes"
	"strings"
)
>>>>>>> Stashed changes

// NodeType identifies various AST nodes
type NodeType int

const (
	UseNamespaceType    NodeType = iota
	CreateNamespaceType NodeType = iota
)

// Node is an interface for AST nodes
type Node interface {
	NodeType() NodeType
	String() string
}

// ExprType identifies various expressions
type ExprType int

// Expr represents AST expressions
type Expr interface {
	Node
	ExprType() ExprType
}

// Statement is the interface for all SKL statements
type Statement interface {
	Node
<<<<<<< Updated upstream
	RequiredPermissions() []string
=======
	RequiredPermissions() string
>>>>>>> Stashed changes
}

// UseStatement represents the USE statement
type UseStatement struct {
	name string
}

// Namespace returns the namespace being requested
func (u UseStatement) Namespace() string {
	return u.name
}

// String returns a string representation
func (s UseStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("USE ")
	buf.WriteString(s.name)
	return buf.String()
}

// NodeType returns an NodeType id
func (s UseStatement) NodeType() NodeType { return UseNamespaceType }

// RequiredPermissions returns the required permissions in order to use this command
<<<<<<< Updated upstream
func (s UseStatement) RequiredPermissions() []string { return []string{} }
=======
func (s UseStatement) RequiredPermissions() string { return "" }
>>>>>>> Stashed changes

// CreateNamespaceStatement represents the CREATE NAMESPACE statement
type CreateNamespaceStatement struct {
	name string
}

// Namespace returns the namespace being requested
func (s CreateNamespaceStatement) Namespace() string {
	return s.name
}

<<<<<<< Updated upstream
=======
// IsRootNamespace determines if the namespace to be created is a top-level namespace
func (s CreateNamespaceStatement) IsRootNamespace() bool {
	return !strings.Contains(s.name, ".")
}

>>>>>>> Stashed changes
// String returns a string representation
func (s CreateNamespaceStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("CREATE NAMESPACE ")
	buf.WriteString(s.name)
	return buf.String()
}

// NodeType returns an NodeType id
func (s CreateNamespaceStatement) NodeType() NodeType { return CreateNamespaceType }

// RequiredPermissions returns the required permissions in order to use this command
<<<<<<< Updated upstream
func (s CreateNamespaceStatement) RequiredPermissions() []string { return []string{"create.namespace"} }
=======
func (s CreateNamespaceStatement) RequiredPermissions() string { return "create.namespace" }
>>>>>>> Stashed changes
