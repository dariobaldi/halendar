// Package calendarimport connects a user's calendar (Google Calendar, a CalDAV server
// such as iCloud or a groupware's calendar, ...) to the app, and turns whatever
// credentials that takes into a Source the rest of the backend can check availability
// against without caring which provider it is.
//
// Adding a new OAuth2-based provider means implementing Provider and registering it
// in a Registry; CalDAV-based providers (Apple, La Suite if it speaks CalDAV, ...)
// don't need a Provider at all -- they go through the existing halendar/calendar
// package directly, configured per account instead of once globally.
package calendarimport

import (
	"context"
	"time"
)

// Token is a provider's access credential for one calendar.
type Token struct {
	AccessToken string
	// RefreshToken is only ever returned by Exchange (the first consent); Refresh
	// calls do not normally hand back a new one, so callers keep reusing the one
	// they stored at connect time.
	RefreshToken string
}

// Source is the minimal interface the email-analysis pipeline needs from a connected
// calendar, regardless of provider.
type Source interface {
	// Busy reports whether [start, end] overlaps an existing event.
	Busy(ctx context.Context, start, end time.Time) (bool, error)
	Timezone() *time.Location
}

// Provider is one OAuth2-based way of connecting and reading a calendar.
type Provider interface {
	// Name identifies the provider, e.g. "google". Stored on CalendarAccount.Provider.
	Name() string

	// AuthCodeURL returns the URL to send the user's browser to, to start the OAuth2
	// consent flow. state is opaque to the provider and must be passed back unchanged
	// to the redirect URI.
	AuthCodeURL(state string) string

	// Exchange trades an OAuth2 authorization code for a token, and returns a display
	// name for the connected calendar (e.g. the account's email) -- never trust a
	// client-supplied one instead.
	Exchange(ctx context.Context, code string) (Token, string, error)

	// Refresh trades a stored refresh token for a fresh access token, e.g. right
	// before a busy-check (access tokens are short-lived).
	Refresh(ctx context.Context, refreshToken string) (Token, error)

	// Source builds a Source authenticated with a fresh access token from Refresh.
	Source(token Token) Source
}

// Registry looks up a Provider by name.
type Registry map[string]Provider

func (r Registry) Get(name string) (Provider, bool) {
	p, ok := r[name]
	return p, ok
}
