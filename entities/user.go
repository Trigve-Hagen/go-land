package entities

type User struct {
	ID         int
	UUID       string
	Fname      string
	Lname      string
	Uname      string
	Email      string
	Password   string
	Userrole   int8
	Facebookid int
	IfLoggedIn bool
	Errors     map[string]string
}
