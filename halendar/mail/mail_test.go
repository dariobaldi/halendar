package mail_test

import (
	"context"
	"strings"
	"testing"

	"halendar/mail"
	"halendar/testutil"
)

func boite(t *testing.T) (*mail.Boite, *testutil.FauxIMAP, *testutil.FauxSMTP) {
	im := testutil.NouveauFauxIMAP(t)
	sm := testutil.NouveauFauxSMTP(t)
	b := mail.Nouveau(mail.Config{
		IMAPHost: im.Addr, IMAPSansTLS: true,
		SMTPHost: sm.Host, SMTPPort: sm.Port,
		User: im.User, Pass: im.Pass,
	})
	return b, im, sm
}

func TestLireRechercherMarquerDeplacer(t *testing.T) {
	ctx := context.Background()
	b, im, _ := boite(t)
	if err := b.Tester(ctx); err != nil {
		t.Fatal(err)
	}

	// Nouveaux : premier appel = point de départ, rien n'est renvoyé
	im.Deposer(t, "<ancien@x>", "vieux@exemple.fr", "Ancien", "historique")
	msgs, dernier, err := b.Nouveaux(ctx, 0)
	if err != nil || len(msgs) != 0 || dernier != 1 {
		t.Fatalf("point de départ : %d msgs, dernier=%d, err=%v", len(msgs), dernier, err)
	}
	im.Deposer(t, "<rdv@x>", "Claire Martin <claire@exemple.fr>", "Point projet", "Dispo jeudi 14h ?")
	im.Deposer(t, "<news@x>", "news@exemple.fr", "Newsletter", "Promo")
	msgs, dernier, err = b.Nouveaux(ctx, dernier)
	if err != nil || len(msgs) != 2 || dernier != 3 {
		t.Fatalf("nouveaux : %d msgs, dernier=%d, err=%v", len(msgs), dernier, err)
	}
	if msgs, _, _ := b.Nouveaux(ctx, dernier); len(msgs) != 0 {
		t.Fatalf("aucun nouveau attendu, obtenu %d", len(msgs))
	}

	// Derniers : du plus récent au plus ancien, avec Message-ID entre crochets
	derniers, err := b.Derniers(ctx, 2)
	if err != nil || len(derniers) != 2 || derniers[0].Sujet != "Newsletter" || derniers[1].ID != "<rdv@x>" {
		t.Fatalf("derniers : %+v %v", derniers, err)
	}
	claire := derniers[1]
	if claire.De != "claire@exemple.fr" || claire.DeNom != "Claire Martin" || claire.Lu {
		t.Fatalf("message mal lu : %+v", claire)
	}

	// Rechercher + MarquerLu
	res, err := b.Rechercher(ctx, mail.Recherche{Sujet: "projet"})
	if err != nil || len(res) != 1 || res[0].UID != claire.UID {
		t.Fatalf("recherche par sujet : %+v %v", res, err)
	}
	if err := b.MarquerLu(ctx, true, claire.UID); err != nil {
		t.Fatal(err)
	}
	nonLus, _ := b.Rechercher(ctx, mail.Recherche{NonLus: true})
	if len(nonLus) != 2 {
		t.Fatalf("attendu 2 non lus, obtenu %d", len(nonLus))
	}
	if m, _ := b.Lire(ctx, claire.UID); !m.Lu {
		t.Fatal("le message devrait être lu")
	}

	// Dossiers + Deplacer
	dossiers, _ := b.Dossiers(ctx)
	if !strings.Contains(strings.Join(dossiers, ","), "Archives") {
		t.Fatalf("dossiers : %v", dossiers)
	}
	if err := b.Deplacer(ctx, "Archives", claire.UID); err != nil {
		t.Fatal(err)
	}
	if im.Compter("INBOX") != 2 || im.Compter("Archives") != 1 {
		t.Fatalf("déplacement : INBOX=%d Archives=%d", im.Compter("INBOX"), im.Compter("Archives"))
	}
}

func TestEnvoyerRepondreBrouillon(t *testing.T) {
	ctx := context.Background()
	b, im, sm := boite(t)
	im.Deposer(t, "<rdv@x>", "Claire <claire@exemple.fr>", "Point projet", "Dispo jeudi ?")
	msgs, _ := b.Derniers(ctx, 1)

	// Répondre : destinataire, Re:, fil de discussion
	if _, err := b.Repondre(ctx, msgs[0], "Jeudi 14h me convient."); err != nil {
		t.Fatal(err)
	}
	// Envoyer avec copie et HTML
	if _, err := b.Envoyer(ctx, mail.Envoi{A: []string{"a@x.fr"}, Cc: []string{"b@x.fr"}, Sujet: "Été ✓", Texte: "Bonjour", HTML: "<p>Bonjour</p>"}); err != nil {
		t.Fatal(err)
	}
	recus := sm.Mails()
	if len(recus) != 2 {
		t.Fatalf("attendu 2 mails, obtenu %d", len(recus))
	}
	r := recus[0].Brut
	for _, attendu := range []string{"To: claire@exemple.fr", "Subject: Re: Point projet", "In-Reply-To: <rdv@x>", "References: <rdv@x>"} {
		if !strings.Contains(r, attendu) {
			t.Errorf("réponse : %q absent de\n%s", attendu, r)
		}
	}
	if len(recus[1].A) != 2 || !strings.Contains(recus[1].Brut, "multipart/alternative") || !strings.Contains(recus[1].Brut, "=?utf-8?q?") {
		t.Errorf("envoi HTML/copie incorrect : %+v", recus[1])
	}

	// Erreur claire sans destinataire
	if _, err := b.Envoyer(ctx, mail.Envoi{Texte: "x"}); err == nil {
		t.Error("erreur attendue sans destinataire")
	}

	// Brouillon : déposé dans le dossier \Drafts, rien n'est envoyé
	dossier, err := b.DeposerBrouillon(ctx, mail.ReponseA(msgs[0], "À relire"))
	if err != nil || dossier != "Drafts" || im.Compter("Drafts") != 1 || len(sm.Mails()) != 2 {
		t.Fatalf("brouillon : dossier=%q err=%v drafts=%d envoyés=%d", dossier, err, im.Compter("Drafts"), len(sm.Mails()))
	}
}

func TestConfigDeduitSMTP(t *testing.T) {
	c := mail.Nouveau(mail.Config{IMAPHost: "imap.gmail.com:993", User: "moi@gmail.com", Pass: "x"}).Config()
	if c.SMTPHost != "smtp.gmail.com" || c.SMTPPort != 587 || c.From != "moi@gmail.com" || c.IMAPSansTLS {
		t.Fatalf("config : %+v", c)
	}
}
