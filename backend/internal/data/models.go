package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrEditConflict    = errors.New("edit conflict")
	ErrDuplicateRecord = errors.New("duplicate record")
	ErrInvalidCustomer = errors.New("invalid customer")
	ErrInvalidAddress  = errors.New("invalid address")
	ErrInvalidShop     = errors.New("invalid shop")
)

type Models struct {
	Tokens TokenModel
	Users  UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Tokens: TokenModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
