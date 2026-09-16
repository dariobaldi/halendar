package mail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	_ "github.com/emersion/go-message/charset" // accented characters, ISO-8859-1, ...
	gomail "github.com/emersion/go-message/mail"
)

// withSession opens an authenticated IMAP connection, selects the folder, and runs fn.
func (m *Mailbox) withSession(folder string, readOnly bool, fn func(c *imapclient.Client, box *imap.SelectData) error) error {
	if m.cfg.IMAPHost == "" {
		return errors.New("IMAP not configured (MAIL_IMAP_HOST)")
	}

	var client *imapclient.Client
	var err error
	if m.cfg.IMAPInsecure {
		client, err = imapclient.DialInsecure(m.cfg.IMAPHost, nil)
	} else {
		client, err = imapclient.DialTLS(m.cfg.IMAPHost, nil)
	}
	if err != nil {
		return fmt.Errorf("IMAP: could not connect to %s: %w", m.cfg.IMAPHost, err)
	}
	defer client.Close()

	if err := client.Login(m.cfg.User, m.cfg.Pass).Wait(); err != nil {
		return fmt.Errorf("IMAP: login refused (is it an app password?): %w", err)
	}

	var box *imap.SelectData
	if folder != "" {
		box, err = client.Select(folder, &imap.SelectOptions{ReadOnly: readOnly}).Wait()
		if err != nil {
			return fmt.Errorf("IMAP: folder %q: %w", folder, err)
		}
	}

	if err := fn(client, box); err != nil {
		return err
	}
	client.Logout().Wait()
	return nil
}

// Count returns the number of messages in the folder.
func (m *Mailbox) Count(ctx context.Context) (uint32, error) {
	var n uint32
	err := m.withSession(m.cfg.Folder, true, func(c *imapclient.Client, box *imap.SelectData) error {
		n = box.NumMessages
		return nil
	})
	return n, err
}

// Folders lists the account's folders (INBOX, Sent, Drafts, ...).
func (m *Mailbox) Folders(ctx context.Context) ([]string, error) {
	var names []string
	err := m.withSession("", true, func(c *imapclient.Client, _ *imap.SelectData) error {
		list, err := c.List("", "*", nil).Collect()
		for _, folder := range list {
			names = append(names, folder.Mailbox)
		}
		return err
	})
	return names, err
}

// Recent returns the n most recent messages, newest first.
func (m *Mailbox) Recent(ctx context.Context, n int) ([]Message, error) {
	var msgs []Message
	err := m.withSession(m.cfg.Folder, true, func(c *imapclient.Client, box *imap.SelectData) error {
		if box.NumMessages == 0 {
			return nil
		}
		start := uint32(1)
		if box.NumMessages > uint32(n) {
			start = box.NumMessages - uint32(n) + 1
		}
		var seq imap.SeqSet
		seq.AddRange(start, box.NumMessages)
		var err error
		msgs, err = fetch(c, seq)
		return err
	})
	return newestFirst(msgs), err
}

// NewSince returns the messages whose UID is greater than lastUID, plus the highest UID seen.
// The first call, with lastUID = 0, returns nothing but gives a starting point (so the whole
// history isn't processed). Keep the returned UID for the next call.
func (m *Mailbox) NewSince(ctx context.Context, lastUID uint32) ([]Message, uint32, error) {
	var msgs []Message
	highest := lastUID
	err := m.withSession(m.cfg.Folder, true, func(c *imapclient.Client, box *imap.SelectData) error {
		if lastUID == 0 {
			if box.UIDNext > 0 {
				highest = uint32(box.UIDNext) - 1
			}
			return nil
		}
		if box.UIDNext != 0 && uint32(box.UIDNext) <= lastUID+1 {
			return nil // nothing new
		}
		var set imap.UIDSet
		set.AddRange(imap.UID(lastUID+1), 0) // 0 means up to the last one
		all, err := fetch(c, set)
		for _, msg := range all {
			if msg.UID > lastUID { // the server always returns at least the last one
				msgs = append(msgs, msg)
				if msg.UID > highest {
					highest = msg.UID
				}
			}
		}
		return err
	})
	return msgs, highest, err
}

