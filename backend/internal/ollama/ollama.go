// Package ollama is a minimal client for a local Ollama server (see the "ollama"
// service in docker-compose.yml), used to send prompts to the configured model.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to one Ollama server, using one model.
type Client struct {
	baseURL string
	model   string
	http    *http.Client
}

// New returns a Client for baseURL (e.g. "http://ollama:11434") using model (e.g. "gemma3:1b").
func New(baseURL, model string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http:    &http.Client{Timeout: 2 * time.Minute},
	}
}

// Model returns the model name this client sends prompts to.
func (c *Client) Model() string { return c.model }

type generateRequest struct {
	Model   string           `json:"model"`
	Prompt  string           `json:"prompt"`
	Stream  bool             `json:"stream"`
	Options *generateOptions `json:"options,omitempty"`
}

type generateOptions struct {
	Temperature float64 `json:"temperature"`
}

type generateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

// Generate sends prompt to the model and returns its full (non-streamed) response
// text, using the model's default sampling.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	return c.generate(ctx, prompt, nil)
}

// GenerateDeterministic is like Generate but disables sampling (temperature 0).
// Prompt-tuning found this makes a real difference for gemma3:1b on tasks that need
// consistent, well-formed output -- both stricter JSON compliance for extraction and
// more reliable instruction-following (no sign-off, matching the original language)
// for the reply draft -- rather than the varied phrasing normal sampling gives.
func (c *Client) GenerateDeterministic(ctx context.Context, prompt string) (string, error) {
	return c.generate(ctx, prompt, &generateOptions{Temperature: 0})
}

func (c *Client) generate(ctx context.Context, prompt string, options *generateOptions) (string, error) {
	body, err := json.Marshal(generateRequest{Model: c.model, Prompt: prompt, Stream: false, Options: options})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: could not reach %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: reading response: %w", err)
	}

	var out generateResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("ollama: invalid response (%s): %w", resp.Status, err)
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != "" {
			return "", fmt.Errorf("ollama: %s", out.Error)
		}
		return "", fmt.Errorf("ollama: %s", resp.Status)
	}
	return out.Response, nil
}
