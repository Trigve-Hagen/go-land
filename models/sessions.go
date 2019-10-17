package models

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Trigve-Hagen/rlayouts/entities"
)

type UserSession struct {
	Db *sql.DB
}

func (userSession UserSession) GetSessionByID(uuid string) ([]entities.Session, error) {
	rows, err := userSession.Db.Query("CALL spGetUserSessionByUUID @UUID = ?", uuid)
	if err != nil {
		return nil, err
	} else {
		sessions := []entities.Session{}
		for rows.Next() {
			var id int64
			var date string
			var uuid string
			err := rows.Scan(&id, &date, &uuid)
			if err != nil {
				return nil, err
			} else {
				session := entities.Session{
					ID:   id,
					UUID: uuid,
					DateTime: date,
				}
				sessions = append(sessions, session)
			}
		}
		return sessions, nil
	}
}

func (userSession UserSession) alreadyLoggedIn(req *http.Request, use UserConnection) bool {
	c, err := req.Cookie("__ibes_")
	if err != nil {
		return false
	}
	session, err := userSession.GetSessionByID(c.Value)
	user, err := use.GetUserByID(session[0].UserUUID)
	_ = user
	if err != nil {
		return false
	}
	return true
}

func (userSession UserSession) CreateSession(us entities.User) int {
	fmt.Println(us)
	return 7
}

func (userSession UserSession) CleanSessions(us entities.User) bool {
	fmt.Println(us)
	return true
}
