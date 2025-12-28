package storage

import (
	"MediaMTXAuth/internal"
	"io"
)

type Storage interface {
	io.Closer
	Init() error

	SetUser(internal.User) error
	GetUser(string) (*internal.User, error)
	GetAllUsers() ([]internal.User, error)
	DeleteUser(string) error

	SetNamespace(internal.Namespace) error
	GetNamespace(string) (*internal.Namespace, error)
	GetAllNamespaces() ([]internal.Namespace, error)
	DeleteNamespace(string) error
}
