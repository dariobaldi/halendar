// Package testutil provides fake servers (IMAP, SMTP, CalDAV) to test the whole
// flow without a real account or network access. Used by the integration tests.
package testutil

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/emersion/go-imap/v2/imapserver/imapmemserver"
	"github.com/emersion/go-webdav/caldav"
)

// ── IMAP ────────────────────────────────────────────────────────────────────

type FakeIMAP struct {
	Addr, User, Pass string
	user             *imapmemserver.User
}

func NewFakeIMAP(t *testing.T) *FakeIMAP {
	f := &FakeIMAP{User: "me@test.local", Pass: "test"}
	mem := imapmemserver.New()
	f.user = imapmemserver.NewUser(f.User, f.Pass)
	f.user.Create("INBOX", nil)
	f.user.Create("Drafts", &imap.CreateOptions{SpecialUse: []imap.MailboxAttr{imap.MailboxAttrDrafts}})
	f.user.Create("Archives", nil)
	mem.AddUser(f.user)

	srv := imapserver.New(&imapserver.Options{
		NewSession: func(*imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
			return mem.NewSession(), nil, nil
		},
		Caps:         imap.CapSet{imap.CapIMAP4rev1: {}, imap.CapIMAP4rev2: {}},
		InsecureAuth: true,
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(ln)
	t.Cleanup(func() { srv.Close() })
	f.Addr = ln.Addr().String()
	return f
}

// Deliver adds a mail to the inbox.
func (f *FakeIMAP) Deliver(t *testing.T, id, from, subject, text string) {
	f.DeliverWithReplyTo(t, id, from, "", subject, text)
}

// DeliverWithReplyTo is like Deliver but adds a Reply-To header when replyTo isn't
// empty, for exercising the case where a reply should be routed somewhere other than
// From (see mail.Message.ReplyTo).
func (f *FakeIMAP) DeliverWithReplyTo(t *testing.T, id, from, replyTo, subject, text string) {
	replyToHeader := ""
	if replyTo != "" {
		replyToHeader = "Reply-To: " + replyTo + "\r\n"
	}
	raw := fmt.Sprintf("From: %s\r\n%sTo: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n",
		from, replyToHeader, f.User, subject, time.Now().Format(time.RFC1123Z), id, text)
	if _, err := f.user.Append("INBOX", bytes.NewReader([]byte(raw)), &imap.AppendOptions{Time: time.Now()}); err != nil {
		t.Fatal(err)
	}
}

// Count returns the number of messages in a folder.
func (f *FakeIMAP) Count(folder string) uint32 {
	status, err := f.user.Status(folder, &imap.StatusOptions{NumMessages: true})
	if err != nil || status.NumMessages == nil {
		return 0
	}
	return *status.NumMessages
}

// ── SMTP ────────────────────────────────────────────────────────────────────

type FakeSMTP struct {
	Host string
	Port int

	mu       sync.Mutex
	received []ReceivedMail
}

type ReceivedMail struct {
	From string
	To   []string
	Raw  string
}

func NewFakeSMTP(t *testing.T) *FakeSMTP {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	f := &FakeSMTP{Host: "127.0.0.1", Port: ln.Addr().(*net.TCPAddr).Port}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go f.session(conn)
		}
	}()
	return f
}

func (f *FakeSMTP) session(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	writeLine := func(s string) { fmt.Fprint(conn, s+"\r\n") }
	writeLine("220 fake-smtp")

	var mail ReceivedMail
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
			writeLine("250-fake-smtp")
			writeLine("250 AUTH PLAIN")
		case strings.HasPrefix(cmd, "AUTH"):
			writeLine("235 ok")
		case strings.HasPrefix(cmd, "MAIL FROM"):
			mail = ReceivedMail{From: strings.TrimSpace(line[10:])}
			writeLine("250 ok")
		case strings.HasPrefix(cmd, "RCPT TO"):
			mail.To = append(mail.To, strings.Trim(strings.TrimSpace(line[8:]), "<>"))
			writeLine("250 ok")
		case cmd == "DATA":
			writeLine("354 go")
			var body strings.Builder
			for {
				l, err := r.ReadString('\n')
				if err != nil || l == ".\r\n" {
					break
				}
				body.WriteString(l)
			}
			mail.Raw = body.String()
			f.mu.Lock()
			f.received = append(f.received, mail)
			f.mu.Unlock()
			writeLine("250 queued")
		case cmd == "QUIT":
			writeLine("221 bye")
			return
		default:
			writeLine("250 ok")
		}
	}
}

