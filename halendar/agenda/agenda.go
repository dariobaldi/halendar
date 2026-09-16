// Package agenda : lire et écrire dans un agenda CalDAV (iCloud, Nextcloud, Radicale…).
//
//	cfg := agenda.ConfigDepuisEnv()
//	ag := agenda.Nouveau(cfg)
//	evs, _ := ag.Evenements(ctx, time.Now(), time.Now().AddDate(0, 0, 7))
//	ag.Ajouter(ctx, agenda.Evenement{Titre: "Pitch", Debut: debut, Fin: fin})
package agenda

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
	_ "time/tzdata" // fuseaux embarqués

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"

	"halendar/envfile"
)

// Config : un compte CalDAV.
type Config struct {
	URL    string // "https://caldav.icloud.com/"
	User   string
	Pass   string // mot de passe (d'application pour iCloud)
	Agenda string // agenda par défaut pour écrire (vide = le premier)
	Fuseau string // "Europe/Paris" par défaut
}

// ConfigDepuisEnv lit CALDAV_* dans l'environnement (après envfile.Charger(".env")).
func ConfigDepuisEnv() Config {
	return Config{
		URL:    envfile.Texte("CALDAV_URL", ""),
		User:   envfile.Texte("CALDAV_USER", ""),
		Pass:   envfile.Texte("CALDAV_PASS", ""),
		Agenda: envfile.Texte("CALDAV_AGENDA", ""),
		Fuseau: envfile.Texte("CALDAV_FUSEAU", "Europe/Paris"),
	}
}

// Manquant renvoie les champs obligatoires vides.
func (c Config) Manquant() []string {
	var m []string
	for k, v := range map[string]string{"CALDAV_URL": c.URL, "CALDAV_USER": c.User, "CALDAV_PASS": c.Pass} {
		if v == "" {
			m = append(m, k)
		}
	}
	sort.Strings(m)
	return m
}

// Statuts d'un événement (iCalendar STATUS).
const (
	Confirme   = "confirme"   // CONFIRMED
	Provisoire = "provisoire" // TENTATIVE
	Annule     = "annule"     // CANCELLED
)

// Evenement : un rendez-vous. En JSON, les dates acceptent "2026-09-18T09:00" (heure locale
// du fuseau), "2026-09-18" (journée entière) ou RFC3339 ; "fin" peut être remplacée par "duree_minutes".
type Evenement struct {
	UID            string    `json:"uid,omitempty"` // même uid = mise à jour au lieu d'un doublon
	Titre          string    `json:"titre"`
	Debut          time.Time `json:"debut"`
	Fin            time.Time `json:"fin"`
	JourneeEntiere bool      `json:"journee_entiere,omitempty"`
	Statut         string    `json:"statut,omitempty"` // confirme (défaut) | provisoire | annule
	Lieu           string    `json:"lieu,omitempty"`
	Description    string    `json:"description,omitempty"`
	Lien           string    `json:"lien,omitempty"`
	Agenda         string    `json:"agenda,omitempty"`    // nom de l'agenda
	Recurrent      bool      `json:"recurrent,omitempty"` // lecture seule : occurrence d'une série
}

// Agenda : connexion (paresseuse, mise en cache) à un compte CalDAV.
type Agenda struct {
	cfg Config
	loc *time.Location

	mu      sync.Mutex
	client  *caldav.Client
	agendas []caldav.Calendar
}

func Nouveau(cfg Config) *Agenda {
	loc, err := time.LoadLocation(cfg.Fuseau)
	if cfg.Fuseau == "" || err != nil {
		loc, _ = time.LoadLocation("Europe/Paris")
	}
	return &Agenda{cfg: cfg, loc: loc}
}

func (a *Agenda) Fuseau() *time.Location { return a.loc }

// Tester vérifie la connexion.
func (a *Agenda) Tester(ctx context.Context) error {
	if m := a.cfg.Manquant(); len(m) > 0 {
		return fmt.Errorf("configuration incomplète : %s", strings.Join(m, ", "))
	}
	_, _, err := a.connexion(ctx)
	return err
}

// Agendas renvoie les noms des agendas qui acceptent des événements.
func (a *Agenda) Agendas(ctx context.Context) ([]string, error) {
	_, cals, err := a.connexion(ctx)
	if err != nil {
		return nil, err
	}
	noms := make([]string, 0, len(cals))
	for _, c := range cals {
		noms = append(noms, c.Name)
	}
	return noms, nil
}

