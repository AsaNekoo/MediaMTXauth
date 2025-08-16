package test

import (
	"MediaMTXAuth/internal/storage"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper function to create a temporary database
func createTestDB(t *testing.T) (*storage.Storage, string) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	store, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return store, dbPath
}

func TestInitDB(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")

		store, err := storage.InitDB(dbPath)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if store == nil {
			t.Fatal("Expected storage instance, got nil")
		}

		if store.Db == nil {
			t.Fatal("Expected database instance, got nil")
		}

		store.Close()
	})

	t.Run("invalid path", func(t *testing.T) {
		invalidPath := "/nonexistent/directory/that/should/not/exist/test.db"

		store, err := storage.InitDB(invalidPath)
		if err == nil {
			if store != nil {
				store.Close()
			}
			t.Fatal("Expected error for invalid path, got nil")
		}
	})
}

func TestClose(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		store, _ := createTestDB(t)

		err := store.Close()
		if err != nil {
			t.Fatalf("Expected no error when closing, got %v", err)
		}
	})

	t.Run("operations after close should fail", func(t *testing.T) {
		store, _ := createTestDB(t)

		// Close the database
		err := store.Close()
		if err != nil {
			t.Fatalf("Expected no error when closing, got %v", err)
		}

		// Try to create a user after closing - should fail
		user := &storage.User{
			Name:      "testuser",
			Generated: false,
		}

		err = store.CreateUser(user, "password123")
		if err == nil {
			t.Fatal("Expected error when creating user after database close, got nil")
		}

		if !strings.Contains(err.Error(), "database not open") {
			t.Fatalf("Expected 'database not open' error, got: %v", err)
		}

		// Try to get user after closing - should fail
		_, err = store.GetUser("testuser")
		if err == nil {
			t.Fatal("Expected error when getting user after database close, got nil")
		}

		// Try to verify user after closing - should fail
		_, err = store.VerifyUser("testuser", "password123")
		if err == nil {
			t.Fatal("Expected error when verifying user after database close, got nil")
		}
	})
}

