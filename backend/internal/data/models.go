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
	Devices          DeviceModel
	Tokens           TokenModel
	Users            UserModel
	EmailAccounts    EmailAccountModel
	EmailMessages    EmailMessageModel
	EmailEvents      EmailEventModel
	CalendarAccounts CalendarAccountModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Devices:          DeviceModel{DB: db},
		Tokens:           TokenModel{DB: db},
		Users:            UserModel{DB: db},
		EmailAccounts:    EmailAccountModel{DB: db},
		EmailMessages:    EmailMessageModel{DB: db},
		EmailEvents:      EmailEventModel{DB: db},
		CalendarAccounts: CalendarAccountModel{DB: db},
	}
}
