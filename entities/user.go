package entities

// User is the user sql server struct.
type User struct {
	ID         int
	UUID       string
	Fname      string
	Lname      string
	Uname      string
	Email      string
	Password   string
	Userrole   int8
	status     int8
	Facebookid int
	IfLoggedIn bool
}
