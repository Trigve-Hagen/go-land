package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "net/http/pprof"

	"github.com/Trigve-Hagen/rlayouts/config"
	"github.com/Trigve-Hagen/rlayouts/entities"
	newsletters "github.com/Trigve-Hagen/rlayouts/models"
	posts "github.com/Trigve-Hagen/rlayouts/models"
	users "github.com/Trigve-Hagen/rlayouts/models"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
)

type message struct {
	Name    string
	Email   string
	Subject string
	Message string
}

type userData struct {
	IfLoggedIn bool
	NEmail     string
	RePassword string
	Message    message
	User       entities.User
	Users      []entities.User
	Post       entities.Post
	Posts      []entities.Post
	Errors     map[string]string
}

type sessionData struct {
	UUID      string
	UserUUID  string
	CreatedAt string
}

var tpl *template.Template
var viewData = map[string]userData{}
var messageData = map[string]message{}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*/*/*.gohtml"))
	db, err := config.GetMSSQLDB()
	if err != nil {
		fmt.Println("Database: ", err)
	}
	userConnection := users.UserConnection{
		Db: db,
	}
	userConnection.CreateAdminUserIfNotExists()
	/*for _, t := range tpl.Templates() {
		fmt.Println(t.Name())
	}*/
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/about", about)
	http.HandleFunc("/contact", contact)
	http.HandleFunc("/admin/email", email)
	http.HandleFunc("/admin/go", goManager)
	http.HandleFunc("/admin/sql", sqlManager)
	http.HandleFunc("/admin/users", userManager)
	http.HandleFunc("/users/create", createUser)
	http.HandleFunc("/users/edit", editUser)
	http.HandleFunc("/users/handle", handleUser)
	http.HandleFunc("/admin/posts", postManager)
	http.HandleFunc("/posts/create", createPost)
	http.HandleFunc("/posts/edit", editPost)
	http.HandleFunc("/posts/handle", handlePost)
	http.HandleFunc("/admin/comments", commentManager)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/register", register)
	http.HandleFunc("/forgot/password", forgotPassword)
	http.HandleFunc("/auth/admin", admin)
	http.HandleFunc("/auth/profile", profile)
	http.HandleFunc("/auth/comments", comments)
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("public"))))
	http.ListenAndServe(":3000", nil)
}