// Read returns one message by UID.
func (m *Mailbox) Read(ctx context.Context, uid uint32) (*Message, error) {
	var msg *Message
	err := m.withSession(m.cfg.Folder, true, func(c *imapclient.Client, _ *imap.SelectData) error {
		msgs, err := fetch(c, imap.UIDSetNum(imap.UID(uid)))
		if err != nil {
			return err
		}
		if len(msgs) == 0 {
			return fmt.Errorf("message UID %d not found", uid)
		}
		msg = &msgs[0]
		return nil
	})
	return msg, err
}

// Search finds the messages matching the given criteria, most recent first.
func (m *Mailbox) Search(ctx context.Context, q SearchQuery) ([]Message, error) {
	criteria := &imap.SearchCriteria{Since: q.Since, Before: q.Before}
	if q.Unread {
		criteria.NotFlag = []imap.Flag{imap.FlagSeen}
	}
	if q.From != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{Key: "From", Value: q.From})
	}
	if q.Subject != "" {
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{Key: "Subject", Value: q.Subject})
	}
	if q.Contains != "" {
		criteria.Body = []string{q.Contains}
	}
	max := q.Max
	if max <= 0 {
		max = 50
	}

	var msgs []Message
	err := m.withSession(m.cfg.Folder, true, func(c *imapclient.Client, _ *imap.SelectData) error {
		result, err := c.UIDSearch(criteria, nil).Wait()
		if err != nil {
			return fmt.Errorf("IMAP: search: %w", err)
		}
		uids := result.AllUIDs()
		if len(uids) == 0 {
			return nil
		}
		sort.Slice(uids, func(i, j int) bool { return uids[i] > uids[j] })
		if len(uids) > max {
			uids = uids[:max]
		}
		msgs, err = fetch(c, imap.UIDSetNum(uids...))
		return err
	})
	return newestFirst(msgs), err
}

// MarkRead marks messages as read (read=true) or unread (read=false).
func (m *Mailbox) MarkRead(ctx context.Context, read bool, uids ...uint32) error {
	op := imap.StoreFlagsAdd
	if !read {
		op = imap.StoreFlagsDel
	}
	return m.withSession(m.cfg.Folder, false, func(c *imapclient.Client, _ *imap.SelectData) error {
		return c.Store(uidSet(uids), &imap.StoreFlags{Op: op, Silent: true, Flags: []imap.Flag{imap.FlagSeen}}, nil).Close()
	})
}

// Move moves messages to another folder (e.g. "Archives").
func (m *Mailbox) Move(ctx context.Context, folder string, uids ...uint32) error {
	return m.withSession(m.cfg.Folder, false, func(c *imapclient.Client, _ *imap.SelectData) error {
		_, err := c.Move(uidSet(uids), folder).Wait()
		return err
	})
}

// SaveDraft stores the mail in the account's Drafts folder instead of sending it:
// the user reviews it and sends it from their usual mail client.
func (m *Mailbox) SaveDraft(ctx context.Context, o Outgoing) (folder string, err error) {
	_, raw, err := m.build(o)
	if err != nil {
		return "", err
	}
	err = m.withSession("", false, func(c *imapclient.Client, _ *imap.SelectData) error {
		folder = draftsFolder(c)
		if folder == "" {
			return errors.New("IMAP: Drafts folder not found")
		}
		cmd := c.Append(folder, int64(len(raw)), &imap.AppendOptions{Flags: []imap.Flag{imap.FlagDraft, imap.FlagSeen}, Time: time.Now()})
		if _, err := cmd.Write(raw); err != nil {
			return err
		}
		if err := cmd.Close(); err != nil {
			return err
		}
		_, err := cmd.Wait()
		return err
	})
	return folder, err
}

// ── internal ────────────────────────────────────────────────────────────────

func draftsFolder(c *imapclient.Client) string {
	list, _ := c.List("", "*", nil).Collect()
	for _, folder := range list {
		for _, attr := range folder.Attrs {
			if attr == imap.MailboxAttrDrafts {
				return folder.Mailbox
			}
		}
	}
	for _, name := range []string{"Drafts", "Brouillons", "[Gmail]/Drafts", "[Gmail]/Brouillons", "INBOX.Drafts", "INBOX/Drafts"} {
		for _, folder := range list {
			if strings.EqualFold(folder.Mailbox, name) {
				return folder.Mailbox
			}
		}
	}
	return ""
}

