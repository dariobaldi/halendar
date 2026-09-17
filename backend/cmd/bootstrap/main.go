// Command bootstrap creates the first user for a fresh instance, with full
// admin rights.
//
// POST /v1/users (the normal registration endpoint) requires an existing
// admin to call it, so on a brand new database there is no way to create any
// user at all without going around the API. Run this once after migrations:
//
//	go run ./cmd/bootstrap -email=you@example.com -username=you
//
// It prompts for anything not passed as a flag, including the password
// (read from stdin, not echoed).
package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/validator"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/term"
)

// Mirrors DevLevel in cmd/api/routes.go -- kept as a separate literal here
// since cmd/bootstrap intentionally doesn't depend on package main of cmd/api.
const devAccessLevel = 100

func main() {
	name := flag.String("name", "", "Full name (prompted if omitted)")
	email := flag.String("email", "", "Email address (prompted if omitted)")
	username := flag.String("username", "", "Username (prompted if omitted)")
	password := flag.String("password", "", "Password (prompted, hidden, if omitted)")
	flag.Parse()

	// cmd/api/env.go loads .env the same way, relative to the backend/
	// directory -- this command is meant to be run with `go run
	// ./cmd/bootstrap` from backend/, same as `go run ./cmd/api`.
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file (run this from the backend/ directory)")
	}

	dsn := os.Getenv("DB_DSN_DEV")
	if os.Getenv("ENV") == "production" {
		dsn = os.Getenv("DB_DSN")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("could not connect to the database (is it running and migrated? see `make setup`): %s", err)
	}

	reader := bufio.NewReader(os.Stdin)

	if *name == "" {
		*name = prompt(reader, "Full name")
	}
	if *email == "" {
		*email = prompt(reader, "Email")
	}
	if *username == "" {
		*username = prompt(reader, "Username")
	}
	if *password == "" {
		*password = promptHidden("Password")
	}

	user := &data.User{
		ID:        uuid.New(),
		Name:      *name,
		Email:     *email,
		Username:  *username,
		Activated: true,
	}
	if err := user.Password.Set(*password); err != nil {
		log.Fatal(err)
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		for field, msg := range v.Errors {
			fmt.Fprintf(os.Stderr, "%s: %s\n", field, msg)
		}
		os.Exit(1)
	}

	models := data.NewModels(db)
	if err := models.Users.Insert(user); err != nil {
		log.Fatal(err)
	}

	// Users.Insert always creates the row at access_level 1 (a plain user,
	// see internal/data/users.go) -- this is the one place that grants admin
	// rights, since there's no API route that can.
	_, err = db.ExecContext(ctx, `UPDATE users SET access_level = $1 WHERE id = $2`, devAccessLevel, user.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created admin user %q (%s), access level %d. Log in from the app now.\n", *username, *email, devAccessLevel)
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Printf("%s: ", label)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(line)
}

func promptHidden(label string) string {
	fmt.Printf("%s: ", label)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(bytePassword))
}
