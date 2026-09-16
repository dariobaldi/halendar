// Package envfile charge un fichier .env (CLE=valeur) dans les variables d'environnement.
package envfile

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Charger lit le fichier. Les variables déjà définies (terminal, Docker) gardent
// la priorité. Un fichier absent n'est pas une erreur.
func Charger(chemin string) error {
	f, err := os.Open(chemin)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		ligne := strings.TrimSpace(sc.Text())
		if ligne == "" || strings.HasPrefix(ligne, "#") {
			continue
		}
		ligne = strings.TrimPrefix(ligne, "export ")
		cle, val, ok := strings.Cut(ligne, "=")
		if !ok {
			return fmt.Errorf("%s ligne %d : '=' manquant", chemin, n)
		}
		cle, val = strings.TrimSpace(cle), strings.TrimSpace(val)
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
			val = val[1 : len(val)-1]
		} else if i := strings.Index(val, " #"); i >= 0 {
			val = strings.TrimSpace(val[:i])
		}
		if _, existe := os.LookupEnv(cle); !existe {
			os.Setenv(cle, val)
		}
	}
	return sc.Err()
}

// Texte renvoie la variable ou la valeur par défaut.
func Texte(cle, defaut string) string {
	if v := strings.TrimSpace(os.Getenv(cle)); v != "" {
		return v
	}
	return defaut
}

// Entier renvoie la variable convertie en entier, ou la valeur par défaut.
func Entier(cle string, defaut int) int {
	if v, err := strconv.Atoi(Texte(cle, "")); err == nil {
		return v
	}
	return defaut
}

// Booleen accepte true/false, 1/0, oui/non.
func Booleen(cle string, defaut bool) bool {
	switch strings.ToLower(Texte(cle, "")) {
	case "true", "1", "oui", "yes":
		return true
	case "false", "0", "non", "no":
		return false
	}
	return defaut
}

// Manquantes liste les variables vides parmi celles demandées.
func Manquantes(cles ...string) []string {
	var m []string
	for _, k := range cles {
		if Texte(k, "") == "" {
			m = append(m, k)
		}
	}
	return m
}
