package models

import (
	"database/sql"
	"log"

	"github.com/Trigve-Hagen/rlayouts/entities"
)

// PostConnection references the database connection.
type PostConnection struct {
	Db *sql.DB
}

// GetPostByID gets the row in Posts associated with the given ID.
func (postConnection PostConnection) GetPostByID(id string) (entities.Post, error) {
	const (
		execTvp = "spGetPostByID @ID"
	)
	result := postConnection.Db.QueryRow(execTvp,
		sql.Named("ID", id),
	)
	var postid int
	var useruuid string
	var image string
	var title string
	var body string
	var status int
	var createdat string

	err := result.Scan(&postid, &useruuid, &image, &title, &body, &status, &createdat)

	post := entities.Post{
		ID:        postid,
		UserUUID:  useruuid,
		Image:     image,
		Title:     title,
		Body:      body,
		Status:    status,
		CreatedAt: createdat,
	}
	if err != nil {
		return post, err
	}
	return post, err
}

// CreatePost creates a row in the post database.
func (postConnection PostConnection) CreatePost(pt entities.Post) bool {
	const (
		execTvp = "spCreatePost @UserUUID, @Image, @Title, @Body"
	)
	_, err := postConnection.Db.Exec(execTvp,
		sql.Named("UserUUID", pt.UserUUID),
		sql.Named("Image", pt.Image),
		sql.Named("Title", pt.Title),
		sql.Named("Body", pt.Body),
	)
	if err != nil {
		log.Fatal(err)
	}

	return true
}

// UpdatePost updates a post in the database.
func (postConnection PostConnection) UpdatePost(pt entities.Post) bool {
	const (
		execTvp = "spUpdatePost @ID, @UserUUID, @Image, @Title, @Body"
	)
	_, err := postConnection.Db.Exec(execTvp,
		sql.Named("ID", pt.ID),
		sql.Named("UserUUID", pt.UserUUID),
		sql.Named("Image", pt.Image),
		sql.Named("Title", pt.Title),
		sql.Named("Body", pt.Body),
	)
	if err != nil {
		log.Fatal(err)
	}

	return true
}

// DeletePost deletes a row in the post database.
func (postConnection PostConnection) DeletePost(id string) bool {
	const (
		execTvp = "spDeletePost @ID"
	)
	_, err := postConnection.Db.Exec(execTvp,
		sql.Named("ID", id),
	)
	if err != nil {
		log.Fatal(err)
	}

	return true
}

// GetTotalPosts gets all posts for admin.
func (postConnection PostConnection) GetTotalPosts() (int, error) {
	const (
		execTvp = "spGetTotalPosts"
	)
	rows, err := postConnection.Db.Query(execTvp)
	if err != nil {
		return 0, err
	}

	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

// GetTotalPostsStatus gets all posts by status.
func (postConnection PostConnection) GetTotalPostsStatus(s int) (int, error) {
	const (
		execTvp = "spGetTotalPostsStatus @Status"
	)
	rows, err := postConnection.Db.Query(execTvp,
		sql.Named("Status", s),
	)
	if err != nil {
		return 0, err
	}

	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

// GetPosts gets all posts and returns them paginated latest to first created.
func (postConnection PostConnection) GetPosts(cp int, pp int) ([]entities.Post, error) {
	const (
		execTvp = "spGetPosts @CurrentPage, @PerPage"
	)
	rows, err := postConnection.Db.Query(execTvp,
		sql.Named("CurrentPage", cp),
		sql.Named("PerPage", pp),
	)
	if err != nil {
		return nil, err
	}

	posts := []entities.Post{}
	for rows.Next() {
		var id int
		var userid string
		var image string
		var title string
		var body string
		var status int
		var createdat string

		err := rows.Scan(&id, &userid, &image, &title, &body, &status, &createdat)
		if err != nil {
			return nil, err
		}
		post := entities.Post{
			ID:        id,
			UserUUID:  userid,
			Image:     image,
			Title:     title,
			Body:      body,
			Status:    status,
			CreatedAt: createdat,
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// GetPostsStatus gets all posts by status and returns them paginated latest to first created.
func (postConnection PostConnection) GetPostsStatus(cp int, pp int, s int) ([]entities.Post, error) {
	const (
		execTvp = "spGetPostsStatus @CurrentPage, @PerPage, @Status"
	)
	rows, err := postConnection.Db.Query(execTvp,
		sql.Named("CurrentPage", cp),
		sql.Named("PerPage", pp),
		sql.Named("Status", s),
	)

	if err != nil {
		return nil, err
	}
	posts := []entities.Post{}
	for rows.Next() {
		var id int
		var userid string
		var image string
		var title string
		var body string
		var createdat string

		err := rows.Scan(&id, &userid, &image, &title, &body, &createdat)
		if err != nil {
			return nil, err
		}
		post := entities.Post{
			ID:        id,
			UserUUID:  userid,
			Image:     image,
			Title:     title,
			Body:      body,
			CreatedAt: createdat,
		}
		posts = append(posts, post)
	}
	return posts, nil
}
