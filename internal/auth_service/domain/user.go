package domain

import "time"

type User struct {
	ID            string    `db:"id"`             // unique identifier
	Username      string    `db:"username"`       // mapped to 'name' in gRPC
	Email         string    `db:"email"`          // email address
	PasswordHash  string    `db:"password_hash"`  // hashed password
	Role          string    `db:"role"`           // user / admin / banned
	EmailVerified bool      `db:"email_verified"` // email confirmation status
	IsBanned      bool      `db:"is_banned"`      // user blocked by admin
	CreatedAt     time.Time `db:"created_at"`     // creation timestamp
	UpdatedAt     time.Time `db:"updated_at"`     // creation timestamp
	UploadBlocked bool      `db:"upload_blocked"`
	AvatarURL     string    `db:"avatar_url"`
}
