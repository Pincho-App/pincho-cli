package crypto

import (
	"encoding/hex"
	"testing"
)

func TestCustomBase64Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "simple input",
			input:    []byte("hello"),
			expected: "aGVsbG8_",
		},
		{
			name:     "with special chars",
			input:    []byte("\x00\xFF\xAB"),
			expected: "AP-r",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CustomBase64Encode(tt.input)
			if result != tt.expected {
				t.Errorf("CustomBase64Encode(%v) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDeriveEncryptionKey(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantLen  int
	}{
		{
			name:     "simple password",
			password: "password123",
			wantLen:  16,
		},
		{
			name:     "complex password",
			password: "c0mpl3x-P@ssw0rd!#$",
			wantLen:  16,
		},
		{
			name:     "empty password",
			password: "",
			wantLen:  16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := DeriveEncryptionKey(tt.password)
			if err != nil {
				t.Fatalf("DeriveEncryptionKey() error = %v", err)
			}
			if len(key) != tt.wantLen {
				t.Errorf("DeriveEncryptionKey() key length = %d, want %d", len(key), tt.wantLen)
			}
		})
	}
}

func TestDeriveEncryptionKey_Consistency(t *testing.T) {
	password := "test-password"

	key1, err := DeriveEncryptionKey(password)
	if err != nil {
		t.Fatalf("DeriveEncryptionKey() error = %v", err)
	}

	key2, err := DeriveEncryptionKey(password)
	if err != nil {
		t.Fatalf("DeriveEncryptionKey() error = %v", err)
	}

	if hex.EncodeToString(key1) != hex.EncodeToString(key2) {
		t.Error("DeriveEncryptionKey() should produce consistent keys for same password")
	}
}

func TestGenerateIV(t *testing.T) {
	// Generate multiple IVs
	iv1, iv1Hex, err := GenerateIV()
	if err != nil {
		t.Fatalf("GenerateIV() error = %v", err)
	}

	_, iv2Hex, err := GenerateIV()
	if err != nil {
		t.Fatalf("GenerateIV() error = %v", err)
	}

	// Verify IV length
	if len(iv1) != 16 {
		t.Errorf("GenerateIV() IV length = %d, want 16", len(iv1))
	}

	// Verify hex string length (16 bytes = 32 hex chars)
	if len(iv1Hex) != 32 {
		t.Errorf("GenerateIV() IV hex length = %d, want 32", len(iv1Hex))
	}

	// Verify IVs are different (randomness check)
	if iv1Hex == iv2Hex {
		t.Error("GenerateIV() should generate unique IVs")
	}

	// Verify hex encoding is correct
	decodedIV, err := hex.DecodeString(iv1Hex)
	if err != nil {
		t.Fatalf("Invalid hex encoding in IV: %v", err)
	}
	if hex.EncodeToString(decodedIV) != iv1Hex {
		t.Error("IV hex encoding/decoding mismatch")
	}
}

func TestEncryptMessage(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		password  string
		ivHex     string
		wantErr   bool
	}{
		{
			name:      "simple message",
			plaintext: "Hello, World!",
			password:  "test-password",
			ivHex:     "0123456789abcdef0123456789abcdef",
			wantErr:   false,
		},
		{
			name:      "empty message",
			plaintext: "",
			password:  "test-password",
			ivHex:     "0123456789abcdef0123456789abcdef",
			wantErr:   false,
		},
		{
			name:      "long message",
			plaintext: "This is a longer message that will definitely require padding for the AES block cipher.",
			password:  "test-password",
			ivHex:     "0123456789abcdef0123456789abcdef",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iv, err := hex.DecodeString(tt.ivHex)
			if err != nil && !tt.wantErr {
				t.Fatalf("Invalid test IV: %v", err)
			}

			encrypted, err := EncryptMessage(tt.plaintext, tt.password, iv)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify encryption produced output
				if len(encrypted) == 0 {
					t.Error("EncryptMessage() produced empty ciphertext")
				}

				// Verify it doesn't contain plaintext
				if len(tt.plaintext) > 0 && encrypted == tt.plaintext {
					t.Error("EncryptMessage() returned plaintext instead of ciphertext")
				}

				// Verify consistent encryption with same IV
				encrypted2, err := EncryptMessage(tt.plaintext, tt.password, iv)
				if err != nil {
					t.Fatalf("EncryptMessage() second call error = %v", err)
				}
				if encrypted != encrypted2 {
					t.Error("EncryptMessage() should produce consistent results with same IV")
				}
			}
		})
	}
}

func TestEncryptMessage_InterSDKCompatibility(t *testing.T) {
	// Test with known values from other SDKs
	plaintext := "This is a secret message that needs to be encrypted securely."
	password := "test_password_123"
	ivHex := "0123456789abcdef0123456789abcdef"

	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		t.Fatalf("Invalid test IV: %v", err)
	}

	encrypted, err := EncryptMessage(plaintext, password, iv)
	if err != nil {
		t.Fatalf("EncryptMessage() error = %v", err)
	}

	// This is the expected output from Python, JavaScript, Go, and Java SDKs
	expected := "y2fzGqnZSgdMqkwYhAUEZi30VFBYvwcCmrQ6BmSliPpPGHXMdMRsLCtG-cfwhhxN4HSIk5Y3UMjM6XoBWPqiHw__"

	if encrypted != expected {
		t.Errorf("EncryptMessage() = %s, want %s (inter-SDK compatibility failed)", encrypted, expected)
	}
}