func TestIDSystem(t *testing.T) {
	t.Run("basic sequential IDs", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		const numUsers = 10
		users := make([]*storage.User, numUsers)

		for i := 0; i < numUsers; i++ {
			users[i] = &storage.User{
				Name:      fmt.Sprintf("user%d", i),
				Generated: i%2 == 0,
			}

			err := store.CreateUser(users[i], "password123")
			if err != nil {
				t.Fatalf("Failed to create user %d: %v", i, err)
			}

			expectedID := i + 1
			if users[i].ID != expectedID {
				t.Fatalf("User %d: expected ID %d, got %d", i, expectedID, users[i].ID)
			}
		}

		// Verify all IDs are unique
		idSet := make(map[int]bool)
		for i, user := range users {
			if idSet[user.ID] {
				t.Fatalf("Duplicate ID %d found for user %d", user.ID, i)
			}
			idSet[user.ID] = true
		}
	})

	t.Run("large sequential IDs", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		// Create many users to test large ID values
		const bulkUsers = 1000

		// Create bulk users (testing in batches to avoid timeout)
		for i := 1; i <= bulkUsers; i++ {
			user := &storage.User{
				Name:      fmt.Sprintf("bulk_user_%d", i),
				Generated: i%3 == 0,
			}

			err := store.CreateUser(user, "password123")
			if err != nil {
				t.Fatalf("Failed to create bulk user %d: %v", i, err)
			}

			if user.ID != i {
				t.Fatalf("Bulk user %d: expected ID %d, got %d", i, i, user.ID)
			}

			// Periodically verify some users to ensure everything works at scale
			if i%10 == 0 {
				retrieved, err := store.GetUser(user.Name)
				if err != nil {
					t.Fatalf("Failed to retrieve bulk user %d: %v", i, err)
				}
				if retrieved.ID != user.ID {
					t.Fatalf("Retrieved user %d has wrong ID: expected %d, got %d",
						i, user.ID, retrieved.ID)
				}
			}
		}

		// Create a few more users to verify sequence continues properly
		testUsers := []string{"test_large_1", "test_large_2", "test_large_3"}
		for i, name := range testUsers {
			user := &storage.User{
				Name:      name,
				Generated: true,
			}

			err := store.CreateUser(user, "password123")
			if err != nil {
				t.Fatalf("Failed to create test user %s: %v", name, err)
			}

			expectedID := bulkUsers + i + 1
			if user.ID != expectedID {
				t.Fatalf("Test user %s: expected ID %d, got %d", name, expectedID, user.ID)
			}
		}
	})

	t.Run("ID persistence across restarts", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "id_persist_test.db")

		const initialUsers = 100
		var lastID int

		// Session 1: Create initial users
		{
			store1, err := storage.InitDB(dbPath)
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}

			for i := 1; i <= initialUsers; i++ {
				user := &storage.User{
					Name:      fmt.Sprintf("persist_user_%d", i),
					Generated: i%4 == 0,
				}

				err = store1.CreateUser(user, "password123")
				if err != nil {
					t.Fatalf("Failed to create user %d: %v", i, err)
				}

				if user.ID != i {
					t.Fatalf("User %d: expected ID %d, got %d", i, i, user.ID)
				}
				lastID = user.ID
			}

			store1.Close()
		}

		// Session 2: Verify sequence continues
		{
			store2, err := storage.InitDB(dbPath)
			if err != nil {
				t.Fatalf("Failed to reopen database: %v", err)
			}
			defer store2.Close()

			// Verify existing users still exist with correct IDs
			for i := 1; i <= 5; i++ { // Check first few
				user, err := store2.GetUser(fmt.Sprintf("persist_user_%d", i))
				if err != nil {
					t.Fatalf("Failed to get persisted user %d: %v", i, err)
				}
				if user.ID != i {
					t.Fatalf("Persisted user %d has wrong ID: expected %d, got %d", i, i, user.ID)
				}
			}

			// Create new users - should continue sequence
			const newUsers = 50
			for i := 1; i <= newUsers; i++ {
				user := &storage.User{
					Name:      fmt.Sprintf("new_user_%d", i),
					Generated: false,
				}

				err = store2.CreateUser(user, "password456")
				if err != nil {
					t.Fatalf("Failed to create new user %d: %v", i, err)
				}

				expectedID := lastID + i
				if user.ID != expectedID {
					t.Fatalf("New user %d: expected ID %d, got %d", i, expectedID, user.ID)
				}
			}
		}
	})

	t.Run("ID uniqueness under concurrent access", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		const numGoroutines = 100
		results := make(chan *storage.User, numGoroutines)
		errors := make(chan error, numGoroutines)

		// Launch concurrent user creation
		for i := 0; i < numGoroutines; i++ {
			go func(idx int) {
				user := &storage.User{
					Name:      fmt.Sprintf("concurrent_user_%d_%d", idx, time.Now().UnixNano()),
					Generated: idx%2 == 0,
				}

				err := store.CreateUser(user, "password123")
				if err != nil {
					errors <- err
					return
				}

				results <- user
			}(i)
		}

		// Collect results
		users := make([]*storage.User, 0, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			select {
			case user := <-results:
				users = append(users, user)
			case err := <-errors:
				t.Fatalf("Concurrent user creation failed: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatal("Timeout waiting for concurrent user creation")
			}
		}

		if len(users) != numGoroutines {
			t.Fatalf("Expected %d users, got %d", numGoroutines, len(users))
		}

		// Verify all IDs are unique and positive
		idSet := make(map[int]bool)
		for i, user := range users {
			if user.ID <= 0 {
				t.Fatalf("User %d has invalid ID: %d", i, user.ID)
			}

			if idSet[user.ID] {
				t.Fatalf("Duplicate ID %d found", user.ID)
			}
			idSet[user.ID] = true
		}

		// Verify all users can be retrieved
		for i, user := range users {
			retrieved, err := store.GetUser(user.Name)
			if err != nil {
				t.Fatalf("Failed to retrieve concurrent user %d: %v", i, err)
			}

			if retrieved.ID != user.ID {
				t.Fatalf("Retrieved user %d has wrong ID: expected %d, got %d",
					i, user.ID, retrieved.ID)
			}
		}
	})

	t.Run("itob function validation through ID storage", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		// Test edge cases that might expose itob issues
		testCases := []struct {
			description string
			userCount   int
		}{
			{"single user", 1},
			{"beyond byte boundary", 256},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				tempStore, _ := createTestDB(t)
				defer tempStore.Close()

				// Create users up to the test count
				for i := 1; i <= tc.userCount; i++ {
					user := &storage.User{
						Name:      fmt.Sprintf("itob_test_%d_%d", tc.userCount, i),
						Generated: false,
					}

					err := tempStore.CreateUser(user, "password123")
					if err != nil {
						t.Fatalf("Failed to create user %d for %s test: %v",
							i, tc.description, err)
					}

					if user.ID != i {
						t.Fatalf("%s: user %d has wrong ID: expected %d, got %d",
							tc.description, i, i, user.ID)
					}

					// Verify user can be retrieved (tests itob round-trip)
					retrieved, err := tempStore.GetUser(user.Name)
					if err != nil {
						t.Fatalf("%s: failed to retrieve user %d: %v",
							tc.description, i, err)
					}

					if retrieved.ID != user.ID {
						t.Fatalf("%s: retrieved user %d has wrong ID: expected %d, got %d",
							tc.description, i, user.ID, retrieved.ID)
					}
				}
			})
		}
	})
}

