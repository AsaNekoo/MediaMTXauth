// File: test/storage_test.go
package test

import (
	"MediaMTXAuth/internal/storage"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

// setupTestDB creates a temporary database for testing
func setupTestDB(t *testing.T) (*storage.Storage, func()) {
	tempDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := storage.InitDB(dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to init test DB: %v", err)
	}

	cleanup := func() {
		storage.Close()
		os.RemoveAll(tempDir)
	}

	return storage, cleanup
}

// TestInitDB tests database initialization
func TestInitDB(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("successful initialization", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "test_success.db")

		s, err := storage.InitDB(dbPath)
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}
		defer s.Close()

		// Check if database file was created
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("Database file was not created")
		}

		// Verify buckets were created
		err = s.Db.View(func(tx *bolt.Tx) error {
			buckets := []string{"users", "namespaces", "user_dash_sessions", "stream_sessions"}

			for _, bucketName := range buckets {
				bucket := tx.Bucket([]byte(bucketName))
				if bucket == nil {
					t.Errorf("Bucket %s was not created", bucketName)
				}
			}
			return nil
		})

		if err != nil {
			t.Errorf("Error checking buckets: %v", err)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		// Try to create database in non-existent directory
		invalidPath := "/nonexistent/directory/test.db"

		s, err := storage.InitDB(invalidPath)
		if err == nil {
			if s != nil {
				s.Close()
			}
			t.Error("Expected error for invalid path, but got none")
		}
	})

	t.Run("database timeout", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "timeout_test.db")

		// Create and hold a lock on the database
		db1, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 100 * time.Millisecond})
		if err != nil {
			t.Fatalf("Failed to create first DB connection: %v", err)
		}
		defer db1.Close()

		// Try to open the same database - should timeout
		s, err := storage.InitDB(dbPath)
		if err == nil {
			if s != nil {
				s.Close()
			}
			t.Error("Expected timeout error but got none")
		}
	})
}

// TestClose tests database closing
func TestClose(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	err := s.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Try to use closed database - should fail
	err = s.CreateUser("test", "test")
	if err == nil {
		t.Error("Expected error when using closed database")
	}
}

// TestCreateUser tests user creation functionality
func TestCreateUser(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	testCases := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "valid user creation",
			username: "testuser",
			password: "testpassword123",
			wantErr:  false,
		},
		{
			name:     "short username - 1 char",
			username: "a",
			password: "validpassword123",
			wantErr:  true, // Should reject username < 3 chars
		},
		{
			name:     "short username - 2 chars",
			username: "ab",
			password: "validpassword123",
			wantErr:  true, // Should reject username < 3 chars
		},
		{
			name:     "minimum valid username - 3 chars",
			username: "abc",
			password: "validpassword123",
			wantErr:  false, // Should accept username = 3 chars
		},
		{
			name:     "empty username",
			username: "",
			password: "validpassword123",
			wantErr:  true, // Should reject empty username
		},
		{
			name:     "whitespace only username",
			username: "   ",
			password: "validpassword123",
			wantErr:  true, // Should reject whitespace-only username
		},
		{
			name:     "short password - 1 char",
			username: "validuser",
			password: "a",
			wantErr:  true, // Should reject password < 8 chars
		},
		{
			name:     "short password - 7 chars",
			username: "validuser",
			password: "1234567",
			wantErr:  true, // Should reject password < 8 chars
		},
		{
			name:     "minimum valid password - 8 chars",
			username: "validuser",
			password: "12345678",
			wantErr:  false, // Should accept password = 8 chars
		},
		{
			name:     "empty password",
			username: "validuser",
			password: "",
			wantErr:  true, // Should reject empty password
		},
		{
			name:     "unicode username",
			username: "пользователь",
			password: "пароль123",
			wantErr:  false,
		},
		{
			name:     "special characters",
			username: "user@domain.com",
			password: "p@ssw0rd!#$%",
			wantErr:  false,
		},
		{
			name:     "long username",
			username: string(make([]byte, 1000)), // Very long username
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "long password",
			username: "longpassuser",
			password: string(make([]byte, 10000)), // Very long password
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.CreateUser(tc.username, tc.password)

			if tc.wantErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If creation succeeded, verify user exists in database
			if !tc.wantErr && err == nil {
				var stored []byte
				err = s.Db.View(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte("users"))
					stored = b.Get([]byte(tc.username))
					return nil
				})

				if err != nil {
					t.Errorf("Error checking stored user: %v", err)
				}

				if stored == nil {
					t.Error("User was not stored in database")
				}

				// Verify it's a proper hash (starts with $argon2id$)
				if len(stored) < 10 || string(stored[:9]) != "$argon2id" {
					t.Error("Stored value is not a proper Argon2 hash")
				}
			}
		})
	}
}

