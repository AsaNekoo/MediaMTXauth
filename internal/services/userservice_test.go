package services

import (
	"MediaMTXAuth/internal"
	"MediaMTXAuth/internal/storage/memory"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const username = "username"
const password = "password"

func TestUserService(t *testing.T) {
	storage := &memory.Storage{}
	if err := storage.Init(); err != nil {
		t.Fatal(err)
	}

	userService := NewUserService(storage)

	t.Run("create user", func(t *testing.T) {
		t.Run("valid user creation", func(t *testing.T) {
			t.Cleanup(storage.Clear)
			user, err := userService.Create(username, password, false, "")
			if err != nil {
				t.Errorf("Failed to create user: %v", err)
				return
			}

			if user == nil {
				t.Errorf("Created user is nil")
				return
			}
		})

		t.Run("duplicate user creation", func(t *testing.T) {
			t.Cleanup(storage.Clear)
			_, err := userService.Create(username, password, false, "")
			if err != nil {
				t.Errorf("Failed to create user: %v", err)
			}
			_, err = userService.Create(username, password, false, "")
			if err != internal.ErrUserAlreadyExists {
				t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
			}
		})

		t.Run("validation errors", func(t *testing.T) {
			t.Cleanup(storage.Clear)
			_, err := userService.Create("ab", password, false, "")
			if err != ErrShortUsername {
				t.Errorf("Expected ErrShortUsername, got %v", err)
			}

			_, err = userService.Create(username, "short", false, "")
			if err != ErrShortPassword {
				t.Errorf("Expected ErrShortPassword, got %v", err)
			}
		})
	})

	t.Run("get all users", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		user, err := userService.Create(username, password, true, "")
		if err != nil {
			t.Errorf("Failed to create user: %v", err)
			return
		}

		users, err := userService.GetAllUsers()
		if err != nil {
			t.Errorf("Failed to get all users: %v", err)
			return
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
			return
		}

		if !cmp.Equal(users[0], *user) {
			t.Errorf("Expected user: %v, got: %v", user, users[0])
		}

		user, err = userService.Create("username2", password, true, "")
		if err != nil {
			t.Errorf("Failed to create user: %v", err)
			return
		}

		users, err = userService.GetAllUsers()
		if err != nil {
			t.Errorf("Failed to get all users: %v", err)
			return
		}

		if len(users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users))
			return
		}

		if !cmp.Equal(users[1], *user) {
			t.Errorf("Expected user: %v, got: %v", user, users[1])
		}
	})

	t.Run("get user", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		createdUser, err := userService.Create(username, password, true, "")

		if err != nil {
			t.Errorf("Failed to create user: %v", err)
			return
		}

		retrievedUser, err := userService.Get(username)
		if err != nil {
			t.Errorf("Failed to get user: %v", err)
			return
		}

		if !cmp.Equal(*createdUser, *retrievedUser) {
			t.Errorf("Created and retrieved users are not equal")
			t.Logf("expected: %v", *createdUser)
			t.Logf("got: %v", *retrievedUser)
		}
	})

	t.Run("delete user", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := userService.Create(username, password, false, "")

		err = userService.Delete(username)
		if err != nil {
			t.Errorf("Failed to delete user: %v", err)
			return
		}

		deletedUser, err := userService.Get(username)
		if deletedUser != nil {
			t.Errorf("User shouldnt exist, got %v", deletedUser)
		}
	})

	t.Run("change password", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		oldPassword := "oldpassword"
		newPassword := "newpassword"

		user, err := userService.Create(username, oldPassword, false, "")

		originalHash := user.Password.Hash

		err = userService.ChangePassword(username, newPassword)
		if err != nil {
			t.Errorf("Failed to change password: %v", err)
			return
		}

		updatedUser, err := userService.Get(username)

		if updatedUser.Password.Hash == originalHash {
			t.Errorf("Password should have changed")
		}
	})

	t.Run("change password for non-existent user", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		err := userService.ChangePassword("nouser", "newpassword")
		if err != internal.ErrUserNotFound {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("reset password", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		user, err := userService.Create(username, password, false, "")

		if err != nil {
			t.Errorf("Failed to create user: %v", err)
			return
		}

		originalHash := user.Password.Hash

		newPassword, err := userService.ResetPassword(username)
		if err != nil {
			t.Errorf("Failed to reset password: %v", err)
			return
		}

		if newPassword == "" {
			t.Errorf("Reset password should not be empty")
		}

		updatedUser, err := userService.Get(username)

		if updatedUser.Password.Hash == originalHash {
			t.Errorf("Password hash should have changed")
		}
	})

	t.Run("reset password for non-existent user", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := userService.ResetPassword("nouser")
		if err != internal.ErrUserNotFound {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("reset stream key", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		user, err := userService.Create(username, password, false, "")

		if err != nil {
			t.Errorf("Failed to create user: %v", err)
			return
		}

		originalStreamKey := user.StreamKey

		newStreamKey, err := userService.ResetStreamKey(username)
		if err != nil {
			t.Errorf("Failed to reset stream key: %v", err)
			return
		}

		if newStreamKey == originalStreamKey {
			t.Errorf("Stream key didnt change")
		}

		updatedUser, err := userService.Get(username)

		if updatedUser.StreamKey != newStreamKey {
			t.Errorf("Expected stream key %s, got %s", newStreamKey, updatedUser.StreamKey)
		}
	})
	t.Run("reset stream key for non-existent user", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := userService.ResetStreamKey("nouser")
		if err != internal.ErrUserNotFound {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("login and logout", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, _ = userService.Create(username, password, false, "")

		t.Run("successful login", func(t *testing.T) {
			user, err := userService.Login(username, password)
			if err != nil {
				t.Errorf("Failed to login: %v", err)
				return
			}

			if user.Session.ID == 0 {
				t.Errorf("Session ID shouldnt be zero")
			}
		})

		t.Run("failed login", func(t *testing.T) {
			_, err := userService.Login(username, "wrongpassword")
			if err != internal.ErrWrongPassword {
				t.Errorf("Expected ErrWrongPassword, got %v", err)
			}
		})

		t.Run("logout", func(t *testing.T) {

			_, err := userService.Login(username, password)

			if err != nil {
				t.Errorf("Failed to login: %v", err)
				return
			}

			_, err = userService.Logout(username)

			if err != nil {
				t.Errorf("Failed to logout: %v", err)
				return
			}

			user, err := userService.Get(username)

			if user.Session.ID != 0 {
				t.Errorf("Session ID should be cleared, got %d", user.Session.ID)
			}

			if !user.Session.Expiration.IsZero() {
				t.Errorf("Session expiration should be zero")
			}
		})
	})

	t.Run("login in to non-existent user", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := userService.Login("nouser", "password")
		if err != internal.ErrUserNotFound {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("create default admin user", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		result, err := userService.CreateDefaultAdminUser()
		if err != nil {
			t.Errorf("Failed to create: %v", err)
			return
		}

		expectedResult := "admin"
		if result != expectedResult {
			t.Errorf("Expected result %s, got %s", expectedResult, result)
		}

		adminUser, err := userService.Get("admin")

		if adminUser == nil {
			t.Errorf("Admin should exist")
			return
		}

		if adminUser.IsAdmin != true {
			t.Errorf("Expected admin, got %v", adminUser.IsAdmin)
		}
	})

	t.Run("verify session", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, err := userService.Create(username, password, false, "")
		if err != nil {
			t.Errorf("Failed to create user: %v", err)
			return
		}

		user, err := userService.Login(username, password)
		if err != nil {
			t.Errorf("Failed to login: %v", err)
			return
		}

		sessionKey := fmt.Sprintf("%d", user.Session.ID)

		t.Run("valid session", func(t *testing.T) {
			valid, err := userService.VerifySession(username, sessionKey)
			if err != nil {
				t.Errorf("VerifySession returned error: %v", err)
			}
			if !valid {
				t.Errorf("Expected valid session, got invalid")
			}
		})

		t.Run("invalid session key", func(t *testing.T) {
			valid, err := userService.VerifySession(username, "wrongkey")
			if err != nil {
				t.Errorf("VerifySession returned error: %v", err)
			}
			if valid {
				t.Errorf("Expected invalid session, got valid")
			}
		})

		t.Run("expired session", func(t *testing.T) {
			user.Session.Expiration = time.Now().Add(-time.Hour)
			_ = storage.SetUser(*user)
			valid, err := userService.VerifySession(username, sessionKey)
			if err != nil {
				t.Errorf("VerifySession returned error: %v", err)
			}
			if valid {
				t.Errorf("Expected invalid session due to expiration, got valid")
			}
		})

		t.Run("user not found", func(t *testing.T) {
			valid, err := userService.VerifySession("nouser", sessionKey)
			if err != internal.ErrUserNotFound {
				t.Errorf("Expected ErrUserNotFound, got %v", err)
			}
			if valid {
				t.Errorf("Expected invalid session for missing user, got valid")
			}
		})
	})
}
