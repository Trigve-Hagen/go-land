package entities

// Post is the post sql server struct.
type Post struct {
	ID        int
	UserUUID  string
	Image     string
	Title     string
	Body      string
	Status    string
	CreatedAt string
}
