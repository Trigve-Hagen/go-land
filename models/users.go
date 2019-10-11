package users

import (
	"database/sql"

	"github.com/Trigve-Hagen/rlayouts/entities"
)

type UserConnection struct {
	Db *sql.DB
}

func (userConnection UserConnection) GetUsers() ([]entities.User, error) {
	rows, err := userConnection.Db.Query("SELECT * FROM Users")
	if err != nil {
		return nil, err
	} else {
		users := []entities.User{}
		for rows.Next() {
			var id int64
			var name string
			var email string
			var age int
			var gender string
			err2 := rows.Scan(&id, &name, &email, &age, &gender)
			if err2 != nil {
				return nil, err2
			} else {
				user := entities.User{
					ID:     id,
					Name:   name,
					Email:  email,
					Age:    age,
					Gender: gender,
				}
				users = append(users, user)
			}
		}
		return users, nil
	}
}
