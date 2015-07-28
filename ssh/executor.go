package ssh

import (
	"fmt"
	"reflect"

	"github.com/subsilent/kappa/datamodel"
	"github.com/subsilent/kappa/skl"
	"golang.org/x/crypto/ssh/terminal"
)

type Session struct {
	namespace string
	user      datamodel.User
	system    datamodel.System
	terminal  *terminal.Terminal
}

type Executor struct {
	Session Session
}

func (e *Executor) Execute(w *ResponseWriter, stmt skl.Statement) {

	switch stmt.NodeType() {
	case skl.UseStatementType:
		e.handleUseStatement(w, stmt)
	}
}

func (e *Executor) handleUseStatement(w *ResponseWriter, stmt skl.Statement) {
	use, ok := stmt.(*skl.UseStatement)
	if !ok {
		w.Fail(InvalidStatementType, "expected UseStatement, got %s instead", reflect.TypeOf(stmt))
		return
	}

	// Get user from session
	user := e.Session.user
	if user == nil {
		w.Fail(InternalServerError, "could not access user data")
		return
	}

	// Get namespace store
	namespaceStore, err := e.Session.system.Namespaces()
	if err != nil {
		w.Fail(InternalServerError, "could not access namespace data")
		return
	}

	// Verify namespace existence
	_, err = namespaceStore.Get(use.Name)
	if err == datamodel.ErrNamespaceDoesNotExist {
		w.Fail(NamespaceDoesNotExist, use.Name)
		return
	} else if err != nil {
		w.Fail(InternalServerError, "could not access namespace data")
		return
	}

	// If the user is an admin, grant access
	if user.IsAdmin() {
		e.Session.namespace = use.Name
		e.Session.terminal.SetPrompt(fmt.Sprintf("kappa: %s> ", use.Name))
		w.Success(OK, "")
		return
	}

	// Verify user has access to the namespace or is admin
	// 		If user has access, update session namespace and terminal
	// 		Otherwise, return access denied error
	for _, namespace := range user.Namespaces() {
		if namespace == use.Name {
			e.Session.namespace = use.Name
			e.Session.terminal.SetPrompt(fmt.Sprintf("kappa: %s> ", use.Name))
			w.Success(OK, "")
			return
		}
	}

	// Otherwise, the user is not authorized
	w.Fail(Unauthorized, "")
}
