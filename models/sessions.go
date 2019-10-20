package models

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Trigve-Hagen/rlayouts/entities"
)

// UserSession references the database connection.
type UserSession struct {
	Db *sql.DB
}

// GetSessionByUUID gets the row in UserSessions associated with the given UUID.
func (userSession UserSession) GetSessionByUUID(uuid string) (entities.Session, error) {
	const (
		execTvp = "spGetUserSessionByUUID @UUID"
	)
	result := userSession.Db.QueryRow(execTvp,
		sql.Named("UUID", uuid),
	)
	var ID int
	var nuuid string
	var useruuid string
	var createdat string

	err := result.Scan(&ID, &nuuid, &useruuid, &createdat)

	sess := entities.Session{
		UUID:     nuuid,
		UserUUID: useruuid,
		DateTime: createdat,
	}
	if err != nil {
		fmt.Println(err)
		return sess, err
	}
	return sess, err
}

// SaveSearchStrings updates the session and saves any searchs to help tailor the users content when they return.
func (userSession UserSession) SaveSearchStrings(us entities.Session) bool {
	fmt.Println(us)
	return true
}

// CreateSession creates arow in the database per user session.
func (userSession UserSession) CreateSession(us entities.Session) (entities.Session, error) {
	const (
		execTvp = "spCreateUserSession @UUID, @UserUUID, @Created_at"
	)
	_, err := userSession.Db.Exec(execTvp,
		sql.Named("UUID", us.UUID),
		sql.Named("UserUUID", us.UserUUID),
		sql.Named("Created_at", us.DateTime),
	)
	if err != nil {
		log.Fatal(err)
	}

	return us, nil
}

// CleanSessions checks the datetime and erases all session rows n days before.
func (userSession UserSession) CleanSessions(us entities.User) bool {
	fmt.Println(us)
	return true
}
