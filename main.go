package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

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
	http.HandleFunc("/login", login)
	http.HandleFunc("/register", register)
	http.HandleFunc("/forgot/password", forgotPassword)
	http.HandleFunc("/auth/admin", admin)
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("public"))))
	http.ListenAndServe(":8080", nil)
}

func index(res http.ResponseWriter, req *http.Request) {
	err := tpl.ExecuteTemplate(res, "index.gohtml", nil)
	if err != nil {
		log.Fatalln("template didn't execute: ", err)
	}
}

func about(res http.ResponseWriter, req *http.Request) {
	err := tpl.ExecuteTemplate(res, "about.gohtml", nil)
	if err != nil {
		log.Fatalln("template didn't execute: ", err)
	}
}

func contact(res http.ResponseWriter, req *http.Request) {
	err := tpl.ExecuteTemplate(res, "contact.gohtml", nil)
	if err != nil {
		log.Fatalln("template didn't execute: ", err)
	}
}

func login(res http.ResponseWriter, req *http.Request) {
	if ifLoggedIn(req) {
		http.Redirect(res, req, "/auth/admin", http.StatusMovedPermanently)
		return
	}
	if req.Method == http.MethodPost {
		uName := req.FormValue("uname")
		pwd := req.FormValue("password")
		db, err := config.GetMSSQLDB()
		if err != nil {
			log.Fatalln("error connecting: ", err)
		}
		userConnection := users.UserConnection{
			Db: db,
		}
		if userConnection.CheckLogin(uName, pwd) {
			http.Redirect(res, req, "/auth/admin", http.StatusMovedPermanently)
			return
		}
	} else {
		err := tpl.ExecuteTemplate(res, "login.gohtml", nil)
		if err != nil {
			log.Fatalln("template didn't execute: ", err)
		}
	}

}

func register(res http.ResponseWriter, req *http.Request) {
	if ifLoggedIn(req) {
		http.Redirect(res, req, "/auth/admin", http.StatusMovedPermanently)
		return
	}
	c, err := req.Cookie("__ibes_")
	_ = c
	if err != nil {
		uuid, err := uuid.NewUUID()
		if err != nil {
			log.Fatalln("uuid failed: ", err)
		}
		http.SetCookie(res, &http.Cookie{
			Name:     "__ibes_",
			Value:    uuid.String(),
			Secure:   false,
			HttpOnly: true,
		})
	}
	//fmt.Println("Here 1")
	if req.Method == http.MethodPost {
		//fmt.Println("Here 2")
		fName := req.FormValue("fname")
		lName := req.FormValue("lname")
		uName := req.FormValue("uname")
		email := req.FormValue("email")
		pwd := req.FormValue("password")
		pwdConfirm := req.FormValue("rePassword")
		if pwd == pwdConfirm {
			password := []byte(pwd)
			hPass := hashAndSalt(password)
			if err != nil {
				err := tpl.ExecuteTemplate(res, "register.gohtml", err)
				if err != nil {
					fmt.Println("template didn't execute: ", err)
				}
			}
			uuidreg, err := uuid.NewUUID()
			if err != nil {
				err := tpl.ExecuteTemplate(res, "register.gohtml", err)
				if err != nil {
					fmt.Println("template didn't execute: ", err)
				}
			}
			us := entities.User{
				UUID:     uuidreg.String(),
				Fname:    fName,
				Lname:    lName,
				Uname:    uName,
				Email:    email,
				Password: hPass,
			}
			db, err := config.GetMSSQLDB()
			if err != nil {
				err := tpl.ExecuteTemplate(res, "register.gohtml", err)
				if err != nil {
					fmt.Println("template didn't execute: ", err)
				}
			} else {
				userConnection := users.UserConnection{
					Db: db,
				}
				userid := userConnection.CreateUser(us)
				fmt.Println(userid)

				userSession := sessions.UserSession{
					Db: db,
				}
				sessionid := userSession.CreateSession(us)
				fmt.Println(sessionid)
			}
			if err != nil {
				err := tpl.ExecuteTemplate(res, "register.gohtml", err)
				if err != nil {
					fmt.Println("template didn't execute: ", err)
				}
			}
			// create a user
			// create a session

			http.Redirect(res, req, "/auth/admin", http.StatusMovedPermanently)
		} else {
			err := tpl.ExecuteTemplate(res, "register.gohtml", nil)
			if err != nil {
				fmt.Println("template didn't execute: ", err)
			}
		}
	} else {
		err := tpl.ExecuteTemplate(res, "register.gohtml", nil)
		if err != nil {
			fmt.Println("template didn't execute: ", err)
		}
	}
}

func admin(res http.ResponseWriter, req *http.Request) {
	err := tpl.ExecuteTemplate(res, "admin.gohtml", nil)
	if err != nil {
		log.Fatalln("template didn't execute: ", err)
	}
}

func forgotPassword(res http.ResponseWriter, req *http.Request) {
	err := tpl.ExecuteTemplate(res, "forgot-password.gohtml", nil)
	if err != nil {
		log.Fatalln("template didn't execute: ", err)
	}
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