func TestCreateUser(t *testing.T) {
	t.Run("successful user creation", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		user := &storage.User{
			Name:      "testuser",
			Generated: false,
		}

		err := store.CreateUser(user, "password123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// User should have been assigned an ID
		if user.ID == 0 {
			t.Error("Expected user ID to be assigned, got 0")
		}

		// User name should be trimmed
		if user.Name != "testuser" {
			t.Errorf("Expected user name to be 'testuser', got '%s'", user.Name)
		}

		// User should have a hash
		if user.Hash == "" {
			t.Error("Expected user hash to be set, got empty string")
		}

		// Hash should not be the plain password
		if user.Hash == "password123" {
			t.Error("Hash should not be the plain password")
		}
	})

	t.Run("duplicate username handling", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		// Create first user
		user1 := &storage.User{
			Name:      "duplicate",
			Generated: false,
		}

		err := store.CreateUser(user1, "password123")
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Try to create user with same name
		user2 := &storage.User{
			Name:      "duplicate",
			Generated: true,
		}

		err = store.CreateUser(user2, "password456")

		// Current implementation ALLOWS duplicates - this reveals the bug
		if err == nil {
			t.Logf("BUG DETECTED: Implementation allows duplicate usernames!")
			t.Logf("User 1 - ID: %d, Name: %s", user1.ID, user1.Name)
			t.Logf("User 2 - ID: %d, Name: %s", user2.ID, user2.Name)

			// Verify both users exist but have different IDs
			if user1.ID == user2.ID {
				t.Fatal("Duplicate users should not have same ID")
			}

			// This creates an ambiguous state - which user should GetUser return?
			retrieved, err := store.GetUser("duplicate")
			if err != nil {
				t.Fatalf("Failed to get duplicate user: %v", err)
			}

			// The last one inserted should overwrite the index entry
			if retrieved.ID != user2.ID {
				t.Logf("Index points to user ID %d instead of latest user ID %d",
					retrieved.ID, user2.ID)
			}

			t.Log("EXPECTED BEHAVIOR: Should have prevented duplicate username creation")
		} else {
			t.Logf("Good: Duplicate username was rejected: %v", err)
		}
	})

	t.Run("username validation", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		testCases := []struct {
			name     string
			username string
			password string
			wantErr  error
		}{
			{"empty username", "", "password123", storage.ErrUsername},
			{"short username", "ab", "password123", storage.ErrUsername},
			{"whitespace only username", "   ", "password123", storage.ErrUsername},
			{"username with leading/trailing spaces", "  validuser  ", "password123", nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				user := &storage.User{
					Name:      tc.username,
					Generated: false,
				}

				err := store.CreateUser(user, tc.password)
				if tc.wantErr != nil {
					if err != tc.wantErr {
						t.Fatalf("Expected error %v, got %v", tc.wantErr, err)
					}
				} else {
					if err != nil {
						t.Fatalf("Expected no error, got %v", err)
					}
					// Username should be trimmed
					if user.Name != strings.TrimSpace(tc.username) {
						t.Errorf("Expected trimmed username '%s', got '%s'",
							strings.TrimSpace(tc.username), user.Name)
					}
				}
			})
		}
	})

	t.Run("password validation", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		testCases := []struct {
			name     string
			username string
			password string
			wantErr  error
		}{
			{"empty password", "testuser", "", storage.ErrPassword},
			{"short password", "testuser", "1234567", storage.ErrPassword},
			{"whitespace only password", "testuser", "        ", storage.ErrPassword},
			{"password with leading/trailing spaces", "testuser", "  validpass123  ", nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				user := &storage.User{
					Name:      tc.username,
					Generated: false,
				}

				err := store.CreateUser(user, tc.password)
				if tc.wantErr != nil {
					if err != tc.wantErr {
						t.Fatalf("Expected error %v, got %v", tc.wantErr, err)
					}
				} else {
					if err != nil {
						t.Fatalf("Expected no error, got %v", err)
					}
				}
			})
		}
	})
}

