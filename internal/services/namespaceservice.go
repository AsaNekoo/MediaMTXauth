package services

import (
	"MediaMTXAuth/internal"
	"MediaMTXAuth/internal/storage"
	"crypto/rand"
	"time"
)

type namespaceService struct {
	storage storage.Storage
}

func NewNamespaceService(storage storage.Storage) internal.NamespaceService {
	return &namespaceService{storage}
}

func (s *namespaceService) Create(namespaceName string) (*internal.Namespace, error) {
	existingNamespace, err := s.storage.GetNamespace(namespaceName)
	if err != nil {
		return nil, err
	}
	if existingNamespace != nil {
		return nil, internal.ErrNamespaceAlreadyExists
	}
	namespace := internal.Namespace{
		Name:     namespaceName,
		Sessions: []internal.NamespaceSession{},
	}
	err = s.storage.SetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return &namespace, nil
}

func (s *namespaceService) Get(namespaceName string) (*internal.Namespace, error) {
	namespace, _ := s.storage.GetNamespace(namespaceName)
	if namespace == nil {
		return nil, internal.ErrNamespaceNotFound
	}
	return namespace, nil
}

func (s *namespaceService) GetAllNamespaces() ([]internal.Namespace, error) {
	return s.storage.GetAllNamespaces()
}

func (s *namespaceService) Delete(namespaceName string) error {
	return s.storage.DeleteNamespace(namespaceName)
}

func (s *namespaceService) AddSession(namespaceName, sessionName, user string) (*internal.NamespaceSession, error) {
	namespace, _ := s.storage.GetNamespace(namespaceName)
	if namespace == nil {
		return nil, internal.ErrNamespaceNotFound
	}

	sessionKey := rand.Text()

	newSession := internal.NamespaceSession{
		Key:     sessionKey,
		Name:    sessionName,
		User:    user,
		Created: time.Now(),
	}

	namespace.Sessions = append(namespace.Sessions, newSession)

	if err := s.storage.SetNamespace(*namespace); err != nil {
		return nil, err
	}

	return &newSession, nil
}

func (s *namespaceService) RemoveSession(namespaceName, sessionKey string) error {
	namespace, _ := s.storage.GetNamespace(namespaceName)
	if namespace == nil {
		return internal.ErrNamespaceNotFound
	}

	newSessions := make([]internal.NamespaceSession, 0, len(namespace.Sessions))
	for _, sess := range namespace.Sessions {
		if sess.Key == sessionKey {
			continue
		}
		newSessions = append(newSessions, sess)
	}

	namespace.Sessions = newSessions
	if err := s.storage.SetNamespace(*namespace); err != nil {
		return err
	}

	return nil
}