func (f *FakeSMTP) Mails() []ReceivedMail {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]ReceivedMail(nil), f.received...)
}

// ── CalDAV ──────────────────────────────────────────────────────────────────

type FakeCalDAV struct {
	URL, User, Pass string
	backend         *calDAVBackend
}

const calendarPath = "/me/calendars/work/"

func NewFakeCalDAV(t *testing.T) *FakeCalDAV {
	f := &FakeCalDAV{User: "me", Pass: "test", backend: &calDAVBackend{objects: map[string]caldav.CalendarObject{}}}
	handler := &caldav.Handler{Backend: f.backend}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, pass, ok := r.BasicAuth(); !ok || user != f.User || pass != f.Pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="fake"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	f.URL = srv.URL + "/"
	return f
}

// Book adds an existing booking.
func (f *FakeCalDAV) Book(uid, title string, start, end time.Time) {
	event := ical.NewEvent()
	event.Props.SetText(ical.PropUID, uid)
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	event.Props.SetText(ical.PropSummary, title)
	event.Props.SetDateTime(ical.PropDateTimeStart, start.UTC())
	event.Props.SetDateTime(ical.PropDateTimeEnd, end.UTC())
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//test//EN")
	cal.Children = append(cal.Children, event.Component)
	f.backend.PutCalendarObject(context.Background(), calendarPath+uid+".ics", cal, nil)
}

// Event returns (title, STATUS) of an event by uid, ok=false if it does not exist.
func (f *FakeCalDAV) Event(uid string) (title, status string, ok bool) {
	f.backend.mu.Lock()
	defer f.backend.mu.Unlock()
	obj, exists := f.backend.objects[calendarPath+uid+".ics"]
	if !exists {
		return "", "", false
	}
	event := obj.Data.Events()[0]
	title, _ = event.Props.Text(ical.PropSummary)
	status, _ = event.Props.Text(ical.PropStatus)
	return title, status, true
}

type calDAVBackend struct {
	mu      sync.Mutex
	objects map[string]caldav.CalendarObject
}

func (b *calDAVBackend) CurrentUserPrincipal(context.Context) (string, error) { return "/me/", nil }
func (b *calDAVBackend) CalendarHomeSetPath(context.Context) (string, error) {
	return "/me/calendars/", nil
}
func (b *calDAVBackend) CreateCalendar(context.Context, *caldav.Calendar) error {
	return fmt.Errorf("not supported")
}
func (b *calDAVBackend) ListCalendars(context.Context) ([]caldav.Calendar, error) {
	return []caldav.Calendar{{Path: calendarPath, Name: "Work", SupportedComponentSet: []string{"VEVENT"}}}, nil
}
func (b *calDAVBackend) GetCalendar(ctx context.Context, path string) (*caldav.Calendar, error) {
	calendars, _ := b.ListCalendars(ctx)
	if path == calendarPath {
		return &calendars[0], nil
	}
	return nil, fmt.Errorf("not found")
}
func (b *calDAVBackend) GetCalendarObject(_ context.Context, path string, _ *caldav.CalendarCompRequest) (*caldav.CalendarObject, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if obj, ok := b.objects[path]; ok {
		return &obj, nil
	}
	return nil, fmt.Errorf("not found")
}
func (b *calDAVBackend) ListCalendarObjects(_ context.Context, path string, _ *caldav.CalendarCompRequest) ([]caldav.CalendarObject, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []caldav.CalendarObject
	for key, obj := range b.objects {
		if strings.HasPrefix(key, path) {
			out = append(out, obj)
		}
	}
	return out, nil
}
func (b *calDAVBackend) QueryCalendarObjects(ctx context.Context, path string, q *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	all, _ := b.ListCalendarObjects(ctx, path, nil)
	return caldav.Filter(q, all)
}
func (b *calDAVBackend) PutCalendarObject(_ context.Context, path string, cal *ical.Calendar, _ *caldav.PutCalendarObjectOptions) (*caldav.CalendarObject, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	obj := caldav.CalendarObject{Path: path, ModTime: time.Now(), ETag: fmt.Sprint(time.Now().UnixNano()), Data: cal}
	b.objects[path] = obj
	return &obj, nil
}
func (b *calDAVBackend) DeleteCalendarObject(_ context.Context, path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.objects, path)
	return nil
}
