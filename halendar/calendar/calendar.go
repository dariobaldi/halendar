// Package calendar reads and writes a CalDAV calendar (iCloud, Nextcloud, Radicale, ...).
//
//	cfg := calendar.ConfigFromEnv()
//	cal := calendar.New(cfg)
//	events, _ := cal.Events(ctx, time.Now(), time.Now().AddDate(0, 0, 7))
//	cal.Add(ctx, calendar.Event{Title: "Pitch", Start: start, End: end})
package calendar

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"sort"
	"strings"
	"sync"
	"time"
	_ "time/tzdata" // bundled timezone data

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"

	"halendar/envfile"
)

// Config describes one CalDAV account.
type Config struct {
	URL      string // "https://caldav.icloud.com/"
	User     string
	Pass     string // password (an app password for iCloud)
	Calendar string // default calendar to write to (empty = the first one)
	Timezone string // "Europe/Paris" by default
}

// ConfigFromEnv reads CALDAV_* from the environment (after envfile.Load(".env")).
func ConfigFromEnv() Config {
	return Config{
		URL:      envfile.String("CALDAV_URL", ""),
		User:     envfile.String("CALDAV_USER", ""),
		Pass:     envfile.String("CALDAV_PASS", ""),
		Calendar: envfile.String("CALDAV_CALENDAR", ""),
		Timezone: envfile.String("CALDAV_TIMEZONE", "Europe/Paris"),
	}
}

