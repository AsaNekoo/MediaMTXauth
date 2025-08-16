package passwords

import (
	"testing"
)

const TestPassword = "hackme"

func TestHashAndVerifyPassword(t *testing.T) {
	hashed, err := Hash(TestPassword)

	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	t.Run("same passwords", func(t *testing.T) {
		ok, err := Verify(TestPassword, hashed)

		if err != nil {
			t.Errorf("Verify failed: %v", err)
		} else if !ok {
			t.Error("should be verified")
		}
	})

	t.Run("different passwords", func(t *testing.T) {
		ok, err := Verify("wrongpassword", hashed)

		if err != nil {
			t.Fatalf("Verify failed: %v", err)
		} else if ok {
			t.Error("different passwords should not be verifiable")
		}
	})
}
func TestPasswordHashUniqueness(t *testing.T) {
	password := "testpassword"

	hash1, err := Hash(password)

	if err != nil {
		t.Fatalf("First Hash failed: %v", err)
	}

	hash2, err := Hash(password)

	if err != nil {
		t.Fatalf("Second Hash failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Hash produced identical hashes for same passwords")
	}
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
			expectError:  false,
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
			result, err := Verify(password, tc.hash)

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
