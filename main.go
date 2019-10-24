package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	UUID       string
	Fname      string
	Lname      string
	Uname      string
	Email      string
	NEmail     string
	Password   string
	RePassword string
	Userrole   int8
	Message    message
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
	http.HandleFunc("/admin/posts", postManager)
	http.HandleFunc("/posts/create", createPost)
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
			mf, fh, err := req.FormFile("imgfile")
			if err != nil {
				fmt.Println("Here 1")
				render(res, "create-post.gohtml", ud)
				return
			}
			defer mf.Close()

			ext := strings.Split(fh.Filename, ".")[1]
			h := sha1.New()
			io.Copy(h, mf)
			fname := fmt.Sprintf("%x", h.Sum(nil)) + "." + ext

			wd, err := os.Getwd()
			if err != nil {
				fmt.Println("Here 2 ", fname)
				render(res, "create-post.gohtml", ud)
				return
			}

			newpath := filepath.Join(wd, "public", "images", "uploads")
			if _, err := os.Stat(newpath); os.IsNotExist(err) {
				os.MkdirAll(newpath, os.ModePerm)
			}

			path := filepath.Join(wd, "public", "images", "uploads", fname)
			nf, err := os.Create(path)
			if err != nil {
				fmt.Println("Here 3 ", fname)
				render(res, "create-post.gohtml", ud)
				return
			}
			defer nf.Close()

			mf.Seek(0, 0)
			io.Copy(nf, mf)

			vpost := &VPosts{
				Title: req.FormValue("title"),
				Body:  req.FormValue("body"),
			}
			if vpost.ValidatePost() == false {
				render(res, "create-post.gohtml", ud)
				return
			}

			apost := entities.Post{
				UserUUID: ud.UUID,
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

			ud.Post = apost

			render(res, "create-post.gohtml", ud)
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

func userManager(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "user-manager.gohtml", ud)
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
			render(res, "login.gohtml", lgn)
			return
		}

		ud.Uname = lgn.Uname
		ud.Password = lgn.Password
		db, err := config.GetMSSQLDB()
		if err != nil {
			ud.Errors["Server"] = "Could not connect to database."
			render(res, "login.gohtml", ud)
			return
		}
		userConnection := users.UserConnection{
			Db: db,
		}
		user, err := userConnection.CheckLoginForm(ud.Uname)
		if err != nil {
			ud.Errors["Server"] = "Failed to validate user."
			render(res, "login.gohtml", ud)
			return
		}
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(ud.Password)); err != nil {
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
		vwd := userData{
			UUID:       user.UUID,
			Fname:      user.Fname,
			Lname:      user.Lname,
			Uname:      user.Uname,
			Email:      user.Email,
			Password:   user.Password,
			Userrole:   user.Userrole,
			IfLoggedIn: true,
		}
		viewData[sess.UUID] = vwd
		render(res, "admin.gohtml", user)
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
			render(res, "register.gohtml", vreg)
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
		user, err = userConnection.CreateUser(user)
		if err != nil {
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
			UUID:       uuidreg.String(),
			Fname:      vreg.Fname,
			Lname:      vreg.Lname,
			Uname:      vreg.Uname,
			Email:      vreg.Email,
			Password:   vreg.Password,
			Userrole:   2,
			IfLoggedIn: true,
		}
		viewData[uuidSess.String()] = vwd
		ud.IfLoggedIn = true
		render(res, "admin.gohtml", vwd)
		return
	}
	render(res, "register.gohtml", ud)
}

func admin(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	fmt.Println(ud)
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
