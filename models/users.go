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

// GetUsers gets a list of users from the database.
func (userConnection UserConnection) GetUsers(cp int, pp int) ([]entities.User, error) {
	const (
		execTvp = "spGetUsers @CurrentPage, @PerPage"
	)
	rows, err := userConnection.Db.Query(execTvp,
		sql.Named("CurrentPage", cp),
		sql.Named("PerPage", pp),
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	users := []entities.User{}
	for rows.Next() {
		var id int
		var uuid string
		var fname string
		var lname string
		var uname string
		var email string
		var password string
		var facebookid int
		var userrole int8
		var status int8
		var image string

		err := rows.Scan(&id, &uuid, &fname, &lname, &uname, &email, &password, &facebookid, &userrole, &status, &image)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		user := entities.User{
			ID:         id,
			UUID:       uuid,
			Fname:      fname,
			Lname:      lname,
			Uname:      uname,
			Email:      email,
			Password:   password,
			Facebookid: facebookid,
			Userrole:   userrole,
			Status:     status,
			Image:      image,
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserByID gets the row in Users table associated with the given UUID.
func (userConnection UserConnection) GetUserByID(id string) (entities.User, error) {
	const (
		execTvp = "spGetUserByID @ID"
	)
	result := userConnection.Db.QueryRow(execTvp,
		sql.Named("ID", id),
	)
	var nid int
	var uuid string
	var fname string
	var lname string
	var uname string
	var email string
	var password string
	var facebookid int
	var userrole int8
	var status int8
	var image string

	err := result.Scan(&nid, &uuid, &fname, &lname, &uname, &email, &password, &facebookid, &userrole, &status, &image)

	user := entities.User{
		ID:       nid,
		UUID:     uuid,
		Fname:    fname,
		Lname:    lname,
		Uname:    uname,
		Email:    email,
		Password: password,
		Userrole: userrole,
		Status:   status,
		Image:    image,
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
	var status int8
	var image string

	err := result.Scan(&ID, &uuid, &fname, &lname, &nuname, &email, &password, &facebookid, &userrole, &status, &image)

	user := entities.User{
		ID:         ID,
		UUID:       uuid,
		Fname:      fname,
		Lname:      lname,
		Uname:      nuname,
		Email:      email,
		Password:   password,
		Facebookid: facebookid,
		Userrole:   userrole,
		Status:     status,
		Image:      image,
	}
	if err != nil {
		return user, err
	}
	return user, err
}

// DeleteUser deletes a row in the user database.
func (userConnection UserConnection) DeleteUser(id string) bool {
	const (
		execTvp = "spDeleteUser @ID"
	)
	_, err := userConnection.Db.Exec(execTvp,
		sql.Named("ID", id),
	)
	if err != nil {
		log.Fatal(err)
	}

	return true
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
func (userConnection UserConnection) CreateUser(us entities.User) bool {
	password := []byte(us.Password)
	hPass := hashAndSalt(password)

	const (
		execTvp = "spCreateUser @UUID, @Image, @Fname, @Lname, @Uname, @Email, @Password, @Facebookid, @Userrole"
	)
	_, err := userConnection.Db.Exec(execTvp,
		sql.Named("UUID", us.UUID),
		sql.Named("Image", us.Image),
		sql.Named("Fname", us.Fname),
		sql.Named("Lname", us.Lname),
		sql.Named("Uname", us.Uname),
		sql.Named("Email", us.Email),
		sql.Named("Password", hPass),
		sql.Named("Facebookid", us.Facebookid),
		sql.Named("Userrole", us.Userrole),
	)
	if err != nil {
		return false
	}

	return true
}

// UpdateUser updates an existing user without image.
func (userConnection UserConnection) UpdateUser(us entities.User) bool {
	const (
		execTvp = "spUpdateUser @ID, @Fname, @Lname, @Uname, @Email"
	)
	_, err := userConnection.Db.Exec(execTvp,
		sql.Named("ID", us.ID),
		sql.Named("Fname", us.Fname),
		sql.Named("Lname", us.Lname),
		sql.Named("Uname", us.Uname),
		sql.Named("Email", us.Email),
	)
	if err != nil {
		return false
	}

	return true
}

// UpdateUserImage updates an existing user with image.
func (userConnection UserConnection) UpdateUserImage(us entities.User) bool {
	const (
		execTvp = "spUpdateUserImage @ID, @Image, @Fname, @Lname, @Uname, @Email"
	)
	_, err := userConnection.Db.Exec(execTvp,
		sql.Named("ID", us.ID),
		sql.Named("Image", us.Image),
		sql.Named("Fname", us.Fname),
		sql.Named("Lname", us.Lname),
		sql.Named("Uname", us.Uname),
		sql.Named("Email", us.Email),
	)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

// UpdatePassword updates an existing users password.
func (userConnection UserConnection) UpdatePassword(password string, ID string) bool {
	passw := []byte(password)
	hPass := hashAndSalt(passw)

	const (
		execTvp = "spUpdatePassword @Password, @ID"
	)
	_, err := userConnection.Db.Exec(execTvp,
		sql.Named("Password", hPass),
		sql.Named("ID", ID),
	)
	if err != nil {
		return false
	}

	return true
}

// UpdateStatus updates an existing users status.
func (userConnection UserConnection) UpdateStatus(status string, ID string) bool {
	const (
		execTvp = "spUpdateUserStatus @Status, @ID"
	)
	if status == "1" {
		_, err := userConnection.Db.Exec(execTvp,
			sql.Named("Status", 1),
			sql.Named("ID", ID),
		)
		if err != nil {
			return false
		}
	} else {
		_, err := userConnection.Db.Exec(execTvp,
			sql.Named("Status", 0),
			sql.Named("ID", ID),
		)
		if err != nil {
			return false
		}
	}

	return true
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}
