package models

import (
	"database/sql"
	"fmt"

	"github.com/Trigve-Hagen/rlayouts/entities"
)

type UserConnection struct {
	Db *sql.DB
}

func (userConnection UserConnection) GetUserByID(uuid string) ([]entities.User, error) {
	rows, err := userConnection.Db.Query("CALL spGetUserByUUID @UUID = ?", uuid)
	if err != nil {
		return nil, err
	} else {
		users := []entities.User{}
		for rows.Next() {
			var uuid string
			var fname string
			var lname string
			var uname string
			var email string
			var password string
			err2 := rows.Scan(&uuid, &fname, &lname, &uname, &email, &password)
			if err2 != nil {
				return nil, err2
			} else {
				user := entities.User{
					UUID:     uuid,
					Fname:    fname,
					Lname:    lname,
					Uname:    uname,
					Email:    email,
					Password: password,
				}
				users = append(users, user)
			}
		}
		return users, nil
	}
}

func (userConnection UserConnection) CheckLoginForm(uname string, password string) bool {
	fmt.Println(uname)
	return true
}

func (userConnection UserConnection) CheckUsername(uname string) bool {
	fmt.Println(uname)
	return true
}

func (userConnection UserConnection) CreateUser(us entities.User) int {
	fmt.Println(us)
	return 7
}