func userManager(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if ud.IfLoggedIn == true {
		db, err := config.GetMSSQLDB()
		if err != nil {
			ud.Errors["Server"] = "Could not connect to database."
			render(res, "user-manager.gohtml", ud)
			return
		}
		userConnection := users.UserConnection{
			Db: db,
		}
		ausers, err := userConnection.GetUsers(1, 10)
		if err != nil {
			ud.Errors["Server"] = "Failed to retreive users."
			render(res, "user-manager.gohtml", ud)
			return
		}

		ud.Users = ausers

		render(res, "user-manager.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func createUser(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if ud.IfLoggedIn == true {
		user := entities.User{
			Fname:    "",
			Lname:    "",
			Uname:    "",
			Email:    "",
			Password: "",
			Userrole: 1,
		}
		ud.User = user
		if req.Method == http.MethodPost {
			vreg := &VCreateUser{
				Fname:      req.FormValue("fname"),
				Lname:      req.FormValue("lname"),
				Uname:      req.FormValue("uname"),
				Email:      req.FormValue("email"),
				Userrole:   req.FormValue("userrole"),
				Password:   req.FormValue("password"),
				RePassword: req.FormValue("rePassword"),
			}
			if vreg.Password == vreg.RePassword {
				if vreg.ValidateCreateUser() == false {
					ud.Errors = vreg.Errors
					render(res, "create-user.gohtml", vreg)
					return
				}

				uuidreg, err := uuid.NewUUID()
				if err != nil {
					ud.Errors["Server"] = "Failed to create registration UUID."
					render(res, "create-user.gohtml", ud)
					return
				}
				user := entities.User{
					UUID:     uuidreg.String(),
					Image:    "",
					Fname:    vreg.Fname,
					Lname:    vreg.Lname,
					Uname:    vreg.Uname,
					Email:    vreg.Email,
					Password: vreg.Password,
					Userrole: 2,
				}
				db, err := config.GetMSSQLDB()
				if err != nil {
					ud.Errors["Server"] = "Could not connect to database."
					render(res, "create-user.gohtml", ud)
					return
				}
				userConnection := users.UserConnection{
					Db: db,
				}
				if userConnection.CreateUser(user) == false {
					ud.Errors["Server"] = "Failed to create user."
					render(res, "create-user.gohtml", ud)
					return
				}
				ud.User = user

				render(res, "user-manager.gohtml", ud)
				return
			}
		}
		render(res, "create-user.gohtml", ud)
		return
	}

	render(res, "index.gohtml", ud)
}

func editUser(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		db, err := config.GetMSSQLDB()
		if err != nil {
			http.Redirect(res, req, "/admin/users", http.StatusServiceUnavailable)
			return
		}
		// create a function that gets the posts to pass to the return page
		// also delete the image from the users folder
		userConnection := users.UserConnection{
			Db: db,
		}
		user, err := userConnection.GetUserByID(req.FormValue("ID"))
		ud.User = user
		ud.User.Userrole = 1
		if err != nil {
			http.Redirect(res, req, "/admin/users", http.StatusServiceUnavailable)
			return
		}
		render(res, "edit-user.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func handleUser(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if ud.IfLoggedIn == true {
		if req.Method == http.MethodPost {
			db, err := config.GetMSSQLDB()
			if err != nil {
				http.Redirect(res, req, "/admin/users", http.StatusServiceUnavailable)
				return
			}
			// create a function that gets the posts to pass to the return page
			// also delete the image from the users folder
			userConnection := users.UserConnection{
				Db: db,
			}
			method := req.FormValue("method")
			switch method {
			case "DELETE":
				if req.FormValue("ID") == "1" {
					http.Redirect(res, req, "/admin/users", http.StatusServiceUnavailable)
					return
				}
				if userConnection.DeleteUser(req.FormValue("ID")) == false {
					http.Redirect(res, req, "/admin/users", http.StatusServiceUnavailable)
					return
				}
				ausers, err := userConnection.GetUsers(1, 10)
				if err != nil {
					http.Redirect(res, req, "/admin/users", http.StatusServiceUnavailable)
					return
				}

				ud.Users = ausers

				ud.Errors["Success"] = "User Deleted."
				render(res, "user-manager.gohtml", ud)
				return
			case "UPDATE-USER":
				vreg := &VUser{
					Fname: req.FormValue("fname"),
					Lname: req.FormValue("lname"),
					Uname: req.FormValue("uname"),
					Email: req.FormValue("email"),
				}
				if vreg.ValidateUser() == false {
					ud.Errors = vreg.Errors
					render(res, "edit-user.gohtml", ud)
					return
				}

				fname := ""
				mf, fh, err := req.FormFile("imgfile")
				if fh != nil {
					if err != nil {
						http.Redirect(res, req, "/users/edit", http.StatusServiceUnavailable)
						return
					}
					defer mf.Close()

					ext := strings.Split(fh.Filename, ".")[1]
					h := sha1.New()
					io.Copy(h, mf)
					fname = fmt.Sprintf("%x", h.Sum(nil)) + "." + ext

					wd, err := os.Getwd()
					if err != nil {
						http.Redirect(res, req, "/users/edit", http.StatusServiceUnavailable)
						return
					}

					newpath := filepath.Join(wd, "public", "images", "uploads", req.FormValue("ID"))
					if _, err := os.Stat(newpath); os.IsNotExist(err) {
						os.MkdirAll(newpath, os.ModePerm)
					}

					path := filepath.Join(wd, "public", "images", "uploads", req.FormValue("ID"), fname)
					nf, err := os.Create(path)
					if err != nil {
						http.Redirect(res, req, "/users/edit", http.StatusServiceUnavailable)
						return
					}
					defer nf.Close()

					mf.Seek(0, 0)
					io.Copy(nf, mf)
				}

				userid, err := strconv.Atoi(req.FormValue("ID"))
				if err != nil {
					http.Redirect(res, req, "/users/edit", http.StatusServiceUnavailable)
					return
				}

				if fname != "" {
					user := entities.User{
						ID:         userid,
						UUID:       ud.User.UUID,
						Fname:      vreg.Fname,
						Lname:      vreg.Lname,
						Uname:      vreg.Uname,
						Email:      vreg.Email,
						Password:   ud.User.Password,
						Userrole:   ud.User.Userrole,
						Facebookid: 0,
						Image:      fname,
					}
					ud.User = user
					if userConnection.UpdateUserImage(user) == false {
						ud.Errors["Server"] = "Failed to update user."
						switch req.FormValue("if_profile") {
						case "0":
							render(res, "edit-user.gohtml", ud)
						case "1":
							render(res, "profile.gohtml", ud)
						}
						return
					}

				} else {
					user := entities.User{
						ID:         userid,
						UUID:       ud.User.UUID,
						Fname:      vreg.Fname,
						Lname:      vreg.Lname,
						Uname:      vreg.Uname,
						Email:      vreg.Email,
						Password:   ud.User.Password,
						Userrole:   ud.User.Userrole,
						Facebookid: 0,
					}
					ud.User = user
					if userConnection.UpdateUser(user) == false {
						ud.Errors["Server"] = "Failed to update user."
						switch req.FormValue("if_profile") {
						case "0":
							render(res, "edit-user.gohtml", ud)
						case "1":
							render(res, "profile.gohtml", ud)
						}
						return
					}
				}

				ud = updateViewData(req, ud)

				ud.Errors["Success"] = "User Updated."
				if req.FormValue("if_profile") == "1" {
					switch req.FormValue("if_profile") {
					case "0":
						render(res, "edit-user.gohtml", ud)
					case "1":
						render(res, "profile.gohtml", ud)
					}
					return
				}
				switch req.FormValue("if_profile") {
				case "0":
					render(res, "edit-user.gohtml", ud)
				case "1":
					render(res, "profile.gohtml", ud)
				}
				return
			case "UPDATE-PASSWORD":
				pass := &VPassword{
					Password:   req.FormValue("password"),
					RePassword: req.FormValue("rePassword"),
				}
				if pass.ValidatePassword() == false {
					ud.Errors = pass.Errors
					switch req.FormValue("if_profile") {
					case "0":
						render(res, "edit-user.gohtml", ud)
					case "1":
						render(res, "profile.gohtml", ud)
					}
					return
				}

				if userConnection.UpdatePassword(pass.Password, req.FormValue("ID")) == false {
					ud.Errors["Server"] = "Failed to update user password."
					switch req.FormValue("if_profile") {
					case "0":
						render(res, "edit-user.gohtml", ud)
					case "1":
						render(res, "profile.gohtml", ud)
					}
					return
				}

				ud.Errors["Success"] = "Password Updated."
				switch req.FormValue("if_profile") {
				case "0":
					render(res, "edit-user.gohtml", ud)
				case "1":
					render(res, "profile.gohtml", ud)
				}
				return
			case "UPDATE-STATUS":
				if userConnection.UpdateStatus(req.FormValue("status"), req.FormValue("ID")) == false {
					ud.Errors["Server"] = "Failed to update user password."
					render(res, "user-manager.gohtml", ud)
					return
				}
				ausers, err := userConnection.GetUsers(1, 10)
				if err != nil {
					http.Redirect(res, req, "/admin/users", http.StatusServiceUnavailable)
					return
				}

				ud.Users = ausers

				ud.Errors["Success"] = "Status Updated."
				render(res, "user-manager.gohtml", ud)
				return
			case "VIEW":
				user, err := userConnection.GetUserByID(req.FormValue("ID"))
				ud.User = user
				ud.User.Userrole = 1
				if err != nil {
					http.Redirect(res, req, "/admin/users", http.StatusServiceUnavailable)
					return
				}
				render(res, "view-user.gohtml", ud)
				return
			}
		}

	}
	render(res, "index.gohtml", ud)
}

func postManager(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if ud.IfLoggedIn == true {
		db, err := config.GetMSSQLDB()
		if err != nil {
			ud.Errors["Server"] = "Could not connect to database."
			render(res, "post-manager.gohtml", ud)
			return
		}
		postConnection := posts.PostConnection{
			Db: db,
		}
		aposts, err := postConnection.GetPosts(1, 10)
		if err != nil {
			ud.Errors["Server"] = "Failed to retreive posts."
			render(res, "post-manager.gohtml", ud)
			return
		}
		ud.Posts = aposts

		render(res, "post-manager.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func handlePost(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if ud.IfLoggedIn == true {
		if req.Method == http.MethodPost {
			db, err := config.GetMSSQLDB()
			if err != nil {
				http.Redirect(res, req, "/admin/posts", http.StatusServiceUnavailable)
				return
			}
			// create a function that gets the posts to pass to the return page
			// also delete the image from the users folder
			postConnection := posts.PostConnection{
				Db: db,
			}
			method := req.FormValue("method")
			switch method {
			case "DELETE":
				if postConnection.DeletePost(req.FormValue("ID")) == false {
					http.Redirect(res, req, "/admin/posts", http.StatusServiceUnavailable)
					return
				}
				aposts, err := postConnection.GetPosts(1, 10)
				if err != nil {
					http.Redirect(res, req, "/admin/posts", http.StatusServiceUnavailable)
					return
				}

				ud.Posts = aposts

				ud.Errors["Success"] = "Post Deleted."
				render(res, "post-manager.gohtml", ud)
				return
			case "UPDATE":
				fname := ""
				mf, fh, err := req.FormFile("imgfile")
				if fh != nil {
					if err != nil {
						http.Redirect(res, req, "/posts/edit", http.StatusServiceUnavailable)
						return
					}
					defer mf.Close()

					ext := strings.Split(fh.Filename, ".")[1]
					h := sha1.New()
					io.Copy(h, mf)
					fname = fmt.Sprintf("%x", h.Sum(nil)) + "." + ext

					wd, err := os.Getwd()
					if err != nil {
						http.Redirect(res, req, "/posts/edit", http.StatusServiceUnavailable)
						return
					}

					newpath := filepath.Join(wd, "public", "images", "uploads", req.FormValue("ID"))
					if _, err := os.Stat(newpath); os.IsNotExist(err) {
						os.MkdirAll(newpath, os.ModePerm)
					}

					path := filepath.Join(wd, "public", "images", "uploads", req.FormValue("ID"), fname)
					nf, err := os.Create(path)
					if err != nil {
						http.Redirect(res, req, "/posts/edit", http.StatusServiceUnavailable)
						return
					}
					defer nf.Close()

					mf.Seek(0, 0)
					io.Copy(nf, mf)
				}
				postid, err := strconv.Atoi(req.FormValue("postid"))
				if err != nil {
					http.Redirect(res, req, "/posts/edit", http.StatusServiceUnavailable)
					return
				}

				vpost := &VPosts{
					Title: req.FormValue("title"),
					Body:  req.FormValue("body"),
				}
				if vpost.ValidatePost() == false {
					ud.Errors = vpost.Errors
					render(res, "edit-post.gohtml", ud)
					return
				}
				if fname == "" {
					apost := entities.Post{
						ID:       postid,
						UserUUID: ud.User.UUID,
						Title:    vpost.Title,
						Body:     vpost.Body,
					}
					ud.Post = apost
					if postConnection.UpdatePost(apost) == false {
						ud.Errors["Server"] = "Post failed to Updated."
						render(res, "edit-post.gohtml", ud)
						return
					}
				} else {
					apost := entities.Post{
						ID:       postid,
						UserUUID: ud.User.UUID,
						Image:    fname,
						Title:    vpost.Title,
						Body:     vpost.Body,
					}
					ud.Post = apost
					if postConnection.UpdatePostImage(apost) == false {
						ud.Errors["Server"] = "Post failed to Updated."
						render(res, "edit-post.gohtml", ud)
						return
					}
				}

				ud.Errors["Success"] = "Post Updated."
				render(res, "edit-post.gohtml", ud)
				return
			case "UPDATE-STATUS":
				if postConnection.UpdateStatus(req.FormValue("status"), req.FormValue("ID")) == false {
					ud.Errors["Server"] = "Failed to update post status."
					render(res, "edit-post.gohtml", ud)
					return
				}
				aposts, err := postConnection.GetPosts(1, 10)
				if err != nil {
					http.Redirect(res, req, "/admin/posts", http.StatusServiceUnavailable)
					return
				}

				ud.Posts = aposts

				ud.Errors["Success"] = "Status Updated."
				render(res, "post-manager.gohtml", ud)
				return
			case "VIEW":
				post, err := postConnection.GetPostByID(req.FormValue("ID"))
				ud.Post = post
				if err != nil {
					http.Redirect(res, req, "/admin/posts", http.StatusServiceUnavailable)
					return
				}
				render(res, "view-post.gohtml", ud)
				return
			}
		}
	}
	render(res, "index.gohtml", ud)
}

func editPost(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		db, err := config.GetMSSQLDB()
		if err != nil {
			http.Redirect(res, req, "/admin/posts", http.StatusServiceUnavailable)
			return
		}
		// create a function that gets the posts to pass to the return page
		// also delete the image from the users folder
		postConnection := posts.PostConnection{
			Db: db,
		}
		post, err := postConnection.GetPostByID(req.FormValue("ID"))
		ud.Post = post
		if err != nil {
			http.Redirect(res, req, "/admin/posts", http.StatusServiceUnavailable)
			return
		}
		render(res, "edit-post.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func createPost(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if ud.IfLoggedIn == true {
		post := entities.Post{
			Image:     "",
			Title:     "",
			Body:      "",
			CreatedAt: "",
		}
		ud.Post = post
		if req.Method == http.MethodPost {
			fname := ""
			mf, fh, err := req.FormFile("imgfile")
			if fh != nil {
				if err != nil {
					render(res, "create-post.gohtml", ud)
					return
				}
				defer mf.Close()

				ext := strings.Split(fh.Filename, ".")[1]
				h := sha1.New()
				io.Copy(h, mf)
				fname = fmt.Sprintf("%x", h.Sum(nil)) + "." + ext

				wd, err := os.Getwd()
				if err != nil {
					render(res, "create-post.gohtml", ud)
					return
				}

				newpath := filepath.Join(wd, "public", "images", "uploads", req.FormValue("ID"))
				fmt.Println(newpath, ud.User.ID)
				if _, err := os.Stat(newpath); os.IsNotExist(err) {
					os.MkdirAll(newpath, os.ModePerm)
				}

				path := filepath.Join(wd, "public", "images", "uploads", req.FormValue("ID"), fname)
				nf, err := os.Create(path)
				if err != nil {
					render(res, "create-post.gohtml", ud)
					return
				}
				defer nf.Close()

				mf.Seek(0, 0)
				io.Copy(nf, mf)
			}

			vpost := &VPosts{
				Title: req.FormValue("title"),
				Body:  req.FormValue("body"),
			}
			if vpost.ValidatePost() == false {
				ud.Errors = vpost.Errors
				render(res, "create-post.gohtml", ud)
				return
			}

			apost := entities.Post{
				UserUUID: ud.User.UUID,
				Image:    fname,
				Title:    vpost.Title,
				Body:     vpost.Body,
			}

			db, err := config.GetMSSQLDB()
			if err != nil {
				ud.Errors["Server"] = "Could not connect to database."
				render(res, "create-post.gohtml", ud)
				return
			}
			postConnection := posts.PostConnection{
				Db: db,
			}
			if postConnection.CreatePost(apost) == false {
				ud.Errors["Server"] = "Failed to create entry."
				render(res, "create-post.gohtml", ud)
				return
			}

			http.Redirect(res, req, "/admin/posts", http.StatusServiceUnavailable)
			return
		}
		render(res, "create-post.gohtml", ud)
		return
	}

	render(res, "index.gohtml", ud)
}

func commentManager(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "comment-manager.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func comments(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "comments.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func profile(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "profile.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func goManager(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "go-manager.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func sqlManager(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "sql-manager.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func index(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if req.Method == http.MethodPost {
		ud.NEmail = req.FormValue("nemail")
		news := &VNewsletter{
			NEmail: ud.NEmail,
		}
		if news.ValidateNewsletter() == false {
			ud.Errors = news.Errors
			render(res, "index.gohtml", ud)
			return
		}
		db, err := config.GetMSSQLDB()
		if err != nil {
			news.Errors["Server"] = "Could not connect to database."
			render(res, "index.gohtml", ud)
			return
		}
		newsletterConnection := newsletters.NewsletterConnection{
			Db: db,
		}
		if newsletterConnection.CreateNewsletter(ud.NEmail) == false {
			ud.Errors["Server"] = "Failed to create entry."
			render(res, "index.gohtml", ud)
			return
		}
		ud.Errors["Success"] = "Thank you for signing up. We appreciate your business."
	}
	render(res, "index.gohtml", ud)
}

func about(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	render(res, "about.gohtml", ud)
}

func contact(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Message = message{
		Name:    "",
		Email:   "",
		Subject: "",
		Message: "",
	}
	ud.Errors = make(map[string]string)
	if req.Method == http.MethodPost {
		msg := &Message{
			Name:    req.FormValue("name"),
			Email:   req.FormValue("email"),
			Subject: req.FormValue("subject"),
			Message: req.FormValue("message"),
		}
		if msg.ValidateMessage() == false {
			ud.Errors = msg.Errors
			ud.Message = message{
				Name:    "",
				Email:   "",
				Subject: "",
				Message: "",
			}
			render(res, "contact.gohtml", ud)
			return
		}
		if err := msg.Deliver(); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		render(res, "contact.gohtml", ud)
		return
	}
	render(res, "contact.gohtml", ud)
}

func email(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "email-manager.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func logout(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:     "__ibes_",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Secure:   false,
		HttpOnly: true,
	})
	ud := userData{}
	ud.NEmail = ""
	ud.IfLoggedIn = false
	render(res, "index.gohtml", ud)
}

func login(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if ud.IfLoggedIn == true {
		render(res, "admin.gohtml", ud)
		return
	}
	if req.Method == http.MethodPost {
		lgn := &Login{
			Uname:    req.FormValue("uname"),
			Password: req.FormValue("password"),
		}
		if lgn.ValidateLogin() == false {
			ud.Errors = lgn.Errors
			render(res, "login.gohtml", ud)
			return
		}

		ud.User.Uname = lgn.Uname
		ud.User.Password = lgn.Password
		db, err := config.GetMSSQLDB()
		if err != nil {
			ud.Errors["Server"] = "Could not connect to database."
			render(res, "login.gohtml", ud)
			return
		}
		userConnection := users.UserConnection{
			Db: db,
		}
		user, err := userConnection.CheckLoginForm(ud.User.Uname)
		if err != nil {
			ud.Errors["Server"] = "Failed to validate user."
			render(res, "login.gohtml", ud)
			return
		}
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(ud.User.Password)); err != nil {
			ud.Errors["Server"] = "Failed to validate user."
			render(res, "login.gohtml", ud)
			return
		}

		uuidSess, err := uuid.NewUUID()
		if err != nil {
			ud.Errors["Server"] = "Failed to create UUID."
			render(res, "login.gohtml", ud)
			return
		}
		http.SetCookie(res, &http.Cookie{
			Name:     "__ibes_",
			Value:    uuidSess.String(),
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
		})
		usersess := entities.Session{
			UUID:     uuidSess.String(),
			UserUUID: user.UUID,
		}
		userSession := users.UserSession{
			Db: db,
		}
		sess, err := userSession.CreateSession(usersess)
		_ = sess
		if err != nil {
			ud.Errors["Server"] = "Failed to create a session."
			render(res, "login.gohtml", ud)
			return
		}
		ud.User = user
		ud.IfLoggedIn = true
		vwd := userData{
			User:       user,
			IfLoggedIn: true,
		}
		viewData[sess.UUID] = vwd
		render(res, "admin.gohtml", ud)
		return
	}
	render(res, "login.gohtml", ud)
}

