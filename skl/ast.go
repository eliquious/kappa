package skl

import "bytes"

// NodeType identifies various AST nodes
type NodeType int

const (
	UseStatementType NodeType = iota
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
	RequiredPermissions() []string
}

// UseStatement represents the USE statement
type UseStatement struct {
	Name string
}

// String returns a string representation
func (s UseStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("USE ")
	buf.WriteString(s.Name)
	return buf.String()
}

// NodeType returns an NodeType id
func (s UseStatement) NodeType() NodeType { return UseStatementType }

// RequiredPermissions returns the required permissions in order to use this command
func (s UseStatement) RequiredPermissions() []string { return []string{} }