// Evenements renvoie les événements de tous les agendas qui chevauchent [debut, fin],
// triés par date. Les événements récurrents sont dépliés en occurrences.
func (a *Agenda) Evenements(ctx context.Context, debut, fin time.Time) ([]Evenement, error) {
	c, cals, err := a.connexion(ctx)
	if err != nil {
		return nil, err
	}
	var out []Evenement
	for _, cal := range cals {
		// La requête trouve QUELS événements tombent dans la période…
		objs, err := c.QueryCalendar(ctx, cal.Path, &caldav.CalendarQuery{
			CompRequest: caldav.CalendarCompRequest{Name: "VCALENDAR", AllProps: true, AllComps: true},
			CompFilter: caldav.CompFilter{
				Name:  "VCALENDAR",
				Comps: []caldav.CompFilter{{Name: "VEVENT", Start: debut.UTC(), End: fin.UTC()}},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("CalDAV : lecture de %q : %w", cal.Name, err)
		}
		// …puis on télécharge chaque événement complet (iCloud renvoie des événements vides sinon).
		for _, o := range objs {
			full, err := c.GetCalendarObject(ctx, o.Path)
			if err != nil || full.Data == nil {
				continue
			}
			for _, ev := range full.Data.Events() {
				out = append(out, a.lire(ev, cal.Name, debut, fin)...)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Debut.Before(out[j].Debut) })
	return out, nil
}

// Occupe indique si [debut, fin] chevauche un événement non annulé.
func (a *Agenda) Occupe(ctx context.Context, debut, fin time.Time) (bool, []Evenement, error) {
	evs, err := a.Evenements(ctx, debut, fin)
	if err != nil {
		return false, nil, err
	}
	var conflits []Evenement
	for _, e := range evs {
		if e.Statut != Annule && e.Debut.Before(fin) && e.Fin.After(debut) {
			conflits = append(conflits, e)
		}
	}
	return len(conflits) > 0, conflits, nil
}

// Ajouter crée l'événement, ou le met à jour si son UID existe déjà. Renvoie l'événement écrit.
func (a *Agenda) Ajouter(ctx context.Context, e Evenement) (Evenement, error) {
	if strings.TrimSpace(e.Titre) == "" {
		return e, errors.New(`champ "titre" manquant`)
	}
	if e.Debut.IsZero() {
		return e, errors.New(`champ "debut" manquant`)
	}
	if e.JourneeEntiere {
		d := e.Debut.In(a.loc)
		e.Debut = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, a.loc)
		if !e.Fin.After(e.Debut) {
			e.Fin = e.Debut.AddDate(0, 0, 1)
		}
	}
	if e.Fin.IsZero() {
		e.Fin = e.Debut.Add(time.Hour)
	}
	if !e.Fin.After(e.Debut) {
		return e, errors.New(`"fin" doit être après "debut"`)
	}
	if e.Statut == "" {
		e.Statut = Confirme
	}
	statutICal := map[string]string{Confirme: "CONFIRMED", Provisoire: "TENTATIVE", Annule: "CANCELLED"}[e.Statut]
	if statutICal == "" {
		return e, fmt.Errorf("statut %q inconnu (confirme, provisoire ou annule)", e.Statut)
	}
	if e.UID == "" {
		h := sha1.Sum([]byte(e.Titre + "|" + e.Debut.UTC().Format(time.RFC3339)))
		e.UID = "evt-" + hex.EncodeToString(h[:8])
	}

	c, _, err := a.connexion(ctx)
	if err != nil {
		return e, err
	}
	cible, err := a.choisir(ctx, e.Agenda)
	if err != nil {
		return e, err
	}
	e.Agenda = cible.Name

	ev := ical.NewEvent()
	ev.Props.SetText(ical.PropUID, e.UID)
	ev.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	ev.Props.SetText(ical.PropSummary, e.Titre)
	if e.JourneeEntiere {
		ev.Props.SetDate(ical.PropDateTimeStart, e.Debut)
		ev.Props.SetDate(ical.PropDateTimeEnd, e.Fin)
	} else {
		ev.Props.SetDateTime(ical.PropDateTimeStart, e.Debut.UTC())
		ev.Props.SetDateTime(ical.PropDateTimeEnd, e.Fin.UTC())
	}
	ev.Props.SetText(ical.PropStatus, statutICal)
	description := e.Description
	if e.Lien != "" {
		ev.Props.SetText(ical.PropURL, e.Lien)
		if !strings.Contains(description, e.Lien) { // toutes les apps n'affichent pas URL
			description = strings.TrimSpace(description + "\n" + e.Lien)
		}
	}
	if e.Lieu != "" {
		ev.Props.SetText(ical.PropLocation, e.Lieu)
	}
	if description != "" {
		ev.Props.SetText(ical.PropDescription, description)
	}
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//halendar//FR")
	cal.Children = append(cal.Children, ev.Component)

	if _, err := c.PutCalendarObject(ctx, cible.Path+nomFichier(e.UID)+".ics", cal); err != nil {
		return e, fmt.Errorf("CalDAV : écriture refusée : %w", err)
	}
	e.Debut, e.Fin = e.Debut.In(a.loc), e.Fin.In(a.loc)
	return e, nil
}

// Supprimer retire un événement par UID (dans l'agenda indiqué, sinon l'agenda par défaut).
func (a *Agenda) Supprimer(ctx context.Context, uid, nomAgenda string) error {
	c, _, err := a.connexion(ctx)
	if err != nil {
		return err
	}
	cible, err := a.choisir(ctx, nomAgenda)
	if err != nil {
		return err
	}
	return c.RemoveAll(ctx, cible.Path+nomFichier(uid)+".ics")
}

// ── JSON ────────────────────────────────────────────────────────────────────

// UnmarshalJSON accepte des dates locales ("2026-09-18T09:00") et "duree_minutes".
// Le fuseau utilisé est "fuseau" dans le JSON, sinon Europe/Paris.
func (e *Evenement) UnmarshalJSON(b []byte) error {
	type alias Evenement
	var brut struct {
		alias
		Debut        string `json:"debut"`
		Fin          string `json:"fin"`
		DureeMinutes int    `json:"duree_minutes"`
		Fuseau       string `json:"fuseau"`
	}
	if err := json.Unmarshal(b, &brut); err != nil {
		return err
	}
	*e = Evenement(brut.alias)
	loc, _ := time.LoadLocation("Europe/Paris")
	if brut.Fuseau != "" {
		l, err := time.LoadLocation(brut.Fuseau)
		if err != nil {
			return fmt.Errorf("fuseau inconnu %q", brut.Fuseau)
		}
		loc = l
	}
	var err error
	if brut.Debut != "" {
		if e.Debut, err = ParseDate(brut.Debut, loc); err != nil {
			return fmt.Errorf(`"debut" : %w`, err)
		}
		if len(brut.Debut) == len("2006-01-02") {
			e.JourneeEntiere = true
		}
	}
	switch {
	case brut.Fin != "":
		if e.Fin, err = ParseDate(brut.Fin, loc); err != nil {
			return fmt.Errorf(`"fin" : %w`, err)
		}
	case brut.DureeMinutes > 0 && !e.Debut.IsZero():
		e.Fin = e.Debut.Add(time.Duration(brut.DureeMinutes) * time.Minute)
	}
	return nil
}

// ParseDate accepte "2026-09-18T09:00", "2026-09-18 09:00", "2026-09-18" ou RFC3339.
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
	return time.Time{}, fmt.Errorf("date illisible %q (format attendu : 2026-09-18T09:00)", s)
}

// ── interne ─────────────────────────────────────────────────────────────────

func (a *Agenda) connexion(ctx context.Context) (*caldav.Client, []caldav.Calendar, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.client != nil {
		return a.client, a.agendas, nil
	}
	if m := a.cfg.Manquant(); len(m) > 0 {
		return nil, nil, fmt.Errorf("CalDAV non configuré : %s", strings.Join(m, ", "))
	}
	httpc := webdav.HTTPClientWithBasicAuth(&http.Client{Timeout: 30 * time.Second}, a.cfg.User, a.cfg.Pass)
	c, err := caldav.NewClient(httpc, a.cfg.URL)
	if err != nil {
		return nil, nil, err
	}
	principal, err := c.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("CalDAV : connexion refusée ou URL incorrecte : %w", err)
	}
	homeSet, err := c.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return nil, nil, fmt.Errorf("CalDAV : %w", err)
	}
	// iCloud héberge les agendas sur un autre serveur (pXX-caldav.icloud.com)
	if host := homeSetHost(httpc, a.cfg.URL, principal); host != "" {
		if c, err = caldav.NewClient(httpc, host); err != nil {
			return nil, nil, err
		}
	}
	cals, err := c.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, nil, fmt.Errorf("CalDAV : liste des agendas : %w", err)
	}
	var evenements []caldav.Calendar
	for _, cal := range cals {
		if accepteEvenements(cal) {
			evenements = append(evenements, cal)
		}
	}
	a.client, a.agendas = c, evenements
	return c, evenements, nil
}

