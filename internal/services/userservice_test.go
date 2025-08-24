package services

import (
	"MediaMTXAuth/internal/storage/memory"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const username = "username"
const password = "password"

func TestUserService(t *testing.T) {
	storage := &memory.Storage{}
	if err := storage.Init(); err != nil {
		t.Fatal(err)
	}

	userService := &userService{storage: storage}

	t.Run("create user", func(t *testing.T) {
		t.Run("valid user creation", func(t *testing.T) {
			user, err := userService.Create(username, password, false)
			if err != nil {
				t.Errorf("Failed to create user: %v", err)
				return
			}

			if user == nil {
				t.Errorf("Created user is nil")
				return
			}
		})

		t.Run("validation errors", func(t *testing.T) {
			_, err := userService.Create("ab", password, false)
			if err != ErrUsername {
				t.Errorf("Expected ErrUsername, got %v", err)
			}

			_, err = userService.Create(username, "short", false)
			if err != ErrPassword {
				t.Errorf("Expected ErrPassword, got %v", err)
			}
		})
	})

	t.Run("get user", func(t *testing.T) {

		createdUser, err := userService.Create(username, password, true)

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

		_, err := userService.Create(username, password, false)

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
		oldPassword := "oldpassword"
		newPassword := "newpassword"

		user, err := userService.Create(username, oldPassword, false)

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

	t.Run("reset password", func(t *testing.T) {

		user, err := userService.Create(username, password, false)

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

	t.Run("reset stream key", func(t *testing.T) {

		user, err := userService.Create(username, password, false)

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

	t.Run("login and logout", func(t *testing.T) {

		_, _ = userService.Create(username, password, false)

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
			if err != ErrIncorrect {
				t.Errorf("Expected ErrIncorrect, got %v", err)
			}
		})

		t.Run("logout", func(t *testing.T) {

			_, err := userService.Login(username, password)

			err = userService.Logout(username)
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

	t.Run("create default admin user", func(t *testing.T) {
		result, err := userService.CreateDefaultAdminUser()
		if err != nil {
			t.Errorf("Failed to create: %v", err)
			return
		}

		expectedResult := "adminadmin"
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
}
