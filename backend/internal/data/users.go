package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/dariobaldi/halendar_back/internal/validator"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateUsername = errors.New("duplicate username")
)

var AnonymousUser = &User{}

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	Password    password  `json:"-"`
	Activated   bool      `json:"activated"`
	AccessLevel int       `json:"access_level"`
	Version     int       `json:"version"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// User model

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (id, name, email, username, password_hash, activated, access_level)
		VALUES ($1, $2, $3, $4, $5, $6, 1)
		RETURNING created_at, version, access_level`

	args := []interface{}{user.ID, user.Name, user.Email, user.Username, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.CreatedAt, &user.Version, &user.AccessLevel)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetList(userID uuid.UUID, filter string) ([]*User, error) {
	query := `
		SELECT
			id,
			CASE
				WHEN id = $1 THEN 'moi-même'
			ELSE
				name
			END AS name
		FROM users
		WHERE
			($2 = 'ALL' OR activated = true)
		`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	args := []any{userID, filter}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.Name,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) GetListAdmin() ([]*User, error) {
	query := `
		SELECT
			id,
			created_at,
			name,
			email,
			username,
			activated,
			access_level,
			version
		FROM users
		ORDER BY activated DESC, name
		`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.Name,
			&user.Email,
			&user.Username,
			&user.Activated,
			&user.AccessLevel,
			&user.Version,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) Get(id uuid.UUID) (*User, error) {
	query := `
		SELECT 
			id,
			created_at,
			name,
			email,
			username,
			password_hash,
			activated,
			access_level,
			version
		FROM users
		WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Username,
		&user.Password.hash,
		&user.Activated,
		&user.AccessLevel,
		&user.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, username, password_hash, activated, access_level, version
		FROM users
		WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Username,
		&user.Password.hash,
		&user.Activated,
		&user.AccessLevel,
		&user.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetByUsername(username string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, username, password_hash, activated, access_level, version
		FROM users
		WHERE username = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Username,
		&user.Password.hash,
		&user.Activated,
		&user.AccessLevel,
		&user.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET
			name = $1,
			email = $2,
			username = $3,
			password_hash = $4,
			activated = $5,
			access_level = $6,
			version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Username,
		user.Password.hash,
		user.Activated,
		user.AccessLevel,
		user.ID,
		user.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetForToken(tokenPlaintext, scope string, delete bool) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	var query string
	if delete {
		query = `
		WITH deleted AS (
			DELETE FROM tokens t
			WHERE hash = $1 AND scope = $2 AND expiry > $3
			RETURNING *
		)
		SELECT u.id, u.created_at, u.name, u.email, u.username, u.password_hash, u.activated, u.access_level, u.version
		FROM users u
		INNER JOIN deleted t
		ON u.id = t.user_id`
	} else {
		query = `
		SELECT u.id, u.created_at, u.name, u.email, u.username, u.password_hash, u.activated, u.access_level, u.version
		FROM users u
		INNER JOIN tokens t
		ON u.id = t.user_id
		WHERE t.hash = $1 AND t.scope = $2 AND t.expiry > $3`
	}

	args := []interface{}{tokenHash[:], scope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Username,
		&user.Password.hash,
		&user.Activated,
		&user.AccessLevel,
		&user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Password methods

func (p *password) Set(plaintextPass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPass), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPass
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPass string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPass))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// Validation methods for Users

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.MatchesPattern(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidateUsername(v *validator.Validator, username string) {
	v.Check(len(username) >= 3, "username", "must be at least 3 bytes long")
	v.Check(len(username) <= 20, "username", "must not be more than 20 bytes long")
	v.Check(validator.MatchesPattern(username, validator.UsernameRx), "username", "must be a valid username, it can contain letters, numbers, and the symbols _ and -")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)
	ValidateUsername(v, user.Username)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}
