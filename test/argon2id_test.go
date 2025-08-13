package test

import (
	"MediaMTXAuth/internal/api"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "supersecret"

	// Hash the password
	hashed, err := api.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hashed == "" {
		t.Error("HashPassword returned empty string")
	}

	t.Logf("Hashed password: %s", hashed)

	// Verify correct password
	verified, err := api.VerifyPassword(password, hashed)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}

	if !verified {
		t.Error("VerifyPassword failed to verify correct password")
	}

	t.Logf("Password verification: %t", verified)

	// Test wrong password
	wrongPassword := "wrongpassword"
	verified, err = api.VerifyPassword(wrongPassword, hashed)
	if err != nil {
		t.Fatalf("VerifyPassword failed with wrong password: %v", err)
	}

	if verified {
		t.Error("VerifyPassword incorrectly verified wrong password")
	}

	t.Logf("Wrong password verification: %t", verified)
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "testpassword"

	// Generate multiple hashes
	hash1, err := api.HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword failed: %v", err)
	}

	hash2, err := api.HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}

	// They should be different due to different salts
	if hash1 == hash2 {
		t.Error("HashPassword produced identical hashes for same password")
	}

	// But both should verify the original password
	verified1, err := api.VerifyPassword(password, hash1)
	if err != nil || !verified1 {
		t.Error("First hash failed to verify original password")
	}

	verified2, err := api.VerifyPassword(password, hash2)
	if err != nil || !verified2 {
		t.Error("Second hash failed to verify original password")
	}

	t.Logf("Hash 1: %s", hash1)
	t.Logf("Hash 2: %s", hash2)
}

func TestVerifyPasswordErrorHandling(t *testing.T) {
	password := "testpassword"

	testCases := []struct {
		name         string
		hash         string
		expectError  bool
		expectResult bool
	}{
		{
			name:         "invalid format - too few parts",
			hash:         "$argon2id$v=19",
			expectError:  true,
			expectResult: false,
		},
		{
			name:         "invalid algorithm",
			hash:         "$bcrypt$v=19$m=19456,t=2,p=1$salt$hash",
			expectError:  true,
			expectResult: false,
		},
		{
			name:         "invalid base64 salt",
			hash:         "$argon2id$v=19$m=19456,t=2,p=1$invalid!!!$hash",
			expectError:  true,
			expectResult: false,
		},
		{
			name:         "empty string",
			hash:         "",
			expectError:  true,
			expectResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := api.VerifyPassword(password, tc.hash)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tc.expectResult {
				t.Errorf("Expected result %t, got %t", tc.expectResult, result)
			}
		})
	}
}

func TestMultiplePasswords(t *testing.T) {
	passwords := []string{
		"simple",
		"complex!@#$%^&*()",
		"unicode_тест_🔒",
		"",
		"   spaces   ",
		"very_long_password_that_is_much_longer_than_typical_passwords_to_test_edge_cases",
	}

	for _, password := range passwords {
		t.Run("password_"+password, func(t *testing.T) {
			// Hash password
			hash, err := api.HashPassword(password)
			if err != nil {
				t.Fatalf("HashPassword failed for '%s': %v", password, err)
			}

			// Verify correct password
			verified, err := api.VerifyPassword(password, hash)
			if err != nil {
				t.Fatalf("VerifyPassword failed for '%s': %v", password, err)
			}

			if !verified {
				t.Errorf("Password verification failed for '%s'", password)
			}

			// Verify wrong password (if not empty)
			if password != "" {
				wrongPassword := password + "wrong"
				verified, err = api.VerifyPassword(wrongPassword, hash)
				if err != nil {
					t.Fatalf("VerifyPassword with wrong password failed for '%s': %v", password, err)
				}

				if verified {
					t.Errorf("Wrong password incorrectly verified for '%s'", password)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkpassword123"

	for i := 0; i < b.N; i++ {
		_, err := api.HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "benchmarkpassword123"
	hash, err := api.HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := api.VerifyPassword(password, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}
