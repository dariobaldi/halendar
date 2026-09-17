package emailimport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"halendar/mail"
)

// gmailScopes: full mail access (IMAP/SMTP need the broad "https://mail.google.com/"
// scope -- the narrower gmail.readonly one only works with the Gmail REST API, not
// IMAP), enough identity scope to look up the verified address that granted consent,
// and Calendar access -- bundled in here rather than requested separately, so
// connecting Gmail also connects the same Google account's calendar in one grant (the
// caller is expected to use the resulting refresh token for both). A user who wants
// Google Calendar without Gmail import still has calendarimport.GoogleProvider for a
// standalone connect.
var gmailScopes = []string{
	"https://mail.google.com/",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/calendar",
}

// GmailProvider implements Provider using Google's OAuth2 endpoints and Gmail's IMAP/
// SMTP servers (authenticated with the resulting access token, not a password).
type GmailProvider struct {
	oauth *oauth2.Config
	http  *http.Client
}

// NewGmailProvider builds a Provider for Gmail. clientID/clientSecret/redirectURL come
// from a Google Cloud OAuth client (Console > APIs & Services > Credentials).
func NewGmailProvider(clientID, clientSecret, redirectURL string) *GmailProvider {
	return &GmailProvider{
		oauth: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       gmailScopes,
			Endpoint:     google.Endpoint,
		},
		http: &http.Client{},
	}
}

func (p *GmailProvider) Name() string { return "gmail" }

func (p *GmailProvider) AuthCodeURL(state string) string {
	// access_type=offline + prompt=consent guarantee a refresh_token comes back even
	// if the user connected this same Google account before.
	return p.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
}

func (p *GmailProvider) Exchange(ctx context.Context, code string) (Token, string, error) {
	tok, err := p.oauth.Exchange(ctx, code)
	if err != nil {
		return Token{}, "", fmt.Errorf("gmail: exchanging code: %w", err)
	}
	if tok.RefreshToken == "" {
		return Token{}, "", fmt.Errorf("gmail: no refresh token returned (consent may not have been granted with offline access)")
	}
	email, err := p.verifiedEmail(ctx, tok.AccessToken)
	if err != nil {
		return Token{}, "", err
	}
	return Token{AccessToken: tok.AccessToken, RefreshToken: tok.RefreshToken}, email, nil
}

func (p *GmailProvider) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	src := p.oauth.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	tok, err := src.Token()
	if err != nil {
		return Token{}, fmt.Errorf("gmail: refreshing token: %w", err)
	}
	return Token{AccessToken: tok.AccessToken, RefreshToken: refreshToken}, nil
}

func (p *GmailProvider) Mailbox(email string, token Token) *mail.Mailbox {
	cfg := mail.Config{
		IMAPHost:    "imap.gmail.com:993",
		SMTPHost:    "smtp.gmail.com",
		SMTPPort:    587,
		User:        email,
		From:        email,
		OAuth2Token: token.AccessToken,
	}
	return mail.New(cfg)
}

// verifiedEmail asks Google which account the access token belongs to -- never trust
// a client-supplied email address for something that gets stored and polled server-side.
func (p *GmailProvider) verifiedEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("gmail: fetching account info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gmail: fetching account info: %s: %s", resp.Status, body)
	}

	var out struct {
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("gmail: invalid account info response: %w", err)
	}
	if out.Email == "" {
		return "", fmt.Errorf("gmail: account info response had no email")
	}
	return out.Email, nil
}
