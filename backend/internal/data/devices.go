package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"-"`
	PushToken  string    `json:"push_token"`
	Platform   string    `json:"platform"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type DeviceModel struct {
	DB *sql.DB
}

// Upsert registers a device's push token for a user, or refreshes it if that token
// is already registered — e.g. the same device re-registering, possibly under a
// different user after a re-login on a shared device.
func (m DeviceModel) Upsert(device *Device) error {
	query := `
		INSERT INTO devices (user_id, push_token, platform)
		VALUES ($1, $2, $3)
		ON CONFLICT (push_token) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			platform = EXCLUDED.platform,
			last_seen_at = NOW()
		RETURNING id, created_at, last_seen_at
	`
	args := []interface{}{device.UserID, device.PushToken, device.Platform}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&device.ID, &device.CreatedAt, &device.LastSeenAt)
}

// GetForUser returns every device registered for a user, most recently seen first.
func (m DeviceModel) GetForUser(userID uuid.UUID) ([]Device, error) {
	query := `
		SELECT id, user_id, push_token, platform, created_at, last_seen_at
		FROM devices
		WHERE user_id = $1
		ORDER BY last_seen_at DESC
	`
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := []Device{}
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.UserID, &d.PushToken, &d.Platform, &d.CreatedAt, &d.LastSeenAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// Delete removes a device by push token, scoped to the owning user so one user
// can't unregister another's device.
func (m DeviceModel) Delete(userID uuid.UUID, pushToken string) error {
	query := `DELETE FROM devices WHERE user_id = $1 AND push_token = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, userID, pushToken)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}
	return nil
}
