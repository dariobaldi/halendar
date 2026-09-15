// Démo en ligne de commande des modules mail et agenda.
// Identifiants dans le fichier .env (voir .env.example).
//
//	go run . sante                   teste la connexion mail et agenda
//	go run . mails [n]               n derniers mails
//	go run . nonlus                  mails non lus
//	go run . lire <uid>              un mail complet
//	go run . envoyer envoi.json      envoie un mail
//	go run . repondre <uid> "texte"  répond dans le même fil
//	go run . brouillon envoi.json    dépose le mail dans les Brouillons
//	go run . agenda [jours]          emploi du temps
//	go run . ajouter evenement.json  ajoute (ou met à jour) un ou plusieurs RDV
//	go run . supprimer <uid>         supprime un RDV
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"halendar/agenda"
	"halendar/envfile"
	"halendar/mail"
)

func main() {
	if err := envfile.Charger(".env"); err != nil {
		fmt.Fprintln(os.Stderr, "✗", err)
		os.Exit(1)
	}
	boite := mail.Nouveau(mail.ConfigDepuisEnv())
	ag := agenda.Nouveau(agenda.ConfigDepuisEnv())

	if len(os.Args) < 2 {
		fmt.Println(aide)
		return
	}
	if err := executer(context.Background(), os.Args[1], os.Args[2:], boite, ag); err != nil {
		fmt.Fprintln(os.Stderr, "✗", err)
		os.Exit(1)
	}
}

func executer(ctx context.Context, cmd string, args []string, boite *mail.Boite, ag *agenda.Agenda) error {
	switch cmd {
	case "sante":
		ok := true
		if err := boite.Tester(ctx); err != nil {
			fmt.Println("✗ mail  ", err)
			ok = false
		} else {
			n, _ := boite.Compter(ctx)
			fmt.Printf("✓ mail   %s (%d messages) · envoi via %s:%d\n", boite.Config().User, n, boite.Config().SMTPHost, boite.Config().SMTPPort)
		}
		if err := ag.Tester(ctx); err != nil {
			fmt.Println("✗ agenda", err)
			ok = false
		} else {
			noms, _ := ag.Agendas(ctx)
			fmt.Println("✓ agenda", strings.Join(noms, ", "))
		}
		if !ok {
			return errors.New("au moins un module ne répond pas")
		}
		return nil

	case "mails":
		msgs, err := boite.Derniers(ctx, entier(args, 5))
		if err != nil {
			return err
		}
		afficherMails(msgs)
		return nil

	case "nonlus":
		msgs, err := boite.Rechercher(ctx, mail.Recherche{NonLus: true, Max: 20})
		if err != nil {
			return err
		}
		afficherMails(msgs)
		return nil

	case "lire":
		uid, err := uidArg(args)
		if err != nil {
			return err
		}
		m, err := boite.Lire(ctx, uid)
		if err != nil {
			return err
		}
		return afficherJSON(m)

	case "envoyer", "brouillon":
		var e mail.Envoi
		if err := lireJSON(args, &e); err != nil {
			return err
		}
		if cmd == "brouillon" {
			dossier, err := boite.DeposerBrouillon(ctx, e)
			if err != nil {
				return err
			}
			fmt.Printf("✓ brouillon déposé dans %q\n", dossier)
			return nil
		}
		id, err := boite.Envoyer(ctx, e)
		if err != nil {
			return err
		}
		fmt.Printf("✓ mail envoyé à %s (%s)\n", strings.Join(e.A, ", "), id)
		return nil

	case "repondre":
		uid, err := uidArg(args)
		if err != nil || len(args) < 2 {
			return errors.New(`usage : go run . repondre <uid> "texte"`)
		}
		m, err := boite.Lire(ctx, uid)
		if err != nil {
			return err
		}
		if _, err := boite.Repondre(ctx, *m, args[1]); err != nil {
			return err
		}
		fmt.Printf("✓ réponse envoyée à %s (« Re: %s »)\n", m.De, m.Sujet)
		return nil

	case "agenda":
		debut := time.Now()
		fin := debut.AddDate(0, 0, entier(args, 7))
		evs, err := ag.Evenements(ctx, debut, fin)
		if err != nil {
			return err
		}
		fmt.Printf("%d événement(s) du %s au %s\n", len(evs), debut.Format("02/01"), fin.Format("02/01"))
		jour := ""
		for _, e := range evs {
			if j := e.Debut.Format("Mon 02/01"); j != jour {
				jour = j
				fmt.Println("\n" + jour)
			}
			horaire := e.Debut.Format("15:04") + "–" + e.Fin.Format("15:04")
			if e.JourneeEntiere {
				horaire = "journée    "
			}
			fmt.Printf("  %s  %s  [%s · %s]  uid=%s\n", horaire, e.Titre, e.Agenda, e.Statut, e.UID)
		}
		return nil

	case "ajouter":
		raw, err := lireBrut(args)
		if err != nil {
			return err
		}
		var evs []agenda.Evenement
		if t := bytes.TrimSpace(raw); len(t) > 0 && t[0] == '[' {
			err = json.Unmarshal(t, &evs)
		} else {
			var e agenda.Evenement
			err = json.Unmarshal(t, &e)
			evs = []agenda.Evenement{e}
		}
		if err != nil {
			return fmt.Errorf("JSON invalide : %w", err)
		}
		reussis := 0
		for i, e := range evs {
			r, err := ag.Ajouter(ctx, e)
			if err != nil {
				fmt.Printf("✗ événement %d (%q) : %v\n", i+1, e.Titre, err)
				continue
			}
			reussis++
			fmt.Printf("✓ « %s » → %s, le %s de %s à %s (%s)\n", r.Titre, r.Agenda,
				r.Debut.Format("02/01/2006"), r.Debut.Format("15:04"), r.Fin.Format("15:04"), r.Statut)
		}
		if reussis < len(evs) {
			return fmt.Errorf("%d/%d événement(s) ajouté(s)", reussis, len(evs))
		}
		return nil

	case "supprimer":
		if len(args) == 0 {
			return errors.New("usage : go run . supprimer <uid> [agenda]")
		}
		nom := ""
		if len(args) > 1 {
			nom = args[1]
		}
		if err := ag.Supprimer(ctx, args[0], nom); err != nil {
			return err
		}
		fmt.Println("✓ supprimé")
		return nil
	}
	return fmt.Errorf("commande inconnue %q\n\n%s", cmd, aide)
}

