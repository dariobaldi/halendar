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
	"strings"
	"time"
)

// Envoyer envoie le mail et renvoie son Message-ID.
func (b *Boite) Envoyer(ctx context.Context, e Envoi) (string, error) {
	id, brut, err := b.construire(e)
	if err != nil {
		return "", err
	}
	c, err := b.smtpClient()
	if err != nil {
		return "", err
	}
	defer c.Close()
	if err := c.Mail(adresse(b.cfg.From)); err != nil {
		return "", fmt.Errorf("SMTP expéditeur refusé : %w", err)
	}
	for _, dest := range append(append(append([]string{}, e.A...), e.Cc...), e.Cci...) {
		if err := c.Rcpt(adresse(dest)); err != nil {
			return "", fmt.Errorf("SMTP destinataire %q refusé : %w", dest, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return "", err
	}
	if _, err := w.Write(brut); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("SMTP envoi refusé : %w", err)
	}
	c.Quit()
	return id, nil
}

// Repondre répond à un message reçu, dans le même fil de discussion.
func (b *Boite) Repondre(ctx context.Context, original Message, texte string) (string, error) {
	return b.Envoyer(ctx, ReponseA(original, texte))
}

// ReponseA prépare (sans l'envoyer) la réponse à un message : destinataire, "Re:" et fil.
// Utile pour DeposerBrouillon ou pour modifier avant envoi.
func ReponseA(original Message, texte string) Envoi {
	sujet := original.Sujet
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(sujet)), "re:") {
		sujet = "Re: " + sujet
	}
	return Envoi{A: []string{original.De}, Sujet: sujet, Texte: texte, EnReponseA: original.ID, References: original.References}
}

func (b *Boite) smtpClient() (*smtp.Client, error) {
	if b.cfg.SMTPHost == "" {
		return nil, errors.New("SMTP non configuré (MAIL_SMTP_HOST)")
	}
	addr := fmt.Sprintf("%s:%d", b.cfg.SMTPHost, b.cfg.SMTPPort)
	tlsCfg := &tls.Config{ServerName: b.cfg.SMTPHost}
	var conn net.Conn
	var err error
	if b.cfg.SMTPPort == 465 {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: 15 * time.Second}, "tcp", addr, tlsCfg)
	} else {
		conn, err = net.DialTimeout("tcp", addr, 15*time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("SMTP : connexion impossible à %s : %w", addr, err)
	}
	c, err := smtp.NewClient(conn, b.cfg.SMTPHost)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if ok, _ := c.Extension("STARTTLS"); ok && b.cfg.SMTPPort != 465 {
		if err := c.StartTLS(tlsCfg); err != nil {
			c.Close()
			return nil, fmt.Errorf("SMTP STARTTLS : %w", err)
		}
	}
	if ok, _ := c.Extension("AUTH"); ok {
		if err := c.Auth(smtp.PlainAuth("", b.cfg.User, b.cfg.Pass, b.cfg.SMTPHost)); err != nil {
			c.Close()
			return nil, fmt.Errorf("SMTP : authentification refusée (mot de passe d'application ?) : %w", err)
		}
	}
	return c, nil
}

// construire produit le mail brut (RFC 5322) : texte seul, ou texte + HTML.
func (b *Boite) construire(e Envoi) (string, []byte, error) {
	if len(e.A) == 0 || strings.TrimSpace(e.A[0]) == "" {
		return "", nil, errors.New(`champ "a" (destinataire) manquant`)
	}
	if strings.TrimSpace(e.Texte) == "" && strings.TrimSpace(e.HTML) == "" {
		return "", nil, errors.New(`champ "texte" manquant`)
	}
	domaine := "halendar.local"
	if i := strings.LastIndex(b.cfg.From, "@"); i >= 0 {
		domaine = strings.Trim(b.cfg.From[i+1:], "> ")
	}
	rnd := make([]byte, 8)
	rand.Read(rnd)
	id := fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), hex.EncodeToString(rnd), domaine)

	var buf bytes.Buffer
	h := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
	}
	h("From", b.cfg.From)
	h("To", strings.Join(e.A, ", "))
	h("Cc", strings.Join(e.Cc, ", "))
	h("Subject", mime.QEncoding.Encode("utf-8", e.Sujet))
	h("Date", time.Now().Format(time.RFC1123Z))
	h("Message-ID", id)
	h("In-Reply-To", Crochets(e.EnReponseA))
	var refs []string
	for _, r := range strings.Fields(e.References + " " + e.EnReponseA) {
		if r = Crochets(r); !contient(refs, r) {
			refs = append(refs, r)
		}
	}
	h("References", strings.Join(refs, " "))
	h("MIME-Version", "1.0")

	partie := func(contentType, corps string) {
		fmt.Fprintf(&buf, "Content-Type: %s; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", contentType)
		qp := quotedprintable.NewWriter(&buf)
		qp.Write([]byte(strings.ReplaceAll(strings.ReplaceAll(corps, "\r\n", "\n"), "\n", "\r\n")))
		qp.Close()
		buf.WriteString("\r\n")
	}
	if e.HTML == "" {
		partie("text/plain", e.Texte)
	} else {
		texte := e.Texte
		if texte == "" {
			texte = HTMLVersTexte(e.HTML)
		}
		frontiere := "halendar-" + hex.EncodeToString(rnd)
		fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", frontiere)
		fmt.Fprintf(&buf, "--%s\r\n", frontiere)
		partie("text/plain", texte)
		fmt.Fprintf(&buf, "--%s\r\n", frontiere)
		partie("text/html", e.HTML)
		fmt.Fprintf(&buf, "--%s--\r\n", frontiere)
	}
	return id, buf.Bytes(), nil
}

func adresse(s string) string {
	if i := strings.Index(s, "<"); i >= 0 {
		if j := strings.Index(s[i:], ">"); j > 0 {
			return s[i+1 : i+j]
		}
	}
	return strings.TrimSpace(s)
}

func contient(l []string, x string) bool {
	for _, v := range l {
		if v == x {
			return true
		}
	}
	return false
}
