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
	_ "github.com/emersion/go-message/charset" // accents, ISO-8859-1…
	gomail "github.com/emersion/go-message/mail"
)

// session ouvre une connexion IMAP authentifiée, sélectionne le dossier et exécute f.
func (b *Boite) session(dossier string, lectureSeule bool, f func(c *imapclient.Client, box *imap.SelectData) error) error {
	if b.cfg.IMAPHost == "" {
		return errors.New("IMAP non configuré (MAIL_IMAP_HOST)")
	}
	var c *imapclient.Client
	var err error
	if !b.cfg.IMAPSansTLS {
		c, err = imapclient.DialTLS(b.cfg.IMAPHost, nil)
	} else {
		c, err = imapclient.DialInsecure(b.cfg.IMAPHost, nil)
	}
	if err != nil {
		return fmt.Errorf("IMAP : connexion impossible à %s : %w", b.cfg.IMAPHost, err)
	}
	defer c.Close()
	if err := c.Login(b.cfg.User, b.cfg.Pass).Wait(); err != nil {
		return fmt.Errorf("IMAP : login refusé (mot de passe d'application ?) : %w", err)
	}
	var box *imap.SelectData
	if dossier != "" {
		if box, err = c.Select(dossier, &imap.SelectOptions{ReadOnly: lectureSeule}).Wait(); err != nil {
			return fmt.Errorf("IMAP : dossier %q : %w", dossier, err)
		}
	}
	if err := f(c, box); err != nil {
		return err
	}
	c.Logout().Wait()
	return nil
}

// Compter renvoie le nombre de messages du dossier.
func (b *Boite) Compter(ctx context.Context) (uint32, error) {
	var n uint32
	err := b.session(b.cfg.Dossier, true, func(c *imapclient.Client, box *imap.SelectData) error {
		n = box.NumMessages
		return nil
	})
	return n, err
}

// Dossiers liste les dossiers du compte (INBOX, Envoyés, Brouillons…).
func (b *Boite) Dossiers(ctx context.Context) ([]string, error) {
	var noms []string
	err := b.session("", true, func(c *imapclient.Client, _ *imap.SelectData) error {
		liste, err := c.List("", "*", nil).Collect()
		for _, d := range liste {
			noms = append(noms, d.Mailbox)
		}
		return err
	})
	return noms, err
}

// Derniers renvoie les n derniers messages, du plus récent au plus ancien.
func (b *Boite) Derniers(ctx context.Context, n int) ([]Message, error) {
	var msgs []Message
	err := b.session(b.cfg.Dossier, true, func(c *imapclient.Client, box *imap.SelectData) error {
		if box.NumMessages == 0 {
			return nil
		}
		debut := uint32(1)
		if box.NumMessages > uint32(n) {
			debut = box.NumMessages - uint32(n) + 1
		}
		var seq imap.SeqSet
		seq.AddRange(debut, box.NumMessages)
		var err error
		msgs, err = recuperer(c, seq)
		return err
	})
	return plusRecentsDabord(msgs), err
}

// Nouveaux renvoie les messages dont l'UID est > dernierUID, et le plus grand UID connu.
// Premier appel avec dernierUID = 0 : ne renvoie rien, mais donne le point de départ
// (pour ne pas traiter tout l'historique). Gardez l'UID renvoyé pour l'appel suivant.
func (b *Boite) Nouveaux(ctx context.Context, dernierUID uint32) ([]Message, uint32, error) {
	var msgs []Message
	max := dernierUID
	err := b.session(b.cfg.Dossier, true, func(c *imapclient.Client, box *imap.SelectData) error {
		if dernierUID == 0 {
			if box.UIDNext > 0 {
				max = uint32(box.UIDNext) - 1
			}
			return nil
		}
		if box.UIDNext != 0 && uint32(box.UIDNext) <= dernierUID+1 {
			return nil // rien de nouveau
		}
		var set imap.UIDSet
		set.AddRange(imap.UID(dernierUID+1), 0) // 0 = jusqu'au dernier
		tous, err := recuperer(c, set)
		for _, m := range tous {
			if m.UID > dernierUID { // le serveur renvoie toujours au moins le dernier
				msgs = append(msgs, m)
				if m.UID > max {
					max = m.UID
				}
			}
		}
		return err
	})
	return msgs, max, err
}

