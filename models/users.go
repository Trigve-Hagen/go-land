package models

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Trigve-Hagen/rlayouts/entities"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserConnection struct {
	Db *sql.DB
}

func (userConnection UserConnection) GetUserByID(uuid string) ([]entities.User, error) {
	rows, err := userConnection.Db.Query("CALL spGetUserByUUID @UUID = ?", uuid)
	if err != nil {
		return nil, err
	}
	users := []entities.User{}
	for rows.Next() {
		var uuid string
		var fname string
		var lname string
		var uname string
		var email string
		var password string
		err := rows.Scan(&uuid, &fname, &lname, &uname, &email, &password)
		if err != nil {
			return nil, err
		}
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
	return users, nil
}

func (userConnection UserConnection) CheckLoginForm(uname string, password string) bool {
	fmt.Println(uname)
	return true
}

func (userConnection UserConnection) CheckForUnique(rowName string, rowValue string) bool {
	fmt.Println(rowValue)
	return true
}

func (userConnection UserConnection) CreateAdminUserIfNotExists() (entities.User, error) {
	password := []byte("password")
	hPass := hashAndSalt(password)
	uuidreg, err := uuid.NewUUID()
	us := entities.User{
		UUID:       uuidreg.String(),
		Fname:      "Trigve",
		Lname:      "Hagen",
		Uname:      "trigve",
		Email:      "trigve.hagen@gmail.com",
		Password:   hPass,
		Facebookid: 0,
		Userrole:   1,
	}
	if err != nil {
		return us, err
	}
	const (
		execTvp = "spCreateAdminIfNotExists @UUID, @Fname, @Lname, @Uname, @Email, @Password, @Facebookid, @Userrole"
	)
	_, err = userConnection.Db.Exec(execTvp,
		sql.Named("UUID", us.UUID),
		sql.Named("Fname", us.Fname),
		sql.Named("Lname", us.Lname),
		sql.Named("Uname", us.Uname),
		sql.Named("Email", us.Email),
		sql.Named("Password", us.Password),
		sql.Named("Facebookid", us.Facebookid),
		sql.Named("Userrole", us.Userrole),
	)
	if err != nil {
		log.Fatal(err)
	}

	return us, nil
}

func (userConnection UserConnection) CreateUser(us entities.User) int {
	fmt.Println(us)
	return 7
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}
