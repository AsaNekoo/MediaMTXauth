package internal

import (
	"errors"
	"time"
)

type UserPassword struct {
	Hash        string
	IsGenerated bool
}

type UserSession struct {
	ID         uint64
	Expiration time.Time
}

type User struct {
	Name      string
	StreamKey string
	IsAdmin   bool
	Password  UserPassword
	Session   UserSession
	Namespace string
}

func (ns User) GetID() string {
	return ns.Name
}

type NamespaceSession struct {
	Key  string
	Name string
	User string

	Created time.Time
}

func (ns NamespaceSession) GetID() string {
	return ns.Key
}

type Namespace struct {
	Name     string
	Sessions []NamespaceSession
}

func (ns Namespace) GetID() string {
	return ns.Name
}

type WithID interface {
	GetID() string
}

type UserService interface {
	Create(username, password string, isAdmin bool, namespace string) (*User, error)
	CreateDefaultAdminUser() (string, error)
	Get(username string) (*User, error)
	Delete(name string) error
	GetAllUsers() ([]User, error)

	ChangePassword(username, password string) error
	ResetPassword(username string) (string, error)
	ResetStreamKey(username string) (string, error)
	Login(username, password string) (*User, error)
	Logout(username string) (*User, error)
	VerifySession(username, sessionID string) (bool, error)
}

type NamespaceService interface {
	Create(name string) (*Namespace, error)
	Get(name string) (*Namespace, error)
	GetAllNamespaces() ([]Namespace, error)
	Delete(name string) error

	AddSession(namespace, sessionName, user string) (*NamespaceSession, error)
	RemoveSession(namespace, sessionKey string) error
}

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrWrongPassword          = errors.New("wrong password")
	ErrNamespaceNotFound      = errors.New("namespace not found")
	ErrNamespaceAlreadyExists = errors.New("namespace already exists")
	ErrSessionNotFound        = errors.New("session not found")
)