// Lire renvoie un message par UID.
func (b *Boite) Lire(ctx context.Context, uid uint32) (*Message, error) {
	var msg *Message
	err := b.session(b.cfg.Dossier, true, func(c *imapclient.Client, _ *imap.SelectData) error {
		msgs, err := recuperer(c, imap.UIDSetNum(imap.UID(uid)))
		if err != nil {
			return err
		}
		if len(msgs) == 0 {
			return fmt.Errorf("message UID %d introuvable", uid)
		}
		msg = &msgs[0]
		return nil
	})
	return msg, err
}

// Rechercher trouve les messages correspondant aux critères (les plus récents d'abord).
func (b *Boite) Rechercher(ctx context.Context, r Recherche) ([]Message, error) {
	crit := &imap.SearchCriteria{Since: r.Depuis, Before: r.Avant}
	if r.NonLus {
		crit.NotFlag = []imap.Flag{imap.FlagSeen}
	}
	if r.De != "" {
		crit.Header = append(crit.Header, imap.SearchCriteriaHeaderField{Key: "From", Value: r.De})
	}
	if r.Sujet != "" {
		crit.Header = append(crit.Header, imap.SearchCriteriaHeaderField{Key: "Subject", Value: r.Sujet})
	}
	if r.Contient != "" {
		crit.Body = []string{r.Contient}
	}
	max := r.Max
	if max <= 0 {
		max = 50
	}
	var msgs []Message
	err := b.session(b.cfg.Dossier, true, func(c *imapclient.Client, _ *imap.SelectData) error {
		res, err := c.UIDSearch(crit, nil).Wait()
		if err != nil {
			return fmt.Errorf("IMAP : recherche : %w", err)
		}
		uids := res.AllUIDs()
		if len(uids) == 0 {
			return nil
		}
		sort.Slice(uids, func(i, j int) bool { return uids[i] > uids[j] })
		if len(uids) > max {
			uids = uids[:max]
		}
		msgs, err = recuperer(c, imap.UIDSetNum(uids...))
		return err
	})
	return plusRecentsDabord(msgs), err
}

// MarquerLu marque des messages comme lus (lu=true) ou non lus (lu=false).
func (b *Boite) MarquerLu(ctx context.Context, lu bool, uids ...uint32) error {
	op := imap.StoreFlagsAdd
	if !lu {
		op = imap.StoreFlagsDel
	}
	return b.session(b.cfg.Dossier, false, func(c *imapclient.Client, _ *imap.SelectData) error {
		return c.Store(uidSet(uids), &imap.StoreFlags{Op: op, Silent: true, Flags: []imap.Flag{imap.FlagSeen}}, nil).Close()
	})
}

// Deplacer range des messages dans un autre dossier (ex : "Archives").
func (b *Boite) Deplacer(ctx context.Context, dossier string, uids ...uint32) error {
	return b.session(b.cfg.Dossier, false, func(c *imapclient.Client, _ *imap.SelectData) error {
		_, err := c.Move(uidSet(uids), dossier).Wait()
		return err
	})
}

// DeposerBrouillon enregistre le mail dans les Brouillons du compte au lieu de l'envoyer :
// l'utilisateur le relit et l'envoie depuis sa messagerie habituelle.
func (b *Boite) DeposerBrouillon(ctx context.Context, e Envoi) (dossier string, err error) {
	_, brut, err := b.construire(e)
	if err != nil {
		return "", err
	}
	err = b.session("", false, func(c *imapclient.Client, _ *imap.SelectData) error {
		dossier = dossierBrouillons(c)
		if dossier == "" {
			return errors.New("IMAP : dossier Brouillons introuvable")
		}
		cmd := c.Append(dossier, int64(len(brut)), &imap.AppendOptions{Flags: []imap.Flag{imap.FlagDraft, imap.FlagSeen}, Time: time.Now()})
		if _, err := cmd.Write(brut); err != nil {
			return err
		}
		if err := cmd.Close(); err != nil {
			return err
		}
		_, err := cmd.Wait()
		return err
	})
	return dossier, err
}

// ── interne ─────────────────────────────────────────────────────────────────

