package entities

// Session is a good place to store user search information for improving sales or customizing content.
type Session struct {
	UUID      string
	UserUUID  string
	CreatedAt string
	SearchSt  string
}
