package models

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Trigve-Hagen/rlayouts/entities"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserConnection references the database connection.
type UserConnection struct {
	Db *sql.DB
}

// GetUserByID gets the row in Users table associated with the given UUID.
func (userConnection UserConnection) GetUserByID(uuid string) (entities.User, error) {
	const (
		execTvp = "spGetUserByUUID @UUID"
	)
	result := userConnection.Db.QueryRow(execTvp,
		sql.Named("UUID", uuid),
	)
	var ID int
	var nuuid string
	var fname string
	var lname string
	var nuname string
	var email string
	var password string
	var facebookid int
	var userrole int8

	err := result.Scan(&ID, &nuuid, &fname, &lname, &nuname, &email, &password, &facebookid, &userrole)

	user := entities.User{
		ID:         ID,
		UUID:       uuid,
		Fname:      fname,
		Lname:      lname,
		Uname:      nuname,
		Email:      email,
		Password:   password,
		Userrole:   userrole,
		IfLoggedIn: true,
	}
	if err != nil {
		fmt.Println(err)
		return user, err
	}
	return user, err
}

// CheckLoginForm takes the username from login form and gets the row in the database.
func (userConnection UserConnection) CheckLoginForm(uname string) (entities.User, error) {
	const (
		execTvp = "spCheckLogin @Uname"
	)
	result := userConnection.Db.QueryRow(execTvp,
		sql.Named("Uname", uname),
	)
	var ID int
	var uuid string
	var fname string
	var lname string
	var nuname string
	var email string
	var password string
	var facebookid int
	var userrole int8

	err := result.Scan(&ID, &uuid, &fname, &lname, &nuname, &email, &password, &facebookid, &userrole)

	user := entities.User{
		ID:         ID,
		UUID:       uuid,
		Fname:      fname,
		Lname:      lname,
		Uname:      nuname,
		Email:      email,
		Password:   password,
		Userrole:   userrole,
		IfLoggedIn: true,
	}
	if err != nil {
		fmt.Println(err)
		return user, err
	}
	return user, err
}

// CheckEmailForgotPassword checks for an email then sends an email to the submitted email to change password.
func (userConnection UserConnection) CheckEmailForgotPassword(email string) bool {
	fmt.Println(email)
	return true
}

// CheckForUnique could check for uniqueness but the check is done in the database so im going to see what is returned.
func (userConnection UserConnection) CheckForUnique(rowName string, rowValue string) bool {
	fmt.Println(rowValue)
	return true
}

// CreateAdminUserIfNotExists is called in main.go init and creates an admin user if one does not exist.
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

// CreateUser is call in Register and creates a new user.
func (userConnection UserConnection) CreateUser(us entities.User) (entities.User, error) {
	password := []byte(us.Password)
	hPass := hashAndSalt(password)

	const (
		execTvp = "spCreateAdminIfNotExists @UUID, @Fname, @Lname, @Uname, @Email, @Password, @Facebookid, @Userrole"
	)
	_, err := userConnection.Db.Exec(execTvp,
		sql.Named("UUID", us.UUID),
		sql.Named("Fname", us.Fname),
		sql.Named("Lname", us.Lname),
		sql.Named("Uname", us.Uname),
		sql.Named("Email", us.Email),
		sql.Named("Password", hPass),
		sql.Named("Facebookid", us.Facebookid),
		sql.Named("Userrole", us.Userrole),
	)
	if err != nil {
		log.Fatal(err)
	}

	return us, nil
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}
