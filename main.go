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

type viewData struct {
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
	vwd := viewData{}
	vwd.IfLoggedIn = ifLoggedIn(req)
	render(res, "index.gohtml", vwd)
}

func about(res http.ResponseWriter, req *http.Request) {
	vwd := viewData{}
	vwd.IfLoggedIn = ifLoggedIn(req)
	render(res, "about.gohtml", vwd)
}

func contact(res http.ResponseWriter, req *http.Request) {
	msg := &Message{
		Name:       "",
		Email:      "",
		Subject:    "",
		Message:    "",
		IfLoggedIn: ifLoggedIn(req),
	}
	if req.Method == http.MethodPost {
		msg.Name = req.FormValue("name")
		msg.Email = req.FormValue("email")
		msg.Subject = req.FormValue("subject")
		msg.Message = req.FormValue("message")

		if msg.ValidateMessage() == false {
			render(res, "contact.gohtml", msg)
			return
		}
		if err := msg.Deliver(); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		render(res, "contact.gohtml", msg)
		return
	}
	render(res, "contact.gohtml", msg)
}

func email(res http.ResponseWriter, req *http.Request) {
	vwd := viewData{}
	vwd.IfLoggedIn = ifLoggedIn(req)
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
	vwd := viewData{}
	vwd.IfLoggedIn = false
	render(res, "index.gohtml", vwd)
}

func login(res http.ResponseWriter, req *http.Request) {
	lgn := &Login{
		Uname:      "",
		Password:   "",
		IfLoggedIn: ifLoggedIn(req),
	}
	if lgn.IfLoggedIn == true {
		render(res, "admin.gohtml", lgn)
		return
	}

	if req.Method == http.MethodPost {
		lgn := &Login{
			Uname:    req.FormValue("uname"),
			Password: req.FormValue("password"),
		}
		if lgn.ValidateLogin() == false {
			lgn.IfLoggedIn = false
			render(res, "login.gohtml", lgn)
			return
		}
		db, err := config.GetMSSQLDB()
		if err != nil {
			lgn.IfLoggedIn = false
			lgn.Errors["Server"] = "Could not connect to database."
			render(res, "login.gohtml", lgn)
			return
		}
		uuid, err := uuid.NewUUID()
		if err != nil {
			lgn.IfLoggedIn = false
			lgn.Errors["Server"] = "Failed to create UUID."
			render(res, "login.gohtml", lgn)
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
		if userConnection.CheckLoginForm(lgn.Uname, lgn.Password) {
			lgn.IfLoggedIn = true
			render(res, "admin.gohtml", lgn)
			return
		}
	}
	render(res, "login.gohtml", lgn)
}

func register(res http.ResponseWriter, req *http.Request) {
	vreg := &Register{
		Fname:      "",
		Lname:      "",
		Uname:      "",
		Email:      "",
		Password:   "",
		RePassword: "",
		IfLoggedIn: ifLoggedIn(req),
	}
	if vreg.IfLoggedIn == true {
		render(res, "admin.gohtml", vreg)
		return
	}

	if req.Method == http.MethodPost {
		vreg.Fname = req.FormValue("fname")
		vreg.Lname = req.FormValue("lname")
		vreg.Uname = req.FormValue("uname")
		vreg.Email = req.FormValue("email")
		vreg.Password = req.FormValue("password")
		vreg.RePassword = req.FormValue("rePassword")

		if vreg.ValidateRegister() == false {
			render(res, "register.gohtml", vreg)
			return
		}

		password := []byte(vreg.Password)
		hPass := hashAndSalt(password)
		uuidreg, err := uuid.NewUUID()
		if err != nil {
			vreg.Errors["Server"] = "Failed to create registration UUID."
			render(res, "register.gohtml", vreg)
			return
		}

		us := entities.User{
			UUID:     uuidreg.String(),
			Fname:    vreg.Fname,
			Lname:    vreg.Lname,
			Uname:    vreg.Uname,
			Email:    vreg.Email,
			Password: hPass,
			Role:     1,
		}
		db, err := config.GetMSSQLDB()
		if err != nil {
			vreg.Errors["Server"] = "Could not connect to database."
			render(res, "register.gohtml", vreg)
			return
		}
		uuid, err := uuid.NewUUID()
		if err != nil {
			vreg.Errors["Server"] = "Failed to create session UUID."
			render(res, "register.gohtml", vreg)
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

		render(res, "/auth/admin", vreg)
		return
	}
	render(res, "register.gohtml", vreg)
}

func admin(res http.ResponseWriter, req *http.Request) {
	vwd := viewData{}
	if ifLoggedIn(req) == true {
		vwd.IfLoggedIn = true
		render(res, "admin.gohtml", vwd)
		return
	}
	render(res, "index.gohtml", vwd)
}

func forgotPassword(res http.ResponseWriter, req *http.Request) {
	vwd := viewData{}
	if ifLoggedIn(req) == true {
		vwd.IfLoggedIn = true
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

func ifLoggedIn(req *http.Request) bool {
	c, err := req.Cookie("__ibes_")
	_ = c

	if err != nil {
		return false
	}

	return true
}

func render(res http.ResponseWriter, filename string, data interface{}) {
	if err := tpl.ExecuteTemplate(res, filename, data); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