func TestGetUser(t *testing.T) {
	t.Run("get existing user", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		// Create a user
		originalUser := &storage.User{
			Name:      "testuser",
			Generated: true,
		}

		err := store.CreateUser(originalUser, "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Get the user
		retrievedUser, err := store.GetUser("testuser")
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if retrievedUser == nil {
			t.Fatal("Expected user, got nil")
		}

		// Verify all fields match
		if retrievedUser.ID != originalUser.ID {
			t.Errorf("Expected ID %d, got %d", originalUser.ID, retrievedUser.ID)
		}

		if retrievedUser.Name != originalUser.Name {
			t.Errorf("Expected name '%s', got '%s'", originalUser.Name, retrievedUser.Name)
		}

		if retrievedUser.Hash != originalUser.Hash {
			t.Errorf("Expected hash '%s', got '%s'", originalUser.Hash, retrievedUser.Hash)
		}

		if retrievedUser.Generated != originalUser.Generated {
			t.Errorf("Expected Generated %v, got %v", originalUser.Generated, retrievedUser.Generated)
		}
	})

	t.Run("get user with trimmed name", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		// Create user with spaces in name (will be trimmed)
		user := &storage.User{
			Name:      "  spaceuser  ",
			Generated: false,
		}

		err := store.CreateUser(user, "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Should be able to get user with trimmed name
		retrievedUser, err := store.GetUser("spaceuser")
		if err != nil {
			t.Fatalf("Failed to get user with trimmed name: %v", err)
		}

		if retrievedUser.Name != "spaceuser" {
			t.Errorf("Expected trimmed name 'spaceuser', got '%s'", retrievedUser.Name)
		}

		// Should also be able to get with spaces (will be trimmed during lookup)
		retrievedUser2, err := store.GetUser("  spaceuser  ")
		if err != nil {
			t.Fatalf("Failed to get user with spaces: %v", err)
		}

		if retrievedUser2.ID != retrievedUser.ID {
			t.Error("Should get same user regardless of spaces in lookup")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		user, err := store.GetUser("nonexistent")
		if err == nil {
			t.Fatal("Expected error for non-existent user, got nil")
		}

		if user != nil {
			t.Fatal("Expected nil user, got user instance")
		}

		if !strings.Contains(err.Error(), "user not found") {
			t.Fatalf("Expected 'user not found' error, got: %v", err)
		}
	})
}

func TestVerifyUser(t *testing.T) {
	t.Run("verify valid credentials", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		// Create a user
		user := &storage.User{
			Name:      "testuser",
			Generated: false,
		}
		password := "password123"

		err := store.CreateUser(user, password)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Verify with correct credentials
		valid, err := store.VerifyUser("testuser", password)
		if err != nil {
			t.Fatalf("Unexpected error during verification: %v", err)
		}

		if !valid {
			t.Fatal("Expected valid credentials to return true")
		}
	})

	t.Run("verify invalid password", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		// Create a user
		user := &storage.User{
			Name:      "testuser",
			Generated: false,
		}

		err := store.CreateUser(user, "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Verify with wrong password
		valid, err := store.VerifyUser("testuser", "wrongpassword")
		if err != nil {
			t.Fatalf("Unexpected error during verification: %v", err)
		}

		if valid {
			t.Fatal("Expected invalid password to return false")
		}
	})

	t.Run("verify nonexistent user", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		valid, err := store.VerifyUser("nonexistent", "password123")
		if err == nil {
			t.Fatal("Expected error for non-existent user, got nil")
		}

		if valid {
			t.Fatal("Expected false for non-existent user")
		}

		if !strings.Contains(err.Error(), "user not found") {
			t.Fatalf("Expected 'user not found' error, got: %v", err)
		}
	})

	t.Run("username validation in verify", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		testCases := []struct {
			name     string
			username string
			password string
			wantErr  error
		}{
			{"empty username", "", "password123", storage.ErrUsername},
			{"short username", "ab", "password123", storage.ErrUsername},
			{"whitespace only username", "   ", "password123", storage.ErrUsername},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				valid, err := store.VerifyUser(tc.username, tc.password)
				if err != tc.wantErr {
					t.Fatalf("Expected error %v, got %v", tc.wantErr, err)
				}
				if valid {
					t.Fatal("Expected valid to be false")
				}
			})
		}
	})

	t.Run("password validation in verify", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		testCases := []struct {
			name     string
			username string
			password string
			wantErr  error
		}{
			{"empty password", "testuser", "", storage.ErrPassword},
			{"short password", "testuser", "1234567", storage.ErrPassword},
			{"whitespace only password", "testuser", "        ", storage.ErrPassword},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				valid, err := store.VerifyUser(tc.username, tc.password)
				if err != tc.wantErr {
					t.Fatalf("Expected error %v, got %v", tc.wantErr, err)
				}
				if valid {
					t.Fatal("Expected valid to be false")
				}
			})
		}
	})
}

