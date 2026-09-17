package calendarimport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleProvider implements Provider using Google's OAuth2 endpoints and the Google
// Calendar REST API (a single freeBusy call is all the analysis pipeline needs today,
// so this hand-rolls that one request rather than pulling in the full generated
// Calendar API client).
type GoogleProvider struct {
	oauth *oauth2.Config
	http  *http.Client
}

// NewGoogleProvider builds a Provider for Google Calendar. Pass the calendar-only
// scope set when connecting a calendar on its own; when bundled into the Gmail
// connect flow, the calendar scope is requested there instead and this provider is
// only used to refresh the token and read free/busy data.
func NewGoogleProvider(clientID, clientSecret, redirectURL string) *GoogleProvider {
	return &GoogleProvider{
		oauth: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{CalendarScope, "https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		},
		http: &http.Client{},
	}
}

// CalendarScope is the Google OAuth2 scope for full Calendar access (read + write,
// since a later step will need to write confirmed events, not just read busy times).
// Also added to the Gmail connect flow's scope list, so that grant covers this too.
const CalendarScope = "https://www.googleapis.com/auth/calendar"

func (p *GoogleProvider) Name() string { return "google" }

func (p *GoogleProvider) AuthCodeURL(state string) string {
	return p.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
}

func (p *GoogleProvider) Exchange(ctx context.Context, code string) (Token, string, error) {
	tok, err := p.oauth.Exchange(ctx, code)
	if err != nil {
		return Token{}, "", fmt.Errorf("google calendar: exchanging code: %w", err)
	}
	if tok.RefreshToken == "" {
		return Token{}, "", fmt.Errorf("google calendar: no refresh token returned (consent may not have been granted with offline access)")
	}
	email, err := p.verifiedEmail(ctx, tok.AccessToken)
	if err != nil {
		return Token{}, "", err
	}
	return Token{AccessToken: tok.AccessToken, RefreshToken: tok.RefreshToken}, email, nil
}

func (p *GoogleProvider) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	src := p.oauth.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	tok, err := src.Token()
	if err != nil {
		return Token{}, fmt.Errorf("google calendar: refreshing token: %w", err)
	}
	return Token{AccessToken: tok.AccessToken, RefreshToken: refreshToken}, nil
}

func (p *GoogleProvider) Source(token Token) Source {
	return &googleSource{http: p.http, accessToken: token.AccessToken}
}

func (p *GoogleProvider) verifiedEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("google calendar: fetching account info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google calendar: fetching account info: %s: %s", resp.Status, body)
	}

	var out struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("google calendar: invalid account info response: %w", err)
	}
	if out.Email == "" {
		return "", fmt.Errorf("google calendar: account info response had no email")
	}
	return out.Email, nil
}

// googleSource implements Source against the "primary" calendar of one Google
// account.
type googleSource struct {
	http        *http.Client
	accessToken string

	loc     *time.Location
	loadedT bool
}

func (s *googleSource) Busy(ctx context.Context, start, end time.Time) (bool, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"timeMin": start.UTC().Format(time.RFC3339),
		"timeMax": end.UTC().Format(time.RFC3339),
		"items":   []map[string]string{{"id": "primary"}},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://www.googleapis.com/calendar/v3/freeBusy", bytes.NewReader(reqBody))
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("google calendar: freeBusy request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("google calendar: freeBusy: %s: %s", resp.Status, body)
	}

	var out struct {
		Calendars map[string]struct {
			Busy []struct {
				Start string `json:"start"`
				End   string `json:"end"`
			} `json:"busy"`
		} `json:"calendars"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return false, fmt.Errorf("google calendar: invalid freeBusy response: %w", err)
	}
	return len(out.Calendars["primary"].Busy) > 0, nil
}

// Timezone returns the account's configured calendar timezone, falling back to
// Europe/Paris (matching the rest of the app's default) if it can't be read.
func (s *googleSource) Timezone() *time.Location {
	if s.loadedT {
		return s.loc
	}
	s.loadedT = true
	s.loc, _ = time.LoadLocation("Europe/Paris")

	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/calendar/v3/users/me/settings/timezone", nil)
	if err != nil {
		return s.loc
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	resp, err := s.http.Do(req)
	if err != nil {
		return s.loc
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return s.loc
	}

	var out struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out.Value == "" {
		return s.loc
	}
	if loc, err := time.LoadLocation(out.Value); err == nil {
		s.loc = loc
	}
	return s.loc
}
