// Package claude is a minimal client for Anthropic's Messages API, used as an
// alternative to the local Ollama model (internal/ollama) when a user has connected
// their own Claude API key and activated it. Exposes the same Generate /
// GenerateDeterministic shape as internal/ollama.Client so cmd/api's email-analysis
// pipeline can use either one interchangeably -- see aiClient in cmd/api/email_import.go.
package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.anthropic.com"
	apiVersion     = "2023-06-01"
	// maxTokens caps the reply, not the prompt -- generous for both a JSON extraction
	// result and a short drafted email, without letting a runaway response burn
	// through the user's own API quota.
	maxTokens = 1024
)

// Client talks to the Anthropic API using one caller-supplied API key and model.
// Unlike internal/ollama.Client (one shared instance for the whole server), a Client
// is built fresh per request from a user's own decrypted key -- see
// cmd/api's aiClientFor.
type Client struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

// New returns a Client for apiKey using model (e.g. "claude-haiku-4-5-20251001").
func New(apiKey, model string) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{Timeout: 2 * time.Minute},
	}
}

// Model returns the model name this client sends prompts to.
func (c *Client) Model() string { return c.model }

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Messages    []message `json:"messages"`
	Temperature *float64  `json:"temperature,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type messagesResponse struct {
	Content []contentBlock `json:"content"`
	Error   *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Generate sends prompt to the model and returns its full response text, using the
// model's default sampling.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	return c.generate(ctx, prompt, nil)
}

// GenerateDeterministic is like Generate but disables sampling (temperature 0), for
// tasks that need consistent, well-formed output rather than varied phrasing -- see
// the equivalent method on internal/ollama.Client for why that matters here.
func (c *Client) GenerateDeterministic(ctx context.Context, prompt string) (string, error) {
	temp := 0.0
	return c.generate(ctx, prompt, &temp)
}

func (c *Client) generate(ctx context.Context, prompt string, temperature *float64) (string, error) {
	body, err := json.Marshal(messagesRequest{
		Model:       c.model,
		MaxTokens:   maxTokens,
		Messages:    []message{{Role: "user", Content: prompt}},
		Temperature: temperature,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude: could not reach %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("claude: reading response: %w", err)
	}

	var out messagesResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("claude: invalid response (%s): %w", resp.Status, err)
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != nil {
			return "", fmt.Errorf("claude: %s", out.Error.Message)
		}
		return "", fmt.Errorf("claude: %s", resp.Status)
	}
	if len(out.Content) == 0 {
		return "", nil
	}
	return out.Content[0].Text, nil
}

// VerifyKey makes one minimal, cheap request to confirm apiKey is valid for model.
// Used when a user connects a new key, so a typo or revoked key is caught immediately
// instead of silently falling back or failing on the next background analysis pass.
func VerifyKey(ctx context.Context, apiKey, model string) error {
	_, err := New(apiKey, model).GenerateDeterministic(ctx, "Reply with only the word OK.")
	return err
}
