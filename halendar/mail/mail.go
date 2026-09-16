// Package mail : lire (IMAP) et envoyer (SMTP) des mails.
//
//	cfg := mail.ConfigDepuisEnv()          // ou remplir mail.Config à la main
//	boite := mail.Nouveau(cfg)
//	msgs, _ := boite.Derniers(ctx, 10)
//	boite.Repondre(ctx, msgs[0], "Merci, c'est noté.")
package mail

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"halendar/envfile"
)

// Config : un compte mail. Les champs SMTP sont déduits de l'IMAP s'ils sont vides.
type Config struct {
	IMAPHost    string // "imap.gmail.com:993"
	IMAPSansTLS bool   // true seulement pour un serveur local de test
	Dossier     string // "INBOX" par défaut

	SMTPHost string // "smtp.gmail.com" (déduit de IMAPHost si vide)
	SMTPPort int    // 587 (STARTTLS) ou 465 (TLS direct)

	User string // adresse du compte
	Pass string // mot de passe (d'application pour Gmail, iCloud, Yahoo…)
	From string // expéditeur affiché, ex : "Équipe Halendar <moi@gmail.com>" (défaut : User)
}

// ConfigDepuisEnv lit MAIL_* dans l'environnement (après envfile.Charger(".env")).
func ConfigDepuisEnv() Config {
	c := Config{
		IMAPHost:    envfile.Texte("MAIL_IMAP_HOST", ""),
		IMAPSansTLS: !envfile.Booleen("MAIL_IMAP_TLS", true),
		Dossier:     envfile.Texte("MAIL_DOSSIER", "INBOX"),
		SMTPHost:    envfile.Texte("MAIL_SMTP_HOST", ""),
		SMTPPort:    envfile.Entier("MAIL_SMTP_PORT", 0),
		User:        envfile.Texte("MAIL_USER", ""),
		Pass:        envfile.Texte("MAIL_PASS", ""),
		From:        envfile.Texte("MAIL_FROM", ""),
	}
	return c.completer()
}

// Manquant renvoie les champs obligatoires vides (liste vide = configuration OK).
func (c Config) Manquant() []string {
	var m []string
	if c.IMAPHost == "" && c.SMTPHost == "" {
		m = append(m, "MAIL_IMAP_HOST")
	}
	if c.User == "" {
		m = append(m, "MAIL_USER")
	}
	if c.Pass == "" {
		m = append(m, "MAIL_PASS")
	}
	return m
}

func (c Config) completer() Config {
	if c.Dossier == "" {
		c.Dossier = "INBOX"
	}
	if c.SMTPHost == "" && strings.HasPrefix(c.IMAPHost, "imap.") {
		h := strings.TrimPrefix(c.IMAPHost, "imap.")
		if i := strings.LastIndex(h, ":"); i >= 0 {
			h = h[:i]
		}
		c.SMTPHost = "smtp." + h
	}
	if i := strings.LastIndex(c.SMTPHost, ":"); i >= 0 { // "smtp.x.com:465" accepté
		if p, err := strconv.Atoi(c.SMTPHost[i+1:]); err == nil {
			c.SMTPPort, c.SMTPHost = p, c.SMTPHost[:i]
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

// Message : un mail reçu.
type Message struct {
	UID           uint32        `json:"uid"`
	ID            string        `json:"id"` // Message-ID, ex : <abc@exemple.fr>
	De            string        `json:"de"` // adresse à qui répondre
	DeNom         string        `json:"de_nom,omitempty"`
	A             []string      `json:"a,omitempty"`
	Cc            []string      `json:"cc,omitempty"`
	Sujet         string        `json:"sujet"`
	Date          time.Time     `json:"date"`
	Texte         string        `json:"texte"`          // partie texte (ou HTML nettoyé)
	HTML          string        `json:"html,omitempty"` // partie HTML brute si présente
	References    string        `json:"references,omitempty"`
	Lu            bool          `json:"lu"`
	PiecesJointes []PieceJointe `json:"pieces_jointes,omitempty"`
}

// PieceJointe : description (sans le contenu) d'une pièce jointe.
type PieceJointe struct {
	Nom    string `json:"nom"`
	Type   string `json:"type"`
	Taille int    `json:"taille"`
}

// Envoi : un mail à envoyer.
type Envoi struct {
	A          []string `json:"a"`
	Cc         []string `json:"cc,omitempty"`
	Cci        []string `json:"cci,omitempty"`
	Sujet      string   `json:"sujet"`
	Texte      string   `json:"texte"`
	HTML       string   `json:"html,omitempty"`         // optionnel : version HTML en plus du texte
	EnReponseA string   `json:"en_reponse_a,omitempty"` // Message-ID du mail d'origine
	References string   `json:"references,omitempty"`
}

// Recherche : critères combinés (tous optionnels).
type Recherche struct {
	NonLus   bool      `json:"non_lus,omitempty"`
	Depuis   time.Time `json:"depuis,omitempty"`
	Avant    time.Time `json:"avant,omitempty"`
	De       string    `json:"de,omitempty"`
	Sujet    string    `json:"sujet,omitempty"`
	Contient string    `json:"contient,omitempty"` // dans le corps
	Max      int       `json:"max,omitempty"`      // les plus récents d'abord ; 0 = 50
}

// Boite regroupe la lecture et l'envoi pour un compte.
type Boite struct {
	cfg Config
}

func Nouveau(cfg Config) *Boite { return &Boite{cfg: cfg.completer()} }

func (b *Boite) Config() Config { return b.cfg }

// Tester vérifie la connexion IMAP et SMTP sans rien modifier.
func (b *Boite) Tester(ctx context.Context) error {
	if m := b.cfg.Manquant(); len(m) > 0 {
		return fmt.Errorf("configuration incomplète : %s", strings.Join(m, ", "))
	}
	var erreurs []string
	if b.cfg.IMAPHost != "" {
		if _, err := b.Compter(ctx); err != nil {
			erreurs = append(erreurs, err.Error())
		}
	}
	if b.cfg.SMTPHost != "" {
		c, err := b.smtpClient()
		if err != nil {
			erreurs = append(erreurs, err.Error())
		} else {
			c.Quit()
			c.Close()
		}
	}
	if len(erreurs) > 0 {
		return fmt.Errorf("%s", strings.Join(erreurs, " ; "))
	}
	return nil
}

// Crochets normalise un Message-ID au format <id@domaine> (nécessaire pour rester dans le fil).
func Crochets(id string) string {
	id = strings.TrimSpace(id)
	if id == "" || strings.HasPrefix(id, "<") {
		return id
	}
	return "<" + id + ">"
}
