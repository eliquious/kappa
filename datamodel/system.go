package datamodel

type System interface {
    Users() UserStore
    Namespaces() NamespaceStore
}
