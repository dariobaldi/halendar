package main

import (
	"encoding/json"

	"github.com/dariobaldi/halendar_back/internal/secretbox"
)

// sealSecret JSON-encodes v and encrypts it with the app's encryption key, ready to
// store in an *_account_credentials table.
func (app *app) sealSecret(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return secretbox.Seal(app.encryptionKey, raw)
}

// openSecret decrypts a value produced by sealSecret and decodes it into v.
func (app *app) openSecret(sealed []byte, v any) error {
	raw, err := secretbox.Open(app.encryptionKey, sealed)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}