func register(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if ud.IfLoggedIn == true {
		render(res, "admin.gohtml", ud)
		return
	}

	if req.Method == http.MethodPost {
		vreg := &Register{
			Fname:      req.FormValue("fname"),
			Lname:      req.FormValue("lname"),
			Uname:      req.FormValue("uname"),
			Email:      req.FormValue("email"),
			Password:   req.FormValue("password"),
			RePassword: req.FormValue("rePassword"),
		}
		if vreg.ValidateRegister() == false {
			ud.Errors = vreg.Errors
			render(res, "register.gohtml", ud)
			return
		}

		uuidreg, err := uuid.NewUUID()
		if err != nil {
			ud.Errors["Server"] = "Failed to create registration UUID."
			render(res, "register.gohtml", ud)
			return
		}
		db, err := config.GetMSSQLDB()
		if err != nil {
			ud.Errors["Server"] = "Could not connect to database."
			render(res, "register.gohtml", ud)
			return
		}
		user := entities.User{
			UUID:     uuidreg.String(),
			Fname:    vreg.Fname,
			Lname:    vreg.Lname,
			Uname:    vreg.Uname,
			Email:    vreg.Email,
			Password: vreg.Password,
			Userrole: 2,
		}
		userConnection := users.UserConnection{
			Db: db,
		}
		if userConnection.CreateUser(user) == false {
			ud.Errors["Server"] = "Failed to create user."
			render(res, "register.gohtml", ud)
			return
		}

		uuidSess, err := uuid.NewUUID()
		if err != nil {
			ud.Errors["Server"] = "Failed to create session UUID."
			render(res, "register.gohtml", ud)
			return
		}
		http.SetCookie(res, &http.Cookie{
			Name:     "__ibes_",
			Value:    uuidSess.String(),
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
		})
		usersess := entities.Session{
			UUID:     uuidSess.String(),
			UserUUID: user.UUID,
		}
		userSession := users.UserSession{
			Db: db,
		}
		sess, err := userSession.CreateSession(usersess)
		_ = sess
		if err != nil {
			ud.Errors["Server"] = "Failed to create a session."
			render(res, "register.gohtml", ud)
			return
		}
		vwd := userData{
			User:       user,
			IfLoggedIn: true,
		}
		viewData[uuidSess.String()] = vwd
		ud.IfLoggedIn = true
		ud.User = user
		render(res, "admin.gohtml", ud)
		return
	}
	render(res, "register.gohtml", ud)
}

