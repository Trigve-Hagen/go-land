package models

import (
	"database/sql"
)

// NewsletterConnection references the database connection.
type NewsletterConnection struct {
	Db *sql.DB
}

// CreateNewsletter creates a new newsletter entry.
func (newsletterConnection NewsletterConnection) CreateNewsletter(email string) bool {
	const (
		execTvp = "spCreateNewsletter @Email"
	)
	_, err := newsletterConnection.Db.Exec(execTvp,
		sql.Named("Email", email),
	)
	if err != nil {
		return false
	}

	return true
}
