// Package emailimport connects a user's messaging account (Gmail today, others later)
// to the app: the OAuth2 dance to obtain credentials, and turning those credentials
// into a *mail.Mailbox the rest of the backend already knows how to read.
//
// Adding a new provider (e.g. La Messagerie / La Suite numérique) means implementing
// Provider and registering it in a Registry -- nothing else in the import pipeline
// needs to change.
package emailimport

import (
	"context"

	"halendar/mail"
)

// Token is a provider's access credential for one account.
type Token struct {
	AccessToken string
	// RefreshToken is only ever returned by Exchange (the first consent); Refresh
	// calls do not normally hand back a new one, so callers keep reusing the one
	// they stored at connect time.
	RefreshToken string
}

// Provider is one way of connecting and reading a messaging account.
type Provider interface {
	// Name identifies the provider, e.g. "gmail". Stored on EmailAccount.Provider.
	Name() string

	// AuthCodeURL returns the URL to send the user's browser to, to start the OAuth2
	// consent flow. state is opaque to the provider and must be passed back unchanged
	// to the redirect URI.
	AuthCodeURL(state string) string

	// Exchange trades an OAuth2 authorization code (from the redirect back to us) for
	// a token, and returns the verified email address of the account that granted
	// consent -- never trust a client-supplied address instead.
	Exchange(ctx context.Context, code string) (Token, string, error)

	// Refresh trades a stored refresh token for a fresh access token, e.g. right
	// before a sync pass (access tokens are short-lived).
	Refresh(ctx context.Context, refreshToken string) (Token, error)

	// Mailbox builds a *mail.Mailbox authenticated as email using a fresh access
	// token from Refresh.
	Mailbox(email string, token Token) *mail.Mailbox
}

// Registry looks up a Provider by name.
type Registry map[string]Provider

func (r Registry) Get(name string) (Provider, bool) {
	p, ok := r[name]
	return p, ok
}
