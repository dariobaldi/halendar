package mail

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Send sends the mail and returns its Message-ID.
func (m *Mailbox) Send(ctx context.Context, o Outgoing) (string, error) {
	id, raw, err := m.build(o)
	if err != nil {
		return "", err
	}
	client, err := m.smtpClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	if err := client.Mail(parseAddress(m.cfg.From)); err != nil {
		return "", fmt.Errorf("SMTP: sender refused: %w", err)
	}
	recipients := append(append(append([]string{}, o.To...), o.Cc...), o.Bcc...)
	for _, recipient := range recipients {
		if err := client.Rcpt(parseAddress(recipient)); err != nil {
			return "", fmt.Errorf("SMTP: recipient %q refused: %w", recipient, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return "", err
	}
	if _, err := w.Write(raw); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("SMTP: send refused: %w", err)
	}
	client.Quit()
	return id, nil
}

// Reply replies to a received message, in the same conversation thread.
func (m *Mailbox) Reply(ctx context.Context, original Message, text string) (string, error) {
	return m.Send(ctx, ReplyTo(original, text))
}

// ReplyTo prepares (without sending) a reply to a message: recipient, "Re:", and thread headers.
// Useful for SaveDraft, or to edit the reply before it is sent.
func ReplyTo(original Message, text string) Outgoing {
	subject := original.Subject
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(subject)), "re:") {
		subject = "Re: " + subject
	}
	// The message's own Reply-To header, when it has one, is where the sender wants
	// replies routed -- not necessarily their own From address (a newsletter or
	// automated sender is the common case).
	to := original.ReplyTo
	if to == "" {
		to = original.From
	}
	return Outgoing{To: []string{to}, Subject: subject, Text: text, InReplyTo: original.ID, References: original.References}
}

func (m *Mailbox) smtpClient() (*smtp.Client, error) {
	if m.cfg.SMTPHost == "" {
		return nil, errors.New("SMTP not configured (MAIL_SMTP_HOST)")
	}
	addr := net.JoinHostPort(m.cfg.SMTPHost, strconv.Itoa(m.cfg.SMTPPort))
	tlsCfg := &tls.Config{ServerName: m.cfg.SMTPHost}

	var conn net.Conn
	var err error
	if m.cfg.SMTPPort == 465 {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: 15 * time.Second}, "tcp", addr, tlsCfg)
	} else {
		conn, err = net.DialTimeout("tcp", addr, 15*time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("SMTP: could not connect to %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, m.cfg.SMTPHost)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if ok, _ := client.Extension("STARTTLS"); ok && m.cfg.SMTPPort != 465 {
		if err := client.StartTLS(tlsCfg); err != nil {
			client.Close()
			return nil, fmt.Errorf("SMTP STARTTLS: %w", err)
		}
	}
	if ok, _ := client.Extension("AUTH"); ok {
		var auth smtp.Auth
		if m.cfg.OAuth2Token != "" {
			auth = xoauth2SMTPAuth{user: m.cfg.User, token: m.cfg.OAuth2Token}
		} else {
			auth = smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.SMTPHost)
		}
		if err := client.Auth(auth); err != nil {
			client.Close()
			return nil, fmt.Errorf("SMTP: authentication refused (is it an app password or an expired token?): %w", err)
		}
	}
	return client, nil
}

// build produces the raw mail (RFC 5322): text only, or text + HTML.
func (m *Mailbox) build(o Outgoing) (string, []byte, error) {
	if len(o.To) == 0 || strings.TrimSpace(o.To[0]) == "" {
		return "", nil, errors.New(`"to" field missing`)
	}
	if strings.TrimSpace(o.Text) == "" && strings.TrimSpace(o.HTML) == "" {
		return "", nil, errors.New(`"text" field missing`)
	}

	domain := "halendar.local"
	if i := strings.LastIndex(m.cfg.From, "@"); i >= 0 {
		domain = strings.Trim(m.cfg.From[i+1:], "> ")
	}
	random := make([]byte, 8)
	rand.Read(random)
	id := fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), hex.EncodeToString(random), domain)

	var buf bytes.Buffer
	header := func(key, value string) {
		if value != "" {
			fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
		}
	}
	header("From", m.cfg.From)
	header("To", strings.Join(o.To, ", "))
	header("Cc", strings.Join(o.Cc, ", "))
	header("Subject", mime.QEncoding.Encode("utf-8", o.Subject))
	header("Date", time.Now().Format(time.RFC1123Z))
	header("Message-ID", id)
	header("In-Reply-To", Bracket(o.InReplyTo))

	var references []string
	for _, ref := range strings.Fields(o.References + " " + o.InReplyTo) {
		ref = Bracket(ref)
		if !slices.Contains(references, ref) {
			references = append(references, ref)
		}
	}
	header("References", strings.Join(references, " "))
	header("MIME-Version", "1.0")

	writePart := func(contentType, body string) {
		fmt.Fprintf(&buf, "Content-Type: %s; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", contentType)
		body = strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n")
		qp := quotedprintable.NewWriter(&buf)
		qp.Write([]byte(body))
		qp.Close()
		buf.WriteString("\r\n")
	}

	if o.HTML == "" {
		writePart("text/plain", o.Text)
	} else {
		text := o.Text
		if text == "" {
			text = HTMLToText(o.HTML)
		}
		boundary := "halendar-" + hex.EncodeToString(random)
		fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		writePart("text/plain", text)
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		writePart("text/html", o.HTML)
		fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	}
	return id, buf.Bytes(), nil
}

func parseAddress(s string) string {
	if i := strings.Index(s, "<"); i >= 0 {
		if j := strings.Index(s[i:], ">"); j > 0 {
			return s[i+1 : i+j]
		}
	}
	return strings.TrimSpace(s)
}
