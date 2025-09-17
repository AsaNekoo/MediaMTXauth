package services

import (
	"MediaMTXAuth/internal"
	"MediaMTXAuth/internal/storage/memory"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const namespace = "namespace"
const session = "session"

func TestNamespaceService(t *testing.T) {
	storage := &memory.Storage{}
	if err := storage.Init(); err != nil {
		t.Fatal(err)
	}

	namespaceService := NewNamespaceService(storage)

	t.Run("Create namespace", func(t *testing.T) {
		t.Run("valid namespace creation", func(t *testing.T) {
			t.Cleanup(storage.Clear)
			user, err := namespaceService.Create(namespace)
			if err != nil {
				t.Errorf("Failed to create namespace: %v", err)
				return
			}

			if user == nil {
				t.Errorf("Created namespace is nil")
				return
			}
		})
		t.Run("duplicate namespace creation", func(t *testing.T) {
			t.Cleanup(storage.Clear)
			_, err := namespaceService.Create(namespace)
			if err != nil {
				t.Errorf("Failed to create namespace: %v", err)
			}
			_, err = namespaceService.Create(namespace)
			if err != internal.ErrNamespaceAlreadyExists {
				t.Errorf("Expected ErrNamespaceAlreadyExists, got %v", err)
			}
		})
	})

	t.Run("get namespace", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		createdNamespace, err := namespaceService.Create(namespace)

		if err != nil {
			t.Errorf("Failed to create namespace: %v", err)
			return
		}

		retrievedNamespace, err := namespaceService.Get(namespace)
		if err != nil {
			t.Errorf("Failed to get user: %v", err)
			return
		}

		if !cmp.Equal(*createdNamespace, *retrievedNamespace) {
			t.Errorf("Created and retrieved users are not equal")
			t.Logf("expected: %v", *createdNamespace)
			t.Logf("got: %v", *retrievedNamespace)
		}
	})

	t.Run("get from non-existent namespace", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := namespaceService.Get("nonexistent")
		if err != internal.ErrNamespaceNotFound {
			t.Errorf("Expected ErrNamespaceNotFound, got %v", err)
		}
	})

	t.Run("delete namespace", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := namespaceService.Create(namespace)

		err = namespaceService.Delete(namespace)
		if err != nil {
			t.Errorf("Failed to delete namespace: %v", err)
			return
		}

		deletedUser, err := namespaceService.Get(namespace)
		if deletedUser != nil {
			t.Errorf("namespace shouldnt exist, got %v", deletedUser)
		}
	})
	t.Run("add session", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := namespaceService.Create(namespace)
		if err != nil {
			t.Errorf("Failed to create namespace: %v", err)
		}
		_, err = namespaceService.AddSession(namespace, session, username)
		if err != nil {
			t.Errorf("Failed to add session: %v", err)
			return
		}

		retrievedNamespace, err := namespaceService.Get(namespace)
		if err != nil {
			t.Errorf("Failed to get namespace: %v", err)
			return
		}

		if len(retrievedNamespace.Sessions) != 1 {
			t.Errorf("Expected 1 session, got %d", len(retrievedNamespace.Sessions))
		}
	})

	t.Run("add session to non-existent namespace", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := namespaceService.AddSession("nonexistent", session, username)
		if err != internal.ErrNamespaceNotFound {
			t.Errorf("Expected ErrNamespaceNotFound, got %v", err)
		}
	})

	t.Run("session removal", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := namespaceService.Create(namespace)
		if err != nil {
			t.Errorf("Failed to create namespace: %v", err)
			return
		}

		addedSession, err := namespaceService.AddSession(namespace, session, username)
		if err != nil {
			t.Errorf("Failed to add session: %v", err)
			return
		}

		err = namespaceService.RemoveSession(namespace, addedSession.Key)
		if err != nil {
			t.Errorf("Failed to remove session: %v", err)
			return
		}

		retrievedNamespace, err := namespaceService.Get(namespace)
		if err != nil {
			t.Errorf("Failed to get namespace: %v", err)
			return
		}

		if len(retrievedNamespace.Sessions) != 0 {
			t.Errorf("Expected 0 session, got %d", len(retrievedNamespace.Sessions))
		}
	})
	t.Run("remove session from non-existent namespace", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		err := namespaceService.RemoveSession("nonexistent", "somesession")
		if err != internal.ErrNamespaceNotFound {
			t.Errorf("Expected ErrNamespaceNotFound, got %v", err)
		}
	})
}