func (a *Agenda) lire(ev ical.Event, nomAgenda string, debut, fin time.Time) []Evenement {
	base := Evenement{Agenda: nomAgenda, Statut: Confirme}
	base.UID, _ = ev.Props.Text(ical.PropUID)
	base.Titre, _ = ev.Props.Text(ical.PropSummary)
	base.Lieu, _ = ev.Props.Text(ical.PropLocation)
	base.Description, _ = ev.Props.Text(ical.PropDescription)
	base.Lien, _ = ev.Props.Text(ical.PropURL)
	switch s, _ := ev.Props.Text(ical.PropStatus); strings.ToUpper(s) {
	case "TENTATIVE":
		base.Statut = Provisoire
	case "CANCELLED":
		base.Statut = Annule
	}
	if p := ev.Props.Get(ical.PropDateTimeStart); p != nil && p.ValueType() == ical.ValueDate {
		base.JourneeEntiere = true
	}
	d, err1 := ev.DateTimeStart(a.loc)
	f, err2 := ev.DateTimeEnd(a.loc)
	if err1 != nil || err2 != nil || d.IsZero() {
		return nil
	}
	if f.IsZero() || !f.After(d) {
		f = d.Add(time.Hour)
	}
	duree := f.Sub(d)

	debuts := []time.Time{d}
	if rs, err := ev.RecurrenceSet(a.loc); err == nil && rs != nil {
		debuts = rs.Between(debut.Add(-duree), fin, true)
		base.Recurrent = true
	}
	var out []Evenement
	for _, x := range debuts {
		e := base
		e.Debut, e.Fin = x.In(a.loc), x.Add(duree).In(a.loc)
		if e.Fin.After(debut) && e.Debut.Before(fin) {
			out = append(out, e)
		}
	}
	return out
}

