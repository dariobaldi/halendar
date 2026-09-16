package agenda_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"halendar/agenda"
	"halendar/testutil"
)

var paris, _ = time.LoadLocation("Europe/Paris")

func demain(h int) time.Time {
	d := time.Now().In(paris).AddDate(0, 0, 1)
	return time.Date(d.Year(), d.Month(), d.Day(), h, 0, 0, 0, paris)
}

func TestAgendaComplet(t *testing.T) {
	ctx := context.Background()
	f := testutil.NouveauFauxCalDAV(t)
	ag := agenda.Nouveau(agenda.Config{URL: f.URL, User: f.User, Pass: f.Pass})
	if err := ag.Tester(ctx); err != nil {
		t.Fatal(err)
	}
	if noms, _ := ag.Agendas(ctx); len(noms) != 1 || noms[0] != "Travail" {
		t.Fatalf("agendas : %v", noms)
	}

	// Ajouter depuis JSON (date locale + durée)
	var e agenda.Evenement
	raw := `{"titre":"Pitch hackathon","debut":"` + demain(9).Format("2006-01-02T15:04") + `","duree_minutes":30,"statut":"provisoire","lieu":"Salle 1","lien":"https://visio.exemple.fr/x"}`
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		t.Fatal(err)
	}
	cree, err := ag.Ajouter(ctx, e)
	if err != nil || cree.UID == "" || cree.Agenda != "Travail" || !cree.Fin.Equal(demain(9).Add(30*time.Minute)) {
		t.Fatalf("ajouter : %+v %v", cree, err)
	}

	// Lire : tous les champs reviennent
	evs, err := ag.Evenements(ctx, demain(0), demain(23))
	if err != nil || len(evs) != 1 {
		t.Fatalf("evenements : %+v %v", evs, err)
	}
	lu := evs[0]
	if lu.Titre != "Pitch hackathon" || lu.Statut != agenda.Provisoire || lu.Lieu != "Salle 1" || lu.Lien == "" || !lu.Debut.Equal(demain(9)) {
		t.Fatalf("relu : %+v", lu)
	}

	// Mettre à jour (même UID) : confirmé, pas de doublon
	lu.Statut = agenda.Confirme
	lu.Titre = "Pitch hackathon (confirmé)"
	if _, err := ag.Ajouter(ctx, lu); err != nil {
		t.Fatal(err)
	}
	evs, _ = ag.Evenements(ctx, demain(0), demain(23))
	if len(evs) != 1 || evs[0].Statut != agenda.Confirme {
		t.Fatalf("mise à jour : %+v", evs)
	}

	// Occupe
	if occupe, conflits, _ := ag.Occupe(ctx, demain(9).Add(15*time.Minute), demain(10)); !occupe || len(conflits) != 1 {
		t.Fatal("9h15-10h devrait être occupé")
	}
	if occupe, _, _ := ag.Occupe(ctx, demain(14), demain(15)); occupe {
		t.Fatal("14h-15h devrait être libre")
	}

	// Journée entière depuis JSON
	var j agenda.Evenement
	json.Unmarshal([]byte(`{"titre":"Hackathon","debut":"`+demain(0).Format("2006-01-02")+`"}`), &j)
	if r, err := ag.Ajouter(ctx, j); err != nil || !r.JourneeEntiere || r.Fin.Sub(r.Debut) != 24*time.Hour {
		t.Fatalf("journée entière : %+v %v", r, err)
	}

	// Supprimer
	if err := ag.Supprimer(ctx, lu.UID, ""); err != nil {
		t.Fatal(err)
	}
	evs, _ = ag.Evenements(ctx, demain(8), demain(12))
	for _, e := range evs {
		if e.UID == lu.UID {
			t.Fatalf("après suppression, encore présent : %+v", e)
		}
	}

	// Erreurs claires
	if _, err := ag.Ajouter(ctx, agenda.Evenement{Titre: "Sans date"}); err == nil {
		t.Error("erreur attendue sans date")
	}
	if _, err := ag.Ajouter(ctx, agenda.Evenement{Titre: "x", Debut: demain(9), Agenda: "Vacances"}); err == nil {
		t.Error("erreur attendue pour un agenda inconnu")
	}
}