func dossierBrouillons(c *imapclient.Client) string {
	liste, _ := c.List("", "*", nil).Collect()
	for _, d := range liste {
		for _, a := range d.Attrs {
			if a == imap.MailboxAttrDrafts {
				return d.Mailbox
			}
		}
	}
	for _, nom := range []string{"Drafts", "Brouillons", "[Gmail]/Drafts", "[Gmail]/Brouillons", "INBOX.Drafts", "INBOX/Drafts"} {
		for _, d := range liste {
			if strings.EqualFold(d.Mailbox, nom) {
				return d.Mailbox
			}
		}
	}
	return ""
}

func uidSet(uids []uint32) imap.UIDSet {
	var s imap.UIDSet
	for _, u := range uids {
		s.AddNum(imap.UID(u))
	}
	return s
}

func recuperer(c *imapclient.Client, set imap.NumSet) ([]Message, error) {
	section := &imap.FetchItemBodySection{Peek: true} // Peek : ne marque pas comme lu
	bufs, err := c.Fetch(set, &imap.FetchOptions{
		UID: true, Flags: true, Envelope: true,
		BodySection: []*imap.FetchItemBodySection{section},
	}).Collect()
	if err != nil {
		return nil, fmt.Errorf("IMAP : lecture des messages : %w", err)
	}
	out := make([]Message, 0, len(bufs))
	for _, buf := range bufs {
		if buf.Envelope == nil {
			continue
		}
		env := buf.Envelope
		m := Message{
			UID:   uint32(buf.UID),
			ID:    Crochets(env.MessageID), // la bibliothèque retire les < >
			Sujet: env.Subject,
			Date:  env.Date,
		}
		if len(env.From) > 0 {
			m.De, m.DeNom = env.From[0].Addr(), env.From[0].Name
		}
		if len(env.ReplyTo) > 0 && env.ReplyTo[0].Addr() != "" {
			m.De = env.ReplyTo[0].Addr()
		}
		for _, a := range env.To {
			m.A = append(m.A, a.Addr())
		}
		for _, a := range env.Cc {
			m.Cc = append(m.Cc, a.Addr())
		}
		for _, f := range buf.Flags {
			if f == imap.FlagSeen {
				m.Lu = true
			}
		}
		if m.ID == "" {
			m.ID = fmt.Sprintf("<uid-%d@imap>", buf.UID)
		}
		analyserCorps(&m, buf.FindBodySection(section))
		out = append(out, m)
	}
	return out, nil
}

func analyserCorps(m *Message, raw []byte) {
	r, err := gomail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		return
	}
	m.References = r.Header.Get("References")
	if m.Date.IsZero() {
		m.Date, _ = r.Header.Date()
	}
	for {
		p, err := r.NextPart()
		if errors.Is(err, io.EOF) || err != nil {
			break
		}
		switch h := p.Header.(type) {
		case *gomail.InlineHeader:
			ct, _, _ := h.ContentType()
			contenu, _ := io.ReadAll(io.LimitReader(p.Body, 2<<20))
			switch {
			case ct == "text/plain" && m.Texte == "":
				m.Texte = strings.TrimSpace(string(contenu))
			case ct == "text/html" && m.HTML == "":
				m.HTML = string(contenu)
			}
		case *gomail.AttachmentHeader:
			nom, _ := h.Filename()
			ct, _, _ := h.ContentType()
			n, _ := io.Copy(io.Discard, p.Body)
			m.PiecesJointes = append(m.PiecesJointes, PieceJointe{Nom: nom, Type: ct, Taille: int(n)})
		}
	}
	if m.Texte == "" && m.HTML != "" {
		m.Texte = HTMLVersTexte(m.HTML)
	}
}

// HTMLVersTexte : nettoyage simple d'un corps HTML.
func HTMLVersTexte(h string) string {
	var sb strings.Builder
	dansBalise := false
	for _, r := range strings.NewReplacer("<br>", "\n", "<br/>", "\n", "<br />", "\n", "</p>", "\n", "</div>", "\n", "</tr>", "\n").Replace(h) {
		switch {
		case r == '<':
			dansBalise = true
		case r == '>':
			dansBalise = false
		case !dansBalise:
			sb.WriteRune(r)
		}
	}
	t := strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&#39;", "'", "&quot;", `"`).Replace(sb.String())
	var lignes []string
	for _, l := range strings.Split(t, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lignes = append(lignes, l)
		}
	}
	return strings.Join(lignes, "\n")
}

func plusRecentsDabord(msgs []Message) []Message {
	sort.SliceStable(msgs, func(i, j int) bool { return msgs[i].UID > msgs[j].UID })
	return msgs
}
