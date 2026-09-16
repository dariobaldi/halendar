// Package mail reads (IMAP) and sends (SMTP) email.
//
//	cfg := mail.ConfigFromEnv()          // or fill in mail.Config by hand
//	mailbox := mail.New(cfg)
//	msgs, _ := mailbox.Recent(ctx, 10)
//	mailbox.Reply(ctx, msgs[0], "Thanks, noted.")
package mail

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dariobaldi/halendar_back/internal/envfile"
)

// Config describes one mail account. The SMTP fields are guessed from the
// IMAP host when left empty.
type Config struct {
	IMAPHost     string // "imap.gmail.com:993"
	IMAPInsecure bool   // true only for a local test server, never for a real account
	Folder       string // "INBOX" by default

	SMTPHost string // "smtp.gmail.com" (guessed from IMAPHost when empty)
	SMTPPort int    // 587 (STARTTLS) or 465 (direct TLS)

	User string // account address
	Pass string // password (an app password for Gmail, iCloud, Yahoo, ...)
	From string // display sender, e.g. "Halendar Team <me@gmail.com>" (defaults to User)
}

// ConfigFromEnv reads MAIL_* from the environment (after envfile.Load(".env")).
func ConfigFromEnv() Config {
	cfg := Config{
		IMAPHost:     envfile.String("MAIL_IMAP_HOST", ""),
		IMAPInsecure: !envfile.Bool("MAIL_IMAP_TLS", true),
		Folder:       envfile.String("MAIL_FOLDER", "INBOX"),
		SMTPHost:     envfile.String("MAIL_SMTP_HOST", ""),
		SMTPPort:     envfile.Int("MAIL_SMTP_PORT", 0),
		User:         envfile.String("MAIL_USER", ""),
		Pass:         envfile.String("MAIL_PASS", ""),
		From:         envfile.String("MAIL_FROM", ""),
	}
	return cfg.withDefaults()
}

// Missing returns which required fields are still empty (empty slice = configuration is OK).
func (c Config) Missing() []string {
	var missing []string
	if c.IMAPHost == "" && c.SMTPHost == "" {
		missing = append(missing, "MAIL_IMAP_HOST")
	}
	if c.User == "" {
		missing = append(missing, "MAIL_USER")
	}
	if c.Pass == "" {
		missing = append(missing, "MAIL_PASS")
	}
	return missing
}

func (c Config) withDefaults() Config {
	if c.Folder == "" {
		c.Folder = "INBOX"
	}
	if c.SMTPHost == "" && strings.HasPrefix(c.IMAPHost, "imap.") {
		host := strings.TrimPrefix(c.IMAPHost, "imap.")
		if i := strings.LastIndex(host, ":"); i >= 0 {
			host = host[:i]
		}
		c.SMTPHost = "smtp." + host
	}
	if i := strings.LastIndex(c.SMTPHost, ":"); i >= 0 { // "smtp.x.com:465" is accepted
		if port, err := strconv.Atoi(c.SMTPHost[i+1:]); err == nil {
			c.SMTPPort = port
			c.SMTPHost = c.SMTPHost[:i]
		}
	}
	if c.SMTPPort == 0 {
		c.SMTPPort = 587
	}
	if c.From == "" {
		c.From = c.User
	}
	return c
}

// Message is a received email.
type Message struct {
	UID         uint32       `json:"uid"`
	ID          string       `json:"id"` // Message-ID, e.g. <abc@example.com>
	From        string       `json:"from"`
	FromName    string       `json:"from_name,omitempty"`
	To          []string     `json:"to,omitempty"`
	Cc          []string     `json:"cc,omitempty"`
	Subject     string       `json:"subject"`
	Date        time.Time    `json:"date"`
	Text        string       `json:"text"`           // plain-text part (or HTML converted to text)
	HTML        string       `json:"html,omitempty"` // raw HTML part, if present
	References  string       `json:"references,omitempty"`
	Read        bool         `json:"read"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment describes an attachment without its content.
type Attachment struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

// Outgoing is an email to be sent.
type Outgoing struct {
	To         []string `json:"to"`
	Cc         []string `json:"cc,omitempty"`
	Bcc        []string `json:"bcc,omitempty"`
	Subject    string   `json:"subject"`
	Text       string   `json:"text"`
	HTML       string   `json:"html,omitempty"` // optional: HTML version alongside the text
	InReplyTo  string   `json:"in_reply_to,omitempty"`
	References string   `json:"references,omitempty"`
}

// SearchQuery holds combined search criteria (all optional).
type SearchQuery struct {
	Unread   bool      `json:"unread,omitempty"`
	Since    time.Time `json:"since,omitempty"`
	Before   time.Time `json:"before,omitempty"`
	From     string    `json:"from,omitempty"`
	Subject  string    `json:"subject,omitempty"`
	Contains string    `json:"contains,omitempty"` // searched in the body
	Max      int       `json:"max,omitempty"`      // most recent first; 0 means 50
}

// Mailbox groups reading and sending for one account.
type Mailbox struct {
	cfg Config
}

func New(cfg Config) *Mailbox { return &Mailbox{cfg: cfg.withDefaults()} }

func (m *Mailbox) Config() Config { return m.cfg }

// Test checks the IMAP and SMTP connections without changing anything.
func (m *Mailbox) Test(ctx context.Context) error {
	if missing := m.cfg.Missing(); len(missing) > 0 {
		return fmt.Errorf("incomplete configuration: %s", strings.Join(missing, ", "))
	}
	var errs []string
	if m.cfg.IMAPHost != "" {
		if _, err := m.Count(ctx); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if m.cfg.SMTPHost != "" {
		client, err := m.smtpClient()
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			client.Quit()
			client.Close()
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

// Bracket normalizes a Message-ID to the <id@domain> form (required to stay in the thread).
func Bracket(id string) string {
	id = strings.TrimSpace(id)
	if id == "" || strings.HasPrefix(id, "<") {
		return id
	}
	return "<" + id + ">"
}