func (a *Agenda) choisir(ctx context.Context, nom string) (*caldav.Calendar, error) {
	_, cals, err := a.connexion(ctx)
	if err != nil {
		return nil, err
	}
	if nom == "" {
		nom = a.cfg.Agenda
	}
	for i, cal := range cals {
		if nom == "" || strings.EqualFold(cal.Name, nom) {
			return &cals[i], nil
		}
	}
	noms, _ := a.Agendas(ctx)
	return nil, fmt.Errorf("agenda %q introuvable — disponibles : %s", nom, strings.Join(noms, ", "))
}

func accepteEvenements(c caldav.Calendar) bool {
	if len(c.SupportedComponentSet) == 0 {
		return true
	}
	for _, comp := range c.SupportedComponentSet {
		if comp == "VEVENT" {
			return true
		}
	}
	return false // ex : « Rappels » (tâches uniquement)
}

func nomFichier(uid string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, uid)
}

// homeSetHost renvoie "https://hôte/" si les agendas sont sur un autre hôte (cas d'iCloud).
func homeSetHost(httpc webdav.HTTPClient, base, principal string) string {
	b, err := neturl.Parse(base)
	if err != nil {
		return ""
	}
	u := *b
	u.Path = principal
	body := `<?xml version="1.0"?><d:propfind xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav"><d:prop><c:calendar-home-set/></d:prop></d:propfind>`
	req, _ := http.NewRequest("PROPFIND", u.String(), strings.NewReader(body))
	req.Header.Set("Depth", "0")
	req.Header.Set("Content-Type", "application/xml")
	resp, err := httpc.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var ms struct {
		Hrefs []string `xml:"response>propstat>prop>calendar-home-set>href"`
	}
	if xml.Unmarshal(raw, &ms) != nil || len(ms.Hrefs) == 0 {
		return ""
	}
	h, err := neturl.Parse(strings.TrimSpace(ms.Hrefs[0]))
	if err != nil || h.Host == "" || h.Host == b.Host {
		return ""
	}
	return h.Scheme + "://" + h.Host + "/"
}
