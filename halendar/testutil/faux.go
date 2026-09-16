// Package testutil fournit de faux serveurs (IMAP, SMTP, CalDAV) pour tester
// tout le parcours sans compte ni réseau. Utilisé par les tests d'intégration.
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

type FauxIMAP struct {
	Addr, User, Pass string
	user             *imapmemserver.User
}

func NouveauFauxIMAP(t *testing.T) *FauxIMAP {
	f := &FauxIMAP{User: "moi@test.local", Pass: "test"}
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

// Deposer ajoute un mail dans la boîte de réception.
func (f *FauxIMAP) Deposer(t *testing.T, id, de, sujet, texte string) {
	raw := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n",
		de, f.User, sujet, time.Now().Format(time.RFC1123Z), id, texte)
	if _, err := f.user.Append("INBOX", bytes.NewReader([]byte(raw)), &imap.AppendOptions{Time: time.Now()}); err != nil {
		t.Fatal(err)
	}
}

// Compter renvoie le nombre de messages d'un dossier.
func (f *FauxIMAP) Compter(dossier string) uint32 {
	st, err := f.user.Status(dossier, &imap.StatusOptions{NumMessages: true})
	if err != nil || st.NumMessages == nil {
		return 0
	}
	return *st.NumMessages
}

// ── SMTP ────────────────────────────────────────────────────────────────────

type FauxSMTP struct {
	Host  string
	Port  int
	mu    sync.Mutex
	Recus []MailRecu
}

type MailRecu struct {
	De   string
	A    []string
	Brut string
}

func NouveauFauxSMTP(t *testing.T) *FauxSMTP {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	f := &FauxSMTP{Host: "127.0.0.1", Port: ln.Addr().(*net.TCPAddr).Port}
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

func (f *FauxSMTP) session(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	ecrire := func(s string) { fmt.Fprint(conn, s+"\r\n") }
	ecrire("220 faux-smtp")
	var m MailRecu
	for {
		ligne, err := r.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(ligne))
		switch {
		case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
			ecrire("250-faux-smtp")
			ecrire("250 AUTH PLAIN")
		case strings.HasPrefix(cmd, "AUTH"):
			ecrire("235 ok")
		case strings.HasPrefix(cmd, "MAIL FROM"):
			m = MailRecu{De: strings.TrimSpace(ligne[10:])}
			ecrire("250 ok")
		case strings.HasPrefix(cmd, "RCPT TO"):
			m.A = append(m.A, strings.Trim(strings.TrimSpace(ligne[8:]), "<>"))
			ecrire("250 ok")
		case cmd == "DATA":
			ecrire("354 go")
			var b strings.Builder
			for {
				l, err := r.ReadString('\n')
				if err != nil || l == ".\r\n" {
					break
				}
				b.WriteString(l)
			}
			m.Brut = b.String()
			f.mu.Lock()
			f.Recus = append(f.Recus, m)
			f.mu.Unlock()
			ecrire("250 queued")
		case cmd == "QUIT":
			ecrire("221 bye")
			return
		default:
			ecrire("250 ok")
		}
	}
}

func (f *FauxSMTP) Mails() []MailRecu {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]MailRecu(nil), f.Recus...)
}

// ── CalDAV ──────────────────────────────────────────────────────────────────

type FauxCalDAV struct {
	URL, User, Pass string
	b               *backend
}

const cheminAgenda = "/moi/calendars/travail/"

func NouveauFauxCalDAV(t *testing.T) *FauxCalDAV {
	f := &FauxCalDAV{User: "moi", Pass: "test", b: &backend{objets: map[string]caldav.CalendarObject{}}}
	h := &caldav.Handler{Backend: f.b}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, p, ok := r.BasicAuth(); !ok || u != f.User || p != f.Pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="faux"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	f.URL = srv.URL + "/"
	return f
}

// Occuper ajoute une occupation existante.
func (f *FauxCalDAV) Occuper(uid, titre string, debut, fin time.Time) {
	ev := ical.NewEvent()
	ev.Props.SetText(ical.PropUID, uid)
	ev.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	ev.Props.SetText(ical.PropSummary, titre)
	ev.Props.SetDateTime(ical.PropDateTimeStart, debut.UTC())
	ev.Props.SetDateTime(ical.PropDateTimeEnd, fin.UTC())
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//test//FR")
	cal.Children = append(cal.Children, ev.Component)
	f.b.PutCalendarObject(context.Background(), cheminAgenda+uid+".ics", cal, nil)
}

// Evenement renvoie (titre, STATUS) d'un événement par uid, ok=false s'il n'existe pas.
func (f *FauxCalDAV) Evenement(uid string) (titre, statut string, ok bool) {
	f.b.mu.Lock()
	defer f.b.mu.Unlock()
	o, existe := f.b.objets[cheminAgenda+uid+".ics"]
	if !existe {
		return "", "", false
	}
	ev := o.Data.Events()[0]
	titre, _ = ev.Props.Text(ical.PropSummary)
	statut, _ = ev.Props.Text(ical.PropStatus)
	return titre, statut, true
}

type backend struct {
	mu     sync.Mutex
	objets map[string]caldav.CalendarObject
}

func (b *backend) CurrentUserPrincipal(context.Context) (string, error) { return "/moi/", nil }
func (b *backend) CalendarHomeSetPath(context.Context) (string, error)  { return "/moi/calendars/", nil }
func (b *backend) CreateCalendar(context.Context, *caldav.Calendar) error {
	return fmt.Errorf("non géré")
}
func (b *backend) ListCalendars(context.Context) ([]caldav.Calendar, error) {
	return []caldav.Calendar{{Path: cheminAgenda, Name: "Travail", SupportedComponentSet: []string{"VEVENT"}}}, nil
}
func (b *backend) GetCalendar(ctx context.Context, p string) (*caldav.Calendar, error) {
	c, _ := b.ListCalendars(ctx)
	if p == cheminAgenda {
		return &c[0], nil
	}
	return nil, fmt.Errorf("introuvable")
}
func (b *backend) GetCalendarObject(_ context.Context, p string, _ *caldav.CalendarCompRequest) (*caldav.CalendarObject, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if o, ok := b.objets[p]; ok {
		return &o, nil
	}
	return nil, fmt.Errorf("introuvable")
}
func (b *backend) ListCalendarObjects(_ context.Context, p string, _ *caldav.CalendarCompRequest) ([]caldav.CalendarObject, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []caldav.CalendarObject
	for k, o := range b.objets {
		if strings.HasPrefix(k, p) {
			out = append(out, o)
		}
	}
	return out, nil
}
func (b *backend) QueryCalendarObjects(ctx context.Context, p string, q *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	all, _ := b.ListCalendarObjects(ctx, p, nil)
	return caldav.Filter(q, all)
}
func (b *backend) PutCalendarObject(_ context.Context, p string, cal *ical.Calendar, _ *caldav.PutCalendarObjectOptions) (*caldav.CalendarObject, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	o := caldav.CalendarObject{Path: p, ModTime: time.Now(), ETag: fmt.Sprint(time.Now().UnixNano()), Data: cal}
	b.objets[p] = o
	return &o, nil
}
func (b *backend) DeleteCalendarObject(_ context.Context, p string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.objets, p)
	return nil
}
