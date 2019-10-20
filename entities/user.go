package entities

type User struct {
	ID         int
	UUID       string
	Fname      string
	Lname      string
	Uname      string
	Email      string
	Password   string
	Role       int
	Facebookid int
	Userrole   int8
	IfLoggedIn bool
	Errors     map[string]string
}
