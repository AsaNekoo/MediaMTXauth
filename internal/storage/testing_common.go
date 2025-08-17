package storage

import (
	"MediaMTXAuth/internal"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

func XTestStorage(t *testing.T, s Storage) {
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	t.Run("users", func(t *testing.T) {
		u := internal.User{
			Name:      "test",
			StreamKey: "test",
			Password:  internal.UserPassword{"hash", true},
			Session:   internal.UserSession{123, time.Unix(1234567890, 0)},
		}

		t.Run("not found", func(t *testing.T) {
			storedUser, err := s.GetUser(u.Name)

			if err != nil {
				t.Errorf("Failed to get user: %v", err)
			}

			if storedUser != nil {
				t.Errorf("User should not exists yet: %v", storedUser)
			}
		})

		err := s.SetUser(u)

		if err != nil {
			t.Errorf("Failed to set user: %v", err)
			return
		}

		storedUser, err := s.GetUser(u.Name)

		if err != nil {
			t.Errorf("Failed to get user: %v", err)
			return
		}

		if storedUser == nil {
			t.Errorf("Stored user is nil")
			return
		}

		if !cmp.Equal(u, *storedUser) {
			t.Errorf("Stored user is not equal to stored one")
			t.Logf("expected: %v", u)
			t.Logf("got: %v", *storedUser)
			return
		}
	})

	t.Run("namespaces", func(t *testing.T) {
		n := internal.Namespace{
			Name: "test",
			Sessions: []internal.NamespaceSession{
				{
					Key:           "random key",
					Name:          "test",
					User:          "test_user",
					Created:       time.Unix(1234567890, 0),
					LastPublished: time.Unix(1234567899, 0),
				},
			},
		}

		err := s.SetNamespace(n)

		if err != nil {
			t.Errorf("Failed to set namespace: %v", err)
			return
		}

		storedNamespace, err := s.GetNamespace(n.Name)

		if err != nil {
			t.Errorf("Failed to get user: %v", err)
			return
		}

		if storedNamespace == nil {
			t.Errorf("Stored namespace is nil")
			return
		}

		if !cmp.Equal(n, *storedNamespace) {
			t.Errorf("Stored namespace is not equal to stored one")
			t.Logf("expected: %v", n)
			t.Logf("got: %v", *storedNamespace)
			return
		}
	})
}