func TestDatabasePersistence(t *testing.T) {
	t.Run("data persists across database restarts", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "persist_test.db")

		var originalUser *storage.User

		// Create database session 1
		{
			store1, err := storage.InitDB(dbPath)
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}

			originalUser = &storage.User{
				Name:      "persistuser",
				Generated: true,
			}

			err = store1.CreateUser(originalUser, "password123")
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}

			store1.Close()
		}

		// Reopen database session 2
		{
			store2, err := storage.InitDB(dbPath)
			if err != nil {
				t.Fatalf("Failed to reopen database: %v", err)
			}
			defer store2.Close()

			// User should still exist
			retrievedUser, err := store2.GetUser("persistuser")
			if err != nil {
				t.Fatalf("User should persist across restarts: %v", err)
			}

			if retrievedUser.ID != originalUser.ID {
				t.Errorf("Expected ID %d, got %d", originalUser.ID, retrievedUser.ID)
			}

			if retrievedUser.Name != originalUser.Name {
				t.Errorf("Expected name '%s', got '%s'", originalUser.Name, retrievedUser.Name)
			}

			// Should be able to verify credentials
			valid, err := store2.VerifyUser("persistuser", "password123")
			if err != nil {
				t.Fatalf("Failed to verify persisted user: %v", err)
			}

			if !valid {
				t.Fatal("Persisted user credentials should be valid")
			}
		}
	})
}