func uidSet(uids []uint32) imap.UIDSet {
	var set imap.UIDSet
	for _, uid := range uids {
		set.AddNum(imap.UID(uid))
	}
	return set
}

func fetch(c *imapclient.Client, set imap.NumSet) ([]Message, error) {
	section := &imap.FetchItemBodySection{Peek: true} // Peek: does not mark the message as read
	bufs, err := c.Fetch(set, &imap.FetchOptions{
		UID: true, Flags: true, Envelope: true,
		BodySection: []*imap.FetchItemBodySection{section},
	}).Collect()
	if err != nil {
		return nil, fmt.Errorf("IMAP: reading messages: %w", err)
	}

	out := make([]Message, 0, len(bufs))
	for _, buf := range bufs {
		if buf.Envelope == nil {
			continue
		}
		env := buf.Envelope
		msg := Message{
			UID:     uint32(buf.UID),
			ID:      Bracket(env.MessageID), // the library strips the < >
			Subject: env.Subject,
			Date:    env.Date,
		}
		if len(env.From) > 0 {
			msg.From, msg.FromName = env.From[0].Addr(), env.From[0].Name
		}
		if len(env.ReplyTo) > 0 && env.ReplyTo[0].Addr() != "" {
			msg.From = env.ReplyTo[0].Addr()
		}
		for _, addr := range env.To {
			msg.To = append(msg.To, addr.Addr())
		}
		for _, addr := range env.Cc {
			msg.Cc = append(msg.Cc, addr.Addr())
		}
		for _, flag := range buf.Flags {
			if flag == imap.FlagSeen {
				msg.Read = true
			}
		}
		if msg.ID == "" {
			msg.ID = fmt.Sprintf("<uid-%d@imap>", buf.UID)
		}
		parseBody(&msg, buf.FindBodySection(section))
		out = append(out, msg)
	}
	return out, nil
}

func parseBody(msg *Message, raw []byte) {
	r, err := gomail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		return
	}
	msg.References = r.Header.Get("References")
	if msg.Date.IsZero() {
		msg.Date, _ = r.Header.Date()
	}
	for {
		part, err := r.NextPart()
		if errors.Is(err, io.EOF) || err != nil {
			break
		}
		switch header := part.Header.(type) {
		case *gomail.InlineHeader:
			contentType, _, _ := header.ContentType()
			content, _ := io.ReadAll(io.LimitReader(part.Body, 2<<20))
			switch {
			case contentType == "text/plain" && msg.Text == "":
				msg.Text = strings.TrimSpace(string(content))
			case contentType == "text/html" && msg.HTML == "":
				msg.HTML = string(content)
			}
		case *gomail.AttachmentHeader:
			name, _ := header.Filename()
			contentType, _, _ := header.ContentType()
			size, _ := io.Copy(io.Discard, part.Body)
			msg.Attachments = append(msg.Attachments, Attachment{Name: name, Type: contentType, Size: int(size)})
		}
	}
	if msg.Text == "" && msg.HTML != "" {
		msg.Text = HTMLToText(msg.HTML)
	}
}

// HTMLToText does a simple cleanup of an HTML body.
func HTMLToText(html string) string {
	var sb strings.Builder
	insideTag := false
	replacer := strings.NewReplacer("<br>", "\n", "<br/>", "\n", "<br />", "\n", "</p>", "\n", "</div>", "\n", "</tr>", "\n")
	for _, r := range replacer.Replace(html) {
		switch {
		case r == '<':
			insideTag = true
		case r == '>':
			insideTag = false
		case !insideTag:
			sb.WriteRune(r)
		}
	}
	text := strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&#39;", "'", "&quot;", `"`).Replace(sb.String())

	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func newestFirst(msgs []Message) []Message {
	sort.SliceStable(msgs, func(i, j int) bool { return msgs[i].UID > msgs[j].UID })
	return msgs
}