// Missing returns which required fields are still empty.
func (c Config) Missing() []string {
	required := map[string]string{"CALDAV_URL": c.URL, "CALDAV_USER": c.User, "CALDAV_PASS": c.Pass}
	var missing []string
	for key, value := range required {
		if value == "" {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}

// Event statuses (the iCalendar STATUS property).
const (
	Confirmed = "confirmed" // CONFIRMED
	Tentative = "tentative" // TENTATIVE
	Cancelled = "cancelled" // CANCELLED
)

// Event is a calendar booking. In JSON, dates accept "2026-09-18T09:00" (local time in the
// configured timezone), "2026-09-18" (all day), or RFC3339; "end" can be replaced by "duration_minutes".
type Event struct {
	UID         string    `json:"uid,omitempty"` // the same uid updates the event instead of duplicating it
	Title       string    `json:"title"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	AllDay      bool      `json:"all_day,omitempty"`
	Status      string    `json:"status,omitempty"` // confirmed (default) | tentative | cancelled
	Location    string    `json:"location,omitempty"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	Calendar    string    `json:"calendar,omitempty"`  // name of the calendar
	Recurring   bool      `json:"recurring,omitempty"` // read-only: an occurrence of a recurring series
}

// Client is a lazily connected, cached connection to a CalDAV account.
type Client struct {
	cfg Config
	loc *time.Location

	mu        sync.Mutex
	caldav    *caldav.Client
	calendars []caldav.Calendar
}

func New(cfg Config) *Client {
	loc, err := time.LoadLocation(cfg.Timezone)
	if cfg.Timezone == "" || err != nil {
		loc, _ = time.LoadLocation("Europe/Paris")
	}
	return &Client{cfg: cfg, loc: loc}
}

func (c *Client) Timezone() *time.Location { return c.loc }

// Test checks the connection.
func (c *Client) Test(ctx context.Context) error {
	if missing := c.cfg.Missing(); len(missing) > 0 {
		return fmt.Errorf("incomplete configuration: %s", strings.Join(missing, ", "))
	}
	_, _, err := c.connect(ctx)
	return err
}

// Calendars returns the names of the calendars that accept events.
func (c *Client) Calendars(ctx context.Context) ([]string, error) {
	_, calendars, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(calendars))
	for _, cal := range calendars {
		names = append(names, cal.Name)
	}
	return names, nil
}

// Events returns the events from every calendar that overlap [start, end], sorted by date.
// Recurring events are expanded into individual occurrences.
func (c *Client) Events(ctx context.Context, start, end time.Time) ([]Event, error) {
	client, calendars, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	var out []Event
	for _, cal := range calendars {
		// The query finds WHICH events fall in the period ...
		objects, err := client.QueryCalendar(ctx, cal.Path, &caldav.CalendarQuery{
			CompRequest: caldav.CalendarCompRequest{Name: "VCALENDAR", AllProps: true, AllComps: true},
			CompFilter: caldav.CompFilter{
				Name:  "VCALENDAR",
				Comps: []caldav.CompFilter{{Name: "VEVENT", Start: start.UTC(), End: end.UTC()}},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("CalDAV: reading %q: %w", cal.Name, err)
		}
		// ... then each full event is downloaded (iCloud otherwise returns empty events).
		for _, object := range objects {
			full, err := client.GetCalendarObject(ctx, object.Path)
			if err != nil || full.Data == nil {
				continue
			}
			for _, event := range full.Data.Events() {
				out = append(out, c.fromICal(event, cal.Name, start, end)...)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out, nil
}

// Busy reports whether [start, end] overlaps a non-cancelled event.
func (c *Client) Busy(ctx context.Context, start, end time.Time) (bool, []Event, error) {
	events, err := c.Events(ctx, start, end)
	if err != nil {
		return false, nil, err
	}
	var conflicts []Event
	for _, event := range events {
		if event.Status != Cancelled && event.Start.Before(end) && event.End.After(start) {
			conflicts = append(conflicts, event)
		}
	}
	return len(conflicts) > 0, conflicts, nil
}

// Add creates the event, or updates it if its UID already exists. Returns the event as written.
func (c *Client) Add(ctx context.Context, e Event) (Event, error) {
	if strings.TrimSpace(e.Title) == "" {
		return e, errors.New(`"title" field missing`)
	}
	if e.Start.IsZero() {
		return e, errors.New(`"start" field missing`)
	}
	if e.AllDay {
		day := e.Start.In(c.loc)
		e.Start = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, c.loc)
		if !e.End.After(e.Start) {
			e.End = e.Start.AddDate(0, 0, 1)
		}
	}
	if e.End.IsZero() {
		e.End = e.Start.Add(time.Hour)
	}
	if !e.End.After(e.Start) {
		return e, errors.New(`"end" must be after "start"`)
	}
	if e.Status == "" {
		e.Status = Confirmed
	}
	icalStatus := map[string]string{Confirmed: "CONFIRMED", Tentative: "TENTATIVE", Cancelled: "CANCELLED"}[e.Status]
	if icalStatus == "" {
		return e, fmt.Errorf("unknown status %q (confirmed, tentative, or cancelled)", e.Status)
	}
	if e.UID == "" {
		hash := sha1.Sum([]byte(e.Title + "|" + e.Start.UTC().Format(time.RFC3339)))
		e.UID = "evt-" + hex.EncodeToString(hash[:8])
	}

	client, _, err := c.connect(ctx)
	if err != nil {
		return e, err
	}
	target, err := c.pickCalendar(ctx, e.Calendar)
	if err != nil {
		return e, err
	}
	e.Calendar = target.Name

	event := ical.NewEvent()
	event.Props.SetText(ical.PropUID, e.UID)
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	event.Props.SetText(ical.PropSummary, e.Title)
	if e.AllDay {
		event.Props.SetDate(ical.PropDateTimeStart, e.Start)
		event.Props.SetDate(ical.PropDateTimeEnd, e.End)
	} else {
		event.Props.SetDateTime(ical.PropDateTimeStart, e.Start.UTC())
		event.Props.SetDateTime(ical.PropDateTimeEnd, e.End.UTC())
	}
	event.Props.SetText(ical.PropStatus, icalStatus)

	description := e.Description
	if e.URL != "" {
		event.Props.SetText(ical.PropURL, e.URL)
		if !strings.Contains(description, e.URL) { // not every app displays the URL property
			description = strings.TrimSpace(description + "\n" + e.URL)
		}
	}
	if e.Location != "" {
		event.Props.SetText(ical.PropLocation, e.Location)
	}
	if description != "" {
		event.Props.SetText(ical.PropDescription, description)
	}

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//halendar//EN")
	cal.Children = append(cal.Children, event.Component)

	if _, err := client.PutCalendarObject(ctx, target.Path+fileName(e.UID)+".ics", cal); err != nil {
		return e, fmt.Errorf("CalDAV: write refused: %w", err)
	}
	e.Start, e.End = e.Start.In(c.loc), e.End.In(c.loc)
	return e, nil
}

// Delete removes an event by UID (from the given calendar, or the default one).
func (c *Client) Delete(ctx context.Context, uid, calendarName string) error {
	client, _, err := c.connect(ctx)
	if err != nil {
		return err
	}
	target, err := c.pickCalendar(ctx, calendarName)
	if err != nil {
		return err
	}
	return client.RemoveAll(ctx, target.Path+fileName(uid)+".ics")
}

// ── JSON ────────────────────────────────────────────────────────────────────

// UnmarshalJSON accepts local dates ("2026-09-18T09:00") and "duration_minutes".
// The timezone used is "timezone" in the JSON, or Europe/Paris otherwise.
func (e *Event) UnmarshalJSON(b []byte) error {
	type alias Event
	var raw struct {
		alias
		Start           string `json:"start"`
		End             string `json:"end"`
		DurationMinutes int    `json:"duration_minutes"`
		Timezone        string `json:"timezone"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*e = Event(raw.alias)

	loc, _ := time.LoadLocation("Europe/Paris")
	if raw.Timezone != "" {
		l, err := time.LoadLocation(raw.Timezone)
		if err != nil {
			return fmt.Errorf("unknown timezone %q", raw.Timezone)
		}
		loc = l
	}

	var err error
	if raw.Start != "" {
		if e.Start, err = ParseDate(raw.Start, loc); err != nil {
			return fmt.Errorf(`"start": %w`, err)
		}
		if len(raw.Start) == len("2006-01-02") {
			e.AllDay = true
		}
	}
	switch {
	case raw.End != "":
		if e.End, err = ParseDate(raw.End, loc); err != nil {
			return fmt.Errorf(`"end": %w`, err)
		}
	case raw.DurationMinutes > 0 && !e.Start.IsZero():
		e.End = e.Start.Add(time.Duration(raw.DurationMinutes) * time.Minute)
	}
	return nil
}

// ParseDate accepts "2026-09-18T09:00", "2026-09-18 09:00", "2026-09-18", or RFC3339.
func ParseDate(s string, loc *time.Location) (time.Time, error) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.In(loc), nil
	}
	for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unreadable date %q (expected format: 2026-09-18T09:00)", s)
}

// ── internal ────────────────────────────────────────────────────────────────

func (c *Client) connect(ctx context.Context) (*caldav.Client, []caldav.Calendar, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.caldav != nil {
		return c.caldav, c.calendars, nil
	}
	if missing := c.cfg.Missing(); len(missing) > 0 {
		return nil, nil, fmt.Errorf("CalDAV not configured: %s", strings.Join(missing, ", "))
	}

	httpClient := webdav.HTTPClientWithBasicAuth(&http.Client{Timeout: 30 * time.Second}, c.cfg.User, c.cfg.Pass)
	client, err := caldav.NewClient(httpClient, c.cfg.URL)
	if err != nil {
		return nil, nil, err
	}
	principal, err := client.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("CalDAV: connection refused or incorrect URL: %w", err)
	}
	homeSet, err := client.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return nil, nil, fmt.Errorf("CalDAV: %w", err)
	}
	// iCloud hosts calendars on a different server (pXX-caldav.icloud.com)
	if host := homeSetHost(httpClient, c.cfg.URL, principal); host != "" {
		if client, err = caldav.NewClient(httpClient, host); err != nil {
			return nil, nil, err
		}
	}
	all, err := client.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, nil, fmt.Errorf("CalDAV: listing calendars: %w", err)
	}

	var eventCalendars []caldav.Calendar
	for _, cal := range all {
		if acceptsEvents(cal) {
			eventCalendars = append(eventCalendars, cal)
		}
	}
	c.caldav, c.calendars = client, eventCalendars
	return client, eventCalendars, nil
}

func (c *Client) fromICal(event ical.Event, calendarName string, start, end time.Time) []Event {
	base := Event{Calendar: calendarName, Status: Confirmed}
	base.UID, _ = event.Props.Text(ical.PropUID)
	base.Title, _ = event.Props.Text(ical.PropSummary)
	base.Location, _ = event.Props.Text(ical.PropLocation)
	base.Description, _ = event.Props.Text(ical.PropDescription)
	base.URL, _ = event.Props.Text(ical.PropURL)
	switch status, _ := event.Props.Text(ical.PropStatus); strings.ToUpper(status) {
	case "TENTATIVE":
		base.Status = Tentative
	case "CANCELLED":
		base.Status = Cancelled
	}
	if prop := event.Props.Get(ical.PropDateTimeStart); prop != nil && prop.ValueType() == ical.ValueDate {
		base.AllDay = true
	}

	start1, err1 := event.DateTimeStart(c.loc)
	end1, err2 := event.DateTimeEnd(c.loc)
	if err1 != nil || err2 != nil || start1.IsZero() {
		return nil
	}
	if end1.IsZero() || !end1.After(start1) {
		end1 = start1.Add(time.Hour)
	}
	duration := end1.Sub(start1)

	starts := []time.Time{start1}
	if rrule, err := event.RecurrenceSet(c.loc); err == nil && rrule != nil {
		starts = rrule.Between(start.Add(-duration), end, true)
		base.Recurring = true
	}
	var out []Event
	for _, s := range starts {
		occurrence := base
		occurrence.Start, occurrence.End = s.In(c.loc), s.Add(duration).In(c.loc)
		if occurrence.End.After(start) && occurrence.Start.Before(end) {
			out = append(out, occurrence)
		}
	}
	return out
}

func (c *Client) pickCalendar(ctx context.Context, name string) (*caldav.Calendar, error) {
	_, calendars, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = c.cfg.Calendar
	}
	for i, cal := range calendars {
		if name == "" || strings.EqualFold(cal.Name, name) {
			return &calendars[i], nil
		}
	}
	names, _ := c.Calendars(ctx)
	return nil, fmt.Errorf("calendar %q not found — available: %s", name, strings.Join(names, ", "))
}

func acceptsEvents(cal caldav.Calendar) bool {
	if len(cal.SupportedComponentSet) == 0 {
		return true
	}
	for _, comp := range cal.SupportedComponentSet {
		if comp == "VEVENT" {
			return true
		}
	}
	return false // e.g. "Reminders" (tasks only)
}

func fileName(uid string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, uid)
}

// homeSetHost returns "https://host/" if calendars are hosted on a different host (iCloud's case).
func homeSetHost(httpClient webdav.HTTPClient, base, principal string) string {
	baseURL, err := neturl.Parse(base)
	if err != nil {
		return ""
	}
	target := *baseURL
	target.Path = principal
	body := `<?xml version="1.0"?><d:propfind xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav"><d:prop><c:calendar-home-set/></d:prop></d:propfind>`
	req, _ := http.NewRequest("PROPFIND", target.String(), strings.NewReader(body))
	req.Header.Set("Depth", "0")
	req.Header.Set("Content-Type", "application/xml")
	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var multistatus struct {
		Hrefs []string `xml:"response>propstat>prop>calendar-home-set>href"`
	}
	if xml.Unmarshal(raw, &multistatus) != nil || len(multistatus.Hrefs) == 0 {
		return ""
	}
	host, err := neturl.Parse(strings.TrimSpace(multistatus.Hrefs[0]))
	if err != nil || host.Host == "" || host.Host == baseURL.Host {
		return ""
	}
	return host.Scheme + "://" + host.Host + "/"
}