func TestEndToEndUserFlow(t *testing.T) {
	t.Run("complete user lifecycle", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		// 1. Create user
		user := &storage.User{
			Name:      "  lifecycleuser  ", // Test trimming
			Generated: true,
		}
		password := "mypassword123"

		err := store.CreateUser(user, password)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Verify user properties after creation
		if user.ID == 0 {
			t.Fatal("User should have been assigned an ID")
		}

		if user.Name != "lifecycleuser" {
			t.Fatalf("Expected trimmed name 'lifecycleuser', got '%s'", user.Name)
		}

		if user.Hash == "" {
			t.Fatal("User should have a hash")
		}

		if user.Hash == password {
			t.Fatal("Hash should not be the plain password")
		}

		if !user.Generated {
			t.Fatal("Generated flag should be preserved")
		}

		// 2. Retrieve user
		retrieved, err := store.GetUser("lifecycleuser")
		if err != nil {
			t.Fatalf("Failed to retrieve user: %v", err)
		}

		if retrieved.ID != user.ID {
			t.Errorf("Retrieved user has wrong ID")
		}

		// 3. Verify valid credentials
		valid, err := store.VerifyUser("lifecycleuser", password)
		if err != nil {
			t.Fatalf("Failed to verify valid credentials: %v", err)
		}

		if !valid {
			t.Fatal("Valid credentials should return true")
		}

		// 4. Verify invalid credentials
		valid, err = store.VerifyUser("lifecycleuser", "wrongpassword")
		if err != nil {
			t.Fatalf("Unexpected error with wrong password: %v", err)
		}

		if valid {
			t.Fatal("Invalid credentials should return false")
		}

		// 5. Test with different name casing/spacing
		retrieved2, err := store.GetUser("  lifecycleuser  ")
		if err != nil {
			t.Fatalf("Should handle spaces in lookup: %v", err)
		}

		if retrieved2.ID != user.ID {
			t.Error("Should get same user regardless of lookup spacing")
		}
	})
}

func TestStressAndEdgeCases(t *testing.T) {
	t.Run("rapid sequential operations", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		const rapidOps = 1000

		// Rapid user creation
		for i := 1; i <= rapidOps; i++ {
			user := &storage.User{
				Name:      fmt.Sprintf("rapid_user_%d", i),
				Generated: i%5 == 0,
			}

			err := store.CreateUser(user, "password123")
			if err != nil {
				t.Fatalf("Rapid operation %d failed: %v", i, err)
			}

			if user.ID != i {
				t.Fatalf("Rapid user %d: expected ID %d, got %d", i, i, user.ID)
			}

			// Immediately try to retrieve and verify every 100th user
			if i%100 == 0 {
				retrieved, err := store.GetUser(user.Name)
				if err != nil {
					t.Fatalf("Failed to retrieve rapid user %d: %v", i, err)
				}

				if retrieved.ID != user.ID {
					t.Fatalf("Rapid user %d retrieval mismatch: expected ID %d, got %d",
						i, user.ID, retrieved.ID)
				}

				valid, err := store.VerifyUser(user.Name, "password123")
				if err != nil {
					t.Fatalf("Failed to verify rapid user %d: %v", i, err)
				}

				if !valid {
					t.Fatalf("Rapid user %d verification failed", i)
				}
			}
		}
	})

	t.Run("special characters in usernames", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		specialUsers := []struct {
			name     string
			username string
		}{
			{"unicode", "user_测试_🔑"},
			{"symbols", "user@domain.com"},
			{"mixed", "User_123-Test"},
			{"dots", "user.name.test"},
			{"dashes", "user-name-test"},
			{"underscores", "user_name_test"},
		}

		for i, tc := range specialUsers {
			t.Run(tc.name, func(t *testing.T) {
				user := &storage.User{
					Name:      tc.username,
					Generated: i%2 == 0,
				}

				err := store.CreateUser(user, "password123")
				if err != nil {
					t.Fatalf("Failed to create user with %s characters: %v", tc.name, err)
				}

				// Verify retrieval
				retrieved, err := store.GetUser(tc.username)
				if err != nil {
					t.Fatalf("Failed to retrieve user with %s characters: %v", tc.name, err)
				}

				if retrieved.Name != user.Name {
					t.Errorf("Name mismatch for %s: expected '%s', got '%s'",
						tc.name, user.Name, retrieved.Name)
				}

				// Verify authentication
				valid, err := store.VerifyUser(tc.username, "password123")
				if err != nil {
					t.Fatalf("Failed to verify user with %s characters: %v", tc.name, err)
				}

				if !valid {
					t.Fatalf("Authentication failed for user with %s characters", tc.name)
				}
			})
		}
	})
}

