package memory

import (
	"MediaMTXAuth/internal"
)

type Storage struct {
	Users      map[string]internal.User
	Namespaces map[string]internal.Namespace
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) Init() error {
	if s == nil {
		return nil
	}

	if s.Users == nil {
		s.Users = make(map[string]internal.User)
	}

	if s.Namespaces == nil {
		s.Namespaces = make(map[string]internal.Namespace)
	}

	return nil
}

func (s *Storage) SetUser(u internal.User) error {
	if s != nil {
		if s.Users == nil {
			s.Users = make(map[string]internal.User)
		}

		s.Users[u.Name] = u
	}
	return nil
}

func (s *Storage) GetUser(name string) (*internal.User, error) {
	if s != nil && s.Users != nil {
		if u, ok := s.Users[name]; ok {
			return &u, nil
		}
	}
	return nil, nil
}

func (s *Storage) SetNamespace(n internal.Namespace) error {
	if s != nil {
		if s.Namespaces == nil {
			s.Namespaces = make(map[string]internal.Namespace)
		}

		s.Namespaces[n.Name] = n
	}
	return nil
}

func (s *Storage) GetNamespace(name string) (*internal.Namespace, error) {
	if s != nil && s.Namespaces != nil {
		if n, ok := s.Namespaces[name]; ok {
			return &n, nil
		}
	}
	return nil, nil
}

func (s *Storage) DeleteUser(name string) error {
	if s != nil && s.Users != nil {
		delete(s.Users, name)
	}
	return nil
}

func (s *Storage) DeleteNamespace(name string) error {
	if s != nil && s.Namespaces != nil {
		delete(s.Namespaces, name)
	}
	return nil
}

func (s *Storage) Clear() {
	clear(s.Users)
	clear(s.Namespaces)
}

func (s *Storage) GetAllUsers() ([]internal.User, error) {
	if s == nil || s.Users == nil {
		return []internal.User{}, nil
	}

	users := make([]internal.User, 0, len(s.Users))
	for _, user := range s.Users {
		users = append(users, user)
	}

	return users, nil
}
