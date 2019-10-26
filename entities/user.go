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
	Facebookid int
	Userrole   int8
	Status     int8
}