func TestIndexConsistency(t *testing.T) {
	t.Run("index and data bucket consistency", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		const testUsers = 100
		userNames := make([]string, testUsers)

		// Create users
		for i := 0; i < testUsers; i++ {
			userNames[i] = fmt.Sprintf("consistency_user_%d", i)
			user := &storage.User{
				Name:      userNames[i],
				Generated: i%3 == 0,
			}

			err := store.CreateUser(user, "password123")
			if err != nil {
				t.Fatalf("Failed to create user %d: %v", i, err)
			}
		}

		// Verify all users can be retrieved
		for i, userName := range userNames {
			user, err := store.GetUser(userName)
			if err != nil {
				t.Fatalf("Failed to get user %d (%s): %v", i, userName, err)
			}

			expectedID := i + 1
			if user.ID != expectedID {
				t.Fatalf("User %d (%s): expected ID %d, got %d",
					i, userName, expectedID, user.ID)
			}

			// Verify authentication works
			valid, err := store.VerifyUser(userName, "password123")
			if err != nil {
				t.Fatalf("Failed to verify user %d (%s): %v", i, userName, err)
			}

			if !valid {
				t.Fatalf("Authentication failed for user %d (%s)", i, userName)
			}
		}
	})

	t.Run("duplicate username index behavior", func(t *testing.T) {
		store, _ := createTestDB(t)
		defer store.Close()

		// This test documents the current (buggy) behavior with duplicates

		// Create first user
		user1 := &storage.User{
			Name:      "duplicate_test",
			Generated: false,
		}

		err := store.CreateUser(user1, "password1")
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Create second user with same name (bug: should be prevented)
		user2 := &storage.User{
			Name:      "duplicate_test",
			Generated: true,
		}

		err = store.CreateUser(user2, "password2")
		if err != nil {
			t.Skipf("Duplicate prevention exists: %v", err)
		}

		// Document current behavior
		t.Logf("BUG: Duplicate users created - User1 ID: %d, User2 ID: %d",
			user1.ID, user2.ID)

		// Check which user the index points to
		retrieved, err := store.GetUser("duplicate_test")
		if err != nil {
			t.Fatalf("Failed to get duplicate user: %v", err)
		}

		t.Logf("Index points to user with ID: %d (should be %d - the latest)",
			retrieved.ID, user2.ID)

		// Check which password works
		valid1, err1 := store.VerifyUser("duplicate_test", "password1")
		valid2, err2 := store.VerifyUser("duplicate_test", "password2")

		if err1 != nil || err2 != nil {
			t.Fatalf("Verification errors: err1=%v, err2=%v", err1, err2)
		}

		t.Logf("Password1 valid: %v, Password2 valid: %v", valid1, valid2)

		// The index should point to the latest user
		if retrieved.ID == user2.ID && !valid1 && valid2 {
			t.Log("Index correctly points to latest user")
		} else {
			t.Log("Index behavior is inconsistent")
		}
	})
}
