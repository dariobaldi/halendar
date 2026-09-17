// Package push sends push notifications to a specific device through Firebase
// Cloud Messaging's HTTP v1 API, authenticated with a Google service account
// (Firebase Console: Project Settings > Service Accounts > Generate new private key).
//
//	client := push.New(push.Config{ProjectID: "...", ServiceAccountJSON: keyBytes})
//	err := client.Send(ctx, deviceToken, push.Notification{Title: "Hi", Body: "..."})
package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

// Config describes one Firebase project's credentials.
type Config struct {
	ProjectID          string
	ServiceAccountJSON []byte // the raw downloaded service account key file
}

// Missing returns which required fields are still empty.
func (c Config) Missing() []string {
	var missing []string
	if c.ProjectID == "" {
		missing = append(missing, "FCM_PROJECT_ID")
	}
	if len(c.ServiceAccountJSON) == 0 {
		missing = append(missing, "FCM_SERVICE_ACCOUNT_FILE")
	}
	return missing
}

// Notification is one push notification to send to a single device.
type Notification struct {
	Title string
	Body  string
	// Data is an optional custom key/value payload delivered alongside the
	// notification, e.g. {"type": "proposal", "mail_uid": "123"} for the app to
	// deep-link on tap.
	Data map[string]string
}

// Client is a lazily authenticated connection to one Firebase project.
type Client struct {
	cfg Config

	mu   sync.Mutex
	http *http.Client
	err  error
}

func New(cfg Config) *Client { return &Client{cfg: cfg} }

// Test checks that the service account credentials are valid.
func (c *Client) Test(ctx context.Context) error {
	if missing := c.cfg.Missing(); len(missing) > 0 {
		return fmt.Errorf("push not configured: %s", strings.Join(missing, ", "))
	}
	_, err := c.httpClient(ctx)
	return err
}

// Send delivers a notification to one device by its FCM push token.
func (c *Client) Send(ctx context.Context, pushToken string, n Notification) error {
	httpClient, err := c.httpClient(ctx)
	if err != nil {
		return err
	}

	var payload fcmRequest
	payload.Message.Token = pushToken
	if n.Title != "" || n.Body != "" {
		payload.Message.Notification = &fcmNotification{Title: n.Title, Body: n.Body}
	}
	payload.Message.Data = n.Data

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.cfg.ProjectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("push: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push: FCM returned %s: %s", resp.Status, string(raw))
	}
	return nil
}

type fcmRequest struct {
	Message struct {
		Token        string            `json:"token"`
		Notification *fcmNotification  `json:"notification,omitempty"`
		Data         map[string]string `json:"data,omitempty"`
	} `json:"message"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// httpClient builds (once) and caches an HTTP client whose transport auto-attaches
// and refreshes an OAuth2 access token derived from the service account key.
func (c *Client) httpClient(ctx context.Context) (*http.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.http != nil || c.err != nil {
		return c.http, c.err
	}
	if missing := c.cfg.Missing(); len(missing) > 0 {
		c.err = fmt.Errorf("push not configured: %s", strings.Join(missing, ", "))
		return nil, c.err
	}

	creds, err := google.CredentialsFromJSON(ctx, c.cfg.ServiceAccountJSON, fcmScope)
	if err != nil {
		c.err = fmt.Errorf("push: invalid service account credentials: %w", err)
		return nil, c.err
	}
	c.http = oauth2.NewClient(ctx, creds.TokenSource)
	return c.http, nil
}
