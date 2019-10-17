package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/Trigve-Hagen/rlayouts/config"
	"github.com/Trigve-Hagen/rlayouts/entities"
	sessions "github.com/Trigve-Hagen/rlayouts/models"
	users "github.com/Trigve-Hagen/rlayouts/models"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
)

var tpl *template.Template
var formErrors = map[string]string{}
var data = map[string]entities.User{}

type ViewData struct {
	IfLoggedIn bool
	Errors     map[string]string
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*/*/*.gohtml"))
	/*for _, t := range tpl.Templates() {
		fmt.Println(t.Name())
	}*/
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/about", about)
	http.HandleFunc("/contact", contact)
	http.HandleFunc("/email", email)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/register", register)
	http.HandleFunc("/forgot/password", forgotPassword)
	http.HandleFunc("/auth/admin", admin)
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("public"))))
	http.ListenAndServe(":3000", nil)
}

func index(res http.ResponseWriter, req *http.Request) {
	vwd := ViewData{}
	render(res, "index.gohtml", vwd)
}

func about(res http.ResponseWriter, req *http.Request) {
	vwd := ViewData{}
	render(res, "about.gohtml", vwd)
}

func contact(res http.ResponseWriter, req *http.Request) {
	vwd := ViewData{}
	render(res, "contact.gohtml", vwd)
}

func email(res http.ResponseWriter, req *http.Request) {
	vwd := ViewData{}
	render(res, "email.gohtml", vwd)
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
	vwd := ViewData{}
	vwd.IfLoggedIn = false
	render(res, "index.gohtml", vwd)
}

func login(res http.ResponseWriter, req *http.Request) {
	vwd := ViewData{}
	if vwd.ifLoggedIn(req) {
		render(res, "admin.gohtml", vwd)
		return
	}

	if req.Method == http.MethodPost {
		uName := req.FormValue("uname")
		pwd := req.FormValue("password")
		db, err := config.GetMSSQLDB()
		if err != nil {
			vwd.Errors["Server"] = "Could not connect to database."
			render(res, "login.gohtml", vwd)
			return
		}

		uuid, err := uuid.NewUUID()
		if err != nil {
			vwd.Errors["Server"] = "Failed to create UUID."
			render(res, "login.gohtml", vwd)
			return
		}
		http.SetCookie(res, &http.Cookie{
			Name:     "__ibes_",
			Value:    uuid.String(),
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
		})

		userConnection := users.UserConnection{
			Db: db,
		}
		if userConnection.CheckLoginForm(uName, pwd) {
			render(res, "admin.gohtml", vwd)
			return
		}
	}
	render(res, "login.gohtml", vwd)
}

func register(res http.ResponseWriter, req *http.Request) {
	vwd := ViewData{}
	if vwd.ifLoggedIn(req) {
		render(res, "admin.gohtml", vwd)
		return
	}

	if req.Method == http.MethodPost {
		fName := req.FormValue("fname")
		lName := req.FormValue("lname")
		uName := req.FormValue("uname")
		email := req.FormValue("email")
		pwd := req.FormValue("password")
		pwdConfirm := req.FormValue("rePassword")
		if pwd != pwdConfirm {
			vwd.Errors["Form"] = "Passwords do not match."
			render(res, "register.gohtml", vwd)
			return
		}
		password := []byte(pwd)
		hPass := hashAndSalt(password)
		uuidreg, err := uuid.NewUUID()
		if err != nil {
			vwd.Errors["Server"] = "Failed to create registration UUID."
			render(res, "register.gohtml", vwd)
			return
		}
		us := entities.User{
			UUID:     uuidreg.String(),
			Fname:    fName,
			Lname:    lName,
			Uname:    uName,
			Email:    email,
			Password: hPass,
			Role:     1,
		}
		db, err := config.GetMSSQLDB()
		if err != nil {
			vwd.Errors["Server"] = "Could not connect to database."
			render(res, "register.gohtml", vwd)
			return
		}
		uuid, err := uuid.NewUUID()
		if err != nil {
			vwd.Errors["Server"] = "Failed to create session UUID."
			render(res, "register.gohtml", vwd)
			return
		}
		http.SetCookie(res, &http.Cookie{
			Name:     "__ibes_",
			Value:    uuid.String(),
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
		})
		data := us
		userConnection := users.UserConnection{
			Db: db,
		}
		userid := userConnection.CreateUser(data)
		fmt.Println(userid)

		userSession := sessions.UserSession{
			Db: db,
		}
		sessionid := userSession.CreateSession(data)
		fmt.Println(sessionid)

		render(res, "/auth/admin", vwd)
	}
	render(res, "register.gohtml", vwd)
}

func admin(res http.ResponseWriter, req *http.Request) {
	vwd := ViewData{}
	if vwd.ifLoggedIn(req) {
		render(res, "admin.gohtml", vwd)
		return
	}
	render(res, "index.gohtml", vwd)
}

func forgotPassword(res http.ResponseWriter, req *http.Request) {
	vwd := ViewData{}
	if vwd.ifLoggedIn(req) {
		render(res, "admin.gohtml", vwd)
		return
	}
	render(res, "forgot-password.gohtml", vwd)
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

func (vwd *ViewData) ifLoggedIn(req *http.Request) bool {
	c, err := req.Cookie("__ibes_")
	_ = c

	if err != nil {
		vwd.IfLoggedIn = false
		return false
	}
	vwd.IfLoggedIn = true
	return true
}

func render(res http.ResponseWriter, filename string, data interface{}) {
	if err := tpl.ExecuteTemplate(res, filename, data); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