func afficherMails(msgs []mail.Message) {
	if len(msgs) == 0 {
		fmt.Println("aucun mail")
	}
	for _, m := range msgs {
		lu := "●"
		if m.Lu {
			lu = " "
		}
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("%s UID %d · %s\n  De    %s %s\n  Sujet %s\n\n%s\n", lu, m.UID, m.Date.Format("02/01 15:04"), m.DeNom, m.De, m.Sujet, tronquer(m.Texte, 400))
	}
}

func lireBrut(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("fichier JSON manquant (ou « - » pour l'entrée standard)")
	}
	if args[0] == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(args[0])
}

func lireJSON(args []string, v any) error {
	raw, err := lireBrut(args)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("JSON invalide : %w", err)
	}
	return nil
}

func afficherJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func uidArg(args []string) (uint32, error) {
	if len(args) == 0 {
		return 0, errors.New("UID manquant")
	}
	n, err := strconv.ParseUint(args[0], 10, 32)
	return uint32(n), err
}

func entier(args []string, def int) int {
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func tronquer(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

const aide = `Commandes :
  go run . sante                   teste la connexion mail et agenda
  go run . mails [n]               n derniers mails
  go run . nonlus                  mails non lus
  go run . lire <uid>              un mail complet (JSON)
  go run . envoyer envoi.json      envoie un mail
  go run . repondre <uid> "texte"  répond dans le même fil
  go run . brouillon envoi.json    dépose le mail dans les Brouillons
  go run . agenda [jours]          emploi du temps
  go run . ajouter evenement.json  ajoute (ou met à jour) un ou plusieurs RDV
  go run . supprimer <uid>         supprime un RDV`