func admin(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	fmt.Println(ud.User.Image)
	if ud.IfLoggedIn == true {
		render(res, "admin.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func forgotPassword(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	ud.Errors = make(map[string]string)
	if ud.IfLoggedIn == true {
		render(res, "admin.gohtml", ud)
		return
	}
	if req.Method == http.MethodPost {
		fp := &ForgotPassword{
			Email: req.FormValue("email"),
		}
		if fp.ValidateForgotPassword() == false {
			render(res, "forgot-password.gohtml", fp)
			return
		}

		db, err := config.GetMSSQLDB()
		if err != nil {
			ud.Errors["Server"] = "Could not connect to database."
			render(res, "forgot-password.gohtml", ud)
			return
		}
		userConnection := users.UserConnection{
			Db: db,
		}
		if userConnection.CheckEmailForgotPassword(fp.Email) == false {
			ud.Errors["Success"] = "Please check for an email to reset you password."
			render(res, "forgot-password.gohtml", ud)
			return
		}
		ud.Errors["Success"] = "Please check for an email to reset you password."
	}
	render(res, "forgot-password.gohtml", ud)
}

func updateViewData(req *http.Request, ud userData) userData {
	c, err := req.Cookie("__ibes_")
	_ = c

	if err != nil {
		return userData{}
	}
	viewData[c.Value] = ud
	vud := viewData[c.Value]
	return vud
}

// doubles as ifLoggedIn and return userData object to be loaded with data from posts and comments
func ifLoggedIn(req *http.Request) userData {
	c, err := req.Cookie("__ibes_")
	_ = c

	if err != nil {
		return userData{}
	}
	ud := viewData[c.Value]
	return ud
}

func render(res http.ResponseWriter, filename string, data interface{}) {
	if err := tpl.ExecuteTemplate(res, filename, data); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