// TestCreateUserDuplicate tests creating duplicate users
func TestCreateUserDuplicate(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	username := "duplicateuser"
	password1 := "password1"
	password2 := "password2"

	// Create first user
	err := s.CreateUser(username, password1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Create user with same username but different password
	err = s.CreateUser(username, password2)
	if err != nil {
		t.Fatalf("Failed to create duplicate user: %v", err)
	}

	// Verify the password was updated (second password should work)
	verified, err := s.VerifyUser(username, password2)
	if err != nil {
		t.Fatalf("Failed to verify updated user: %v", err)
	}
	if !verified {
		t.Error("Updated password was not verified")
	}

	// Verify old password no longer works
	verified, err = s.VerifyUser(username, password1)
	if err != nil {
		t.Fatalf("Failed to verify old password: %v", err)
	}
	if verified {
		t.Error("Old password still works after update")
	}
}

// TestVerifyUser tests user verification functionality
func TestVerifyUser(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test users (ensure all meet validation requirements)
	testUsers := map[string]string{
		"user1":         "password123",
		"user2":         "different_password",
		"unicode_user":  "юникод_пароль",
		"special_chars": "p@ssw0rd!#$%",
		"min_length":    "12345678", // minimum 8-char password
	}

	for username, password := range testUsers {
		err := s.CreateUser(username, password)
		if err != nil {
			t.Fatalf("Failed to create test user %s: %v", username, err)
		}
	}

	t.Run("successful verification", func(t *testing.T) {
		for username, password := range testUsers {
			verified, err := s.VerifyUser(username, password)
			if err != nil {
				t.Errorf("Verification failed for user %s: %v", username, err)
				continue
			}
			if !verified {
				t.Errorf("Password verification failed for user %s", username)
			}
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		verified, err := s.VerifyUser("user1", "wrongpassword")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if verified {
			t.Error("Wrong password was incorrectly verified")
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		verified, err := s.VerifyUser("nonexistent", "anypassword")
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
		if verified {
			t.Error("Non-existent user was incorrectly verified")
		}

		// Check error message
		expectedMsg := "user not found"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		// Test short username
		verified, err := s.VerifyUser("ab", "validpassword123")
		if err == nil {
			t.Error("Expected error for short username, but got none")
		}
		if verified {
			t.Error("Short username was incorrectly verified")
		}

		// Test empty username
		verified, err = s.VerifyUser("", "validpassword123")
		if err == nil {
			t.Error("Expected error for empty username, but got none")
		}
		if verified {
			t.Error("Empty username was incorrectly verified")
		}

		// Test whitespace-only username
		verified, err = s.VerifyUser("   ", "validpassword123")
		if err == nil {
			t.Error("Expected error for whitespace-only username, but got none")
		}
		if verified {
			t.Error("Whitespace-only username was incorrectly verified")
		}

		// Test short password
		verified, err = s.VerifyUser("validuser", "short")
		if err == nil {
			t.Error("Expected error for short password, but got none")
		}
		if verified {
			t.Error("Short password was incorrectly verified")
		}

		// Test empty password
		verified, err = s.VerifyUser("validuser", "")
		if err == nil {
			t.Error("Expected error for empty password, but got none")
		}
		if verified {
			t.Error("Empty password was incorrectly verified")
		}
	})
}

// TestStorageIntegration tests the complete user lifecycle
func TestStorageIntegration(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	testCases := []struct {
		username    string
		password    string
		newPassword string
	}{
		{"integration_user1", "original_password123", "new_password456"},
		{"integration_user2", "simple12345", "complex!@#$%678"},
		{"интеграция_юзер", "старый_пароль123", "новый_пароль456"},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			// Step 1: Create user
			err := s.CreateUser(tc.username, tc.password)
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}

			// Step 2: Verify original password works
			verified, err := s.VerifyUser(tc.username, tc.password)
			if err != nil {
				t.Fatalf("Failed to verify original password: %v", err)
			}
			if !verified {
				t.Error("Original password verification failed")
			}

			// Step 3: Update password (create user with same name)
			err = s.CreateUser(tc.username, tc.newPassword)
			if err != nil {
				t.Fatalf("Failed to update password: %v", err)
			}

			// Step 4: Verify new password works
			verified, err = s.VerifyUser(tc.username, tc.newPassword)
			if err != nil {
				t.Fatalf("Failed to verify new password: %v", err)
			}
			if !verified {
				t.Error("New password verification failed")
			}

			// Step 5: Verify old password no longer works
			verified, err = s.VerifyUser(tc.username, tc.password)
			if err != nil {
				t.Fatalf("Unexpected error verifying old password: %v", err)
			}
			if verified {
				t.Error("Old password still works after update")
			}

			// Step 6: Verify wrong password fails
			verified, err = s.VerifyUser(tc.username, "completely_wrong")
			if err != nil {
				t.Fatalf("Unexpected error verifying wrong password: %v", err)
			}
			if verified {
				t.Error("Wrong password was incorrectly verified")
			}
		})
	}
}

