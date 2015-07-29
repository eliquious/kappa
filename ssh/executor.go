package ssh

import (
	"fmt"
	"reflect"

	"github.com/subsilent/kappa/datamodel"
	"github.com/subsilent/kappa/skl"
)

// Session provides session and connection related information
type Session struct {
	namespace string
	user      datamodel.User
}

// Executor executes successfully parsed queries
type Executor struct {
	session  Session
	terminal Terminal
	system   datamodel.System
}

// Execute processes each statement
func (e *Executor) Execute(w *ResponseWriter, stmt skl.Statement) {

	switch stmt.NodeType() {
	case skl.UseNamespaceType:
		e.handleUseStatement(w, stmt)
	case skl.CreateNamespaceType:
		e.handleCreateNamespace(w, stmt)
	}
}

func (e *Executor) handleUseStatement(w *ResponseWriter, stmt skl.Statement) {
	use, ok := stmt.(*skl.UseStatement)
	if !ok {
		w.Fail(InvalidStatementType, "expected UseStatement, got %s instead", reflect.TypeOf(stmt))
		return
	}

	// Get user from session
	user := e.session.user
	if user == nil {
		w.Fail(InternalServerError, "could not access user data")
		return
	}

	// Get namespace store
	namespaceStore, err := e.system.Namespaces()
	if err != nil {
		w.Fail(InternalServerError, "could not access namespace data")
		return
	}

	// Get namespace
	name := use.Namespace()

	// Verify namespace existence
	_, err = namespaceStore.Get(name)
	if err == datamodel.ErrNamespaceDoesNotExist {
		w.Fail(NamespaceDoesNotExist, name)
		return
	} else if err != nil {
		w.Fail(InternalServerError, "could not access namespace data")
		return
	}

	// If the user is an admin, grant access
	if user.IsAdmin() {
		e.session.namespace = name
		e.terminal.SetPrompt(fmt.Sprintf("kappa: %s> ", name))
		w.Success(OK, "")
		return
	}

	// Verify user has access to the namespace or is admin
	// 		If user has access, update session namespace and terminal
	// 		Otherwise, return access denied error
	for _, namespace := range user.Namespaces() {
		if namespace == name {
			e.session.namespace = name
			e.terminal.SetPrompt(fmt.Sprintf("kappa: %s> ", name))
			w.Success(OK, "")
			return
		}
	}

	// Otherwise, the user is not authorized
	w.Fail(Unauthorized, "")
}

// Only the admin can create root namespaces.
// Admin can also create sub-namespaces for any existing namespace.
// If the user is not the admin, they must have the 'create.namespace'
//  permission for the parent namespace.
// Root namespaces don't have any periods.
func (e *Executor) handleCreateNamespace(w *ResponseWriter, stmt skl.Statement) {
	create, ok := stmt.(*skl.CreateNamespaceStatement)
	if !ok {
		w.Fail(InvalidStatementType, "expected CreateNamespaceStatement, got %s instead", reflect.TypeOf(stmt))
		return
	}

	// Get user from session
	user := e.session.user
	if user == nil {
		w.Fail(InternalServerError, "could not access user data")
		return
	}

	// // Get namespace store
	// namespaceStore, err := e.system.Namespaces()
	// if err != nil {
	// 	w.Fail(InternalServerError, "could not access namespace data")
	// 	return
	// }

	w.Success(OK, create.Namespace())

	// // Get namespace
	// name := create.Namespace()

	// // Verify namespace existence
	// _, err := namespaceStore.Get(name)

	// // If err == nil, the namespace already existed
	// if err == nil {
	// 	w.Fail(NamespaceAlreadyExists, name)
	// 	return
	// }

	//

	// // Create new namespace
	// namespace, err := namespaceStore.Create(name)

	// // If err !+ nil, namespace could not be created
	// if err == datamodel.ErrNamespaceDoesNotExist {
	// 	w.Fail(NamespaceDoesNotExist, name)
	// 	return
	// } else if err != nil {
	// 	w.Fail(InternalServerError, "could not access namespace data")
	// 	return
	// }

	// // If the user is an admin, grant access
	// if user.IsAdmin() {
	// 	e.session.namespace = name
	// 	e.terminal.SetPrompt(fmt.Sprintf("kappa: %s> ", name))
	// 	w.Success(OK, "")
	// 	return
	// }

	// // Verify user has access to the namespace or is admin
	// // 		If user has access, update session namespace and terminal
	// // 		Otherwise, return access denied error
	// for _, namespace := range user.Namespaces() {
	// 	if namespace == name {
	// 		e.session.namespace = name
	// 		e.terminal.SetPrompt(fmt.Sprintf("kappa: %s> ", name))
	// 		w.Success(OK, "")
	// 		return
	// 	}
	// }

	// // Otherwise, the user is not authorized
	// w.Fail(Unauthorized, "")
}
