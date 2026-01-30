// Package crypto provides message encryption utilities for Pincho notifications.
//
// The package implements AES-128-CBC encryption with custom base64 encoding to
// maintain compatibility with the Pincho mobile app's encryption scheme.
//
// Encryption process:
//  1. Derive 128-bit key from password using SHA1 hash (first 16 bytes)
//  2. Generate random 16-byte initialization vector (IV)
//  3. Encrypt message using AES-128-CBC with PKCS7 padding
//  4. Encode encrypted data using custom Base64 (URL-safe with custom chars)
//  5. Return encrypted message and IV (hex-encoded)
//
// The encryption scheme matches the Pincho iOS/Android app implementation
// for end-to-end encrypted notifications. The encrypted message is stored on
// the server and decrypted locally on the device using the same password.
//
// Example usage:
//
//	ivBytes, ivHex, err := crypto.GenerateIV()
//	encrypted, err := crypto.EncryptMessage("sensitive data", "password", ivBytes)
//	// Send encrypted message and ivHex to API
//
// Note: SHA1 is used for key derivation to maintain compatibility with the
// existing Pincho app implementation. For new implementations, consider
// using PBKDF2 or Argon2.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// CustomBase64Encode encodes bytes using custom Base64 encoding matching Pincho app.
//
// Converts standard Base64 characters to custom encoding:
//   - '+' → '-'
//   - '/' → '.'
//   - '=' → '_'
func CustomBase64Encode(data []byte) string {
	standard := base64.StdEncoding.EncodeToString(data)
	custom := strings.ReplaceAll(standard, "+", "-")
	custom = strings.ReplaceAll(custom, "/", ".")
	custom = strings.ReplaceAll(custom, "=", "_")
	return custom
}

// DeriveEncryptionKey derives AES encryption key from password using SHA1.
//
// Key derivation process:
//  1. SHA1 hash of password
//  2. Lowercase hexadecimal string
//  3. Truncate to 32 characters
//  4. Convert hex string to bytes
//
// Returns 16-byte AES-128 key.
func DeriveEncryptionKey(password string) ([]byte, error) {
	hash := sha1.Sum([]byte(password))
	keyHex := strings.ToLower(hex.EncodeToString(hash[:]))[:32]

	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	return key, nil
}

// pkcs7Pad applies PKCS7 padding to data.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padLength := blockSize - (len(data) % blockSize)
	padding := make([]byte, padLength)
	for i := range padding {
		padding[i] = byte(padLength)
	}
	return append(data, padding...)
}

// EncryptMessage encrypts text using AES-128-CBC with custom Base64 encoding.
//
// Encryption process matching Pincho app:
//  1. Derive key from password using SHA1
//  2. Apply PKCS7 padding to plaintext
//  3. Encrypt using AES-128-CBC with provided IV
//  4. Encode with custom Base64
//
// Returns encrypted and custom Base64 encoded string.
func EncryptMessage(plaintext, password string, iv []byte) (string, error) {
	// Derive encryption key
	key, err := DeriveEncryptionKey(password)
	if err != nil {
		return "", err
	}

	// Apply PKCS7 padding
	plaintextBytes := []byte(plaintext)
	padded := pkcs7Pad(plaintextBytes, aes.BlockSize)

	// Create AES cipher in CBC mode
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(padded))
	mode.CryptBlocks(encrypted, padded)

	// Return custom Base64 encoded result
	return CustomBase64Encode(encrypted), nil
}

// GenerateIV generates a random 16-byte initialization vector.
//
// Returns IV bytes and hexadecimal string representation (32 characters).
func GenerateIV() ([]byte, string, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, "", fmt.Errorf("failed to generate IV: %w", err)
	}

	ivHex := hex.EncodeToString(iv)
	return iv, ivHex, nil
}