// TestMinimumValidLengths tests minimum valid username and password lengths
func TestMinimumValidLengths(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("minimum valid lengths", func(t *testing.T) {
		// Test minimum valid username (3 chars) and password (8 chars)
		username := "abc"      // exactly 3 characters
		password := "12345678" // exactly 8 characters

		// Should create successfully
		err := s.CreateUser(username, password)
		if err != nil {
			t.Errorf("Failed to create user with minimum valid lengths: %v", err)
		}

		// Should verify successfully
		verified, err := s.VerifyUser(username, password)
		if err != nil {
			t.Errorf("Failed to verify user with minimum valid lengths: %v", err)
		}
		if !verified {
			t.Error("Minimum valid lengths were not verified")
		}
	})

	t.Run("username with spaces", func(t *testing.T) {
		// Test username with leading/trailing spaces that becomes valid after trim
		username := " abc " // becomes "abc" after trim - should be valid
		password := "validpassword123"

		err := s.CreateUser(username, password)
		if err != nil {
			t.Errorf("Failed to create user with spaces in username: %v", err)
		}

		// Verify with trimmed username
		verified, err := s.VerifyUser("abc", password)
		if err != nil {
			t.Errorf("Failed to verify user with trimmed username: %v", err)
		}
		if !verified {
			t.Error("User with trimmed username was not verified")
		}
	})
}

