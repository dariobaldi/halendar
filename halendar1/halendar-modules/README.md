# Halendar — modules mail et agenda

Regroupe ce qu'on avait testé séparément (lecture IMAP, envoi SMTP, agenda CalDAV)
en **deux modules Go réutilisables**, qui marchent de la même façon :

```go
envfile.Charger(".env")

boite := mail.Nouveau(mail.ConfigDepuisEnv())       // ou mail.Config{...}
ag    := agenda.Nouveau(agenda.ConfigDepuisEnv())   // ou agenda.Config{...}

boite.Tester(ctx)
ag.Tester(ctx)
```

## Démarrer

```bash
cp .env.example .env    # identifiants du compte mail et de l'agenda
go mod tidy
go run . sante
```

## Module `mail`

| Fonction | Rôle |
|---|---|
| `Derniers(ctx, n)` | n derniers mails (plus récent d'abord) |
| `Nouveaux(ctx, dernierUID)` | mails arrivés depuis le dernier appel (renvoie le nouvel UID à garder) |
| `Lire(ctx, uid)` | un mail : expéditeur, destinataires, texte, HTML, pièces jointes, lu/non lu |
| `Rechercher(ctx, Recherche{...})` | non lus, expéditeur, sujet, contenu, dates |
| `MarquerLu(ctx, lu, uids...)` | lu / non lu |
| `Deplacer(ctx, dossier, uids...)` | ranger un mail |
| `Dossiers(ctx)` · `Compter(ctx)` | dossiers du compte · nombre de mails |
| `Envoyer(ctx, Envoi{...})` | à / cc / cci, texte et HTML optionnel |
| `Repondre(ctx, message, texte)` | réponse dans le même fil (Re:, In-Reply-To, References) |
| `ReponseA(message, texte)` | prépare une réponse sans l'envoyer |
| `DeposerBrouillon(ctx, Envoi{...})` | dépose dans les Brouillons au lieu d'envoyer |

## Module `agenda`

| Fonction | Rôle |
|---|---|
| `Agendas(ctx)` | noms des agendas (les « Rappels » sont ignorés) |
| `Evenements(ctx, debut, fin)` | tous les RDV de la période, récurrences dépliées |
| `Occupe(ctx, debut, fin)` | le créneau est-il pris ? et par quoi |
| `Ajouter(ctx, Evenement{...})` | crée, ou met à jour si l'UID existe (statut confirme / provisoire / annule) |
| `Supprimer(ctx, uid, agenda)` | supprime un RDV |

`agenda.Evenement` se lit directement depuis du JSON :

```json
{ "titre": "Pitch", "debut": "2026-09-18T09:00", "duree_minutes": 60, "statut": "provisoire" }
```

(`"debut": "2026-09-19"` = journée entière ; `"fin"` à la place de `"duree_minutes"` ; `"fuseau"` optionnel.)

## Démo en ligne de commande

```bash
go run . sante
go run . mails 5
go run . nonlus
go run . lire 42
go run . envoyer exemples/envoi.json
go run . repondre 42 "Jeudi 14h me convient."
go run . brouillon exemples/envoi.json
go run . agenda 7
go run . ajouter exemples/evenements.json
go run . supprimer evt-1234abcd
```

## Tests

```bash
go test ./...
```

Les tests utilisent de faux serveurs IMAP, SMTP et CalDAV (`testutil/`) : aucun compte nécessaire.
