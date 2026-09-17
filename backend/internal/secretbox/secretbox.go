// Package secretbox encrypts small secrets (OAuth refresh tokens, IMAP passwords, ...)
// before they are stored in the database, using AES-256-GCM with a single server-held
// key (ENCRYPTION_KEY). It is not meant for large payloads.
package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// Key is a 32-byte AES-256 key.
type Key [32]byte

// ParseKey decodes a base64-encoded 32-byte key, as produced by GenerateKey.
func ParseKey(encoded string) (Key, error) {
	var key Key
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return key, fmt.Errorf("secretbox: invalid key encoding: %w", err)
	}
	if len(raw) != len(key) {
		return key, fmt.Errorf("secretbox: key must be %d bytes, got %d", len(key), len(raw))
	}
	copy(key[:], raw)
	return key, nil
}

// GenerateKey returns a new random key, base64-encoded for storage in an env var.
func GenerateKey() (string, error) {
	var key Key
	if _, err := rand.Read(key[:]); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key[:]), nil
}

// Seal encrypts plaintext, returning nonce||ciphertext.
func Seal(key Key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Open decrypts a value produced by Seal.
func Open(key Key, sealed []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(sealed) < gcm.NonceSize() {
		return nil, errors.New("secretbox: ciphertext too short")
	}
	nonce, ciphertext := sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