func TestValidationErrors(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	// Import storage package to access error variables
	// Note: You might need to adjust the import path
	// import "MediaMTXAuth/internal/storage"

	testCases := []struct {
		name        string
		username    string
		password    string
		expectedErr string
		operation   string
	}{
		{
			name:        "create user - empty username",
			username:    "",
			password:    "validpassword123",
			expectedErr: "username must be at least 3 characters long",
			operation:   "create",
		},
		{
			name:        "create user - whitespace username",
			username:    "   \t\n  ",
			password:    "validpassword123",
			expectedErr: "username must be at least 3 characters long",
			operation:   "create",
		},
		{
			name:        "create user - short username",
			username:    "ab",
			password:    "validpassword123",
			expectedErr: "username must be at least 3 characters long",
			operation:   "create",
		},
		{
			name:        "create user - empty password",
			username:    "validuser",
			password:    "",
			expectedErr: "password must be at least 8 characters long",
			operation:   "create",
		},
		{
			name:        "create user - short password",
			username:    "validuser",
			password:    "short",
			expectedErr: "password must be at least 8 characters long",
			operation:   "create",
		},
		{
			name:        "verify user - empty username",
			username:    "",
			password:    "validpassword123",
			expectedErr: "username must be at least 3 characters long",
			operation:   "verify",
		},
		{
			name:        "verify user - short username",
			username:    "xy",
			password:    "validpassword123",
			expectedErr: "username must be at least 3 characters long",
			operation:   "verify",
		},
		{
			name:        "verify user - empty password",
			username:    "validuser",
			password:    "",
			expectedErr: "password must be at least 8 characters long",
			operation:   "verify",
		},
		{
			name:        "verify user - short password",
			username:    "validuser",
			password:    "1234567",
			expectedErr: "password must be at least 8 characters long",
			operation:   "verify",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			if tc.operation == "create" {
				err = s.CreateUser(tc.username, tc.password)
			} else {
				_, err = s.VerifyUser(tc.username, tc.password)
			}

			if err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if err.Error() != tc.expectedErr {
				t.Errorf("Expected error '%s', got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	s, cleanup := setupTestDB(t)
	defer cleanup()

	const numGoroutines = 10
	const usersPerGoroutine = 10

	// Channel to collect errors
	errChan := make(chan error, numGoroutines*usersPerGoroutine*2) // *2 for create and verify

	// Start concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < usersPerGoroutine; j++ {
				username := fmt.Sprintf("concurrent_user_%d_%d", routineID, j)
				password := fmt.Sprintf("password_%d_%d_123", routineID, j) // Ensure 8+ chars

				// Create user
				if err := s.CreateUser(username, password); err != nil {
					errChan <- fmt.Errorf("goroutine %d: create user %s failed: %v", routineID, username, err)
					continue
				}

				// Verify user
				verified, err := s.VerifyUser(username, password)
				if err != nil {
					errChan <- fmt.Errorf("goroutine %d: verify user %s failed: %v", routineID, username, err)
					continue
				}
				if !verified {
					errChan <- fmt.Errorf("goroutine %d: user %s verification returned false", routineID, username)
				}
			}
		}(i)
	}

	// Wait a bit for goroutines to complete
	time.Sleep(2 * time.Second)

	// Check for errors
	close(errChan)
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Errorf("Concurrent access failed with %d errors:", len(errors))
		for _, err := range errors {
			t.Error(err)
		}
	}
}

// BenchmarkCreateUser benchmarks user creation
func BenchmarkCreateUser(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "storage_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "bench.db")
	s, err := storage.InitDB(dbPath)
	if err != nil {
		b.Fatalf("Failed to init benchmark DB: %v", err)
	}
	defer s.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		username := fmt.Sprintf("bench_user_%d", i)
		password := "benchmark_password_123" // Ensure 8+ chars

		err := s.CreateUser(username, password)
		if err != nil {
			b.Fatalf("CreateUser failed: %v", err)
		}
	}
}

// BenchmarkVerifyUser benchmarks user verification
func BenchmarkVerifyUser(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "storage_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "bench.db")
	s, err := storage.InitDB(dbPath)
	if err != nil {
		b.Fatalf("Failed to init benchmark DB: %v", err)
	}
	defer s.Close()

	// Setup: create test user
	username := "bench_verify_user"
	password := "benchmark_password_123"
	err = s.CreateUser(username, password)
	if err != nil {
		b.Fatalf("Failed to create benchmark user: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		verified, err := s.VerifyUser(username, password)
		if err != nil {
			b.Fatalf("VerifyUser failed: %v", err)
		}
		if !verified {
			b.Fatal("VerifyUser returned false")
		}
	}
}
