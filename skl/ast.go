package skl

import (
	"bytes"
	"strings"
)

// NodeType identifies various AST nodes
type NodeType int

const (
	UseNamespaceType    NodeType = iota
	CreateNamespaceType NodeType = iota
	ShowNamespaceType   NodeType = iota
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
	RequiredPermissions() string
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
func (s UseStatement) RequiredPermissions() string { return "" }

// CreateNamespaceStatement represents the CREATE NAMESPACE statement
type CreateNamespaceStatement struct {
	name string
}

// Namespace returns the namespace being requested
func (s CreateNamespaceStatement) Namespace() string {
	return s.name
}

// IsRootNamespace determines if the namespace to be created is a top-level namespace
func (s CreateNamespaceStatement) IsRootNamespace() bool {
	return !strings.Contains(s.name, ".")
}

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
func (s CreateNamespaceStatement) RequiredPermissions() string { return "create.namespace" }

// CreateNamespaceStatement represents the SHOW NAMESPACES statement
type ShowNamespacesStatement struct {
}

// String returns a string representation
func (s ShowNamespacesStatement) String() string {
	var buf bytes.Buffer
	buf.WriteString("SHOW NAMESPACES")
	return buf.String()
}

// NodeType returns an NodeType id
func (s ShowNamespacesStatement) NodeType() NodeType { return ShowNamespaceType }

// RequiredPermissions returns the required permissions in order to use this command
func (s ShowNamespacesStatement) RequiredPermissions() string { return "show.namespaces" }
