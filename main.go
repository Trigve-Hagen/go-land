package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/Trigve-Hagen/rlayouts/config"
	"github.com/Trigve-Hagen/rlayouts/entities"
	users "github.com/Trigve-Hagen/rlayouts/models"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
)

type userData struct {
	IfLoggedIn bool
	UUID       string
	Fname      string
	Lname      string
	Uname      string
	Email      string
	Password   string
	Userrole   int8
}

type sessionData struct {
	UUID     string
	UserUUID string
	DateTime string
}

var tpl *template.Template
var viewData = map[string]userData{}

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
	ud := ifLoggedIn(req)
	fmt.Println(ud)
	render(res, "index.gohtml", ud)
}

func about(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	render(res, "about.gohtml", ud)
}

func contact(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	msg := &Message{
		Name:       "",
		Email:      "",
		Subject:    "",
		Message:    "",
		IfLoggedIn: false,
	}
	if ud.Fname != "" {
		msg.Fname = ud.Fname
		msg.Lname = ud.Lname
		msg.Email = ud.Email
		msg.IfLoggedIn = true
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
	ud := ifLoggedIn(req)
	render(res, "email.gohtml", ud)
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
	ud.IfLoggedIn = false
	render(res, "index.gohtml", ud)
}

func login(res http.ResponseWriter, req *http.Request) {
	//ctx := context.Background()
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "admin.gohtml", ud)
		return
	}
	lgn := &Login{
		Uname:    "",
		Password: "",
	}
	if req.Method == http.MethodPost {
		lgn.Uname = req.FormValue("uname")
		lgn.Password = req.FormValue("password")

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

		uuidSess, err := uuid.NewUUID()
		if err != nil {
			lgn.IfLoggedIn = false
			lgn.Errors["Server"] = "Failed to create UUID."
			render(res, "login.gohtml", lgn)
			return
		}
		http.SetCookie(res, &http.Cookie{
			Name:     "__ibes_",
			Value:    uuidSess.String(),
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
		})

		userConnection := users.UserConnection{
			Db: db,
		}
		user, err := userConnection.CheckLoginForm(lgn.Uname)
		if err != nil {
			lgn.IfLoggedIn = false
			lgn.Errors["Server"] = "Failed to validate user."
			render(res, "login.gohtml", lgn)
			return
		}
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(lgn.Password)); err != nil {
			lgn.IfLoggedIn = false
			lgn.Errors["Server"] = "Failed to validate user."
			render(res, "login.gohtml", lgn)
			return
		}

		//dt := time.Now().Format("2006-01-01 03:04:05")
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
			lgn.IfLoggedIn = false
			lgn.Errors["Server"] = "Failed to create a session."
			render(res, "login.gohtml", lgn)
			return
		}
		fmt.Println(sess.UUID)
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
	render(res, "login.gohtml", lgn)
}

func register(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	vreg := &Register{
		Fname:      "",
		Lname:      "",
		Uname:      "",
		Email:      "",
		Password:   "",
		RePassword: "",
		IfLoggedIn: false,
	}
	if ud.Fname != "" {
		vreg.Fname = ud.Fname
		vreg.Lname = ud.Lname
		vreg.Uname = ud.Uname
		vreg.Email = ud.Email
		vreg.IfLoggedIn = true
		vreg.Userrole = ud.Userrole
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
		vreg.Userrole = 2

		if vreg.ValidateRegister() == false {
			render(res, "register.gohtml", vreg)
			return
		}

		uuidreg, err := uuid.NewUUID()
		if err != nil {
			vreg.Errors["Server"] = "Failed to create registration UUID."
			render(res, "register.gohtml", vreg)
			return
		}
		db, err := config.GetMSSQLDB()
		if err != nil {
			vreg.Errors["Server"] = "Could not connect to database."
			render(res, "register.gohtml", vreg)
			return
		}
		uuidSess, err := uuid.NewUUID()
		if err != nil {
			vreg.Errors["Server"] = "Failed to create session UUID."
			render(res, "register.gohtml", vreg)
			return
		}
		http.SetCookie(res, &http.Cookie{
			Name:     "__ibes_",
			Value:    uuidSess.String(),
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
		})

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
			vreg.Errors["Server"] = "Failed to create user."
			render(res, "register.gohtml", vreg)
			return
		}
		//fmt.Println(user)

		//dt := time.Now().Format("2006-01-01 03:04:05")
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
			vreg.IfLoggedIn = false
			vreg.Errors["Server"] = "Failed to create a session."
			render(res, "register.gohtml", vreg)
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
		vreg.IfLoggedIn = true
		render(res, "admin.gohtml", vreg)
		return
	}
	render(res, "register.gohtml", vreg)
}

func admin(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "admin.gohtml", ud)
		return
	}
	render(res, "index.gohtml", ud)
}

func forgotPassword(res http.ResponseWriter, req *http.Request) {
	ud := ifLoggedIn(req)
	if ud.IfLoggedIn == true {
		render(res, "admin.gohtml", ud)
		return
	}
	render(res, "forgot-password.gohtml", ud)
}

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
