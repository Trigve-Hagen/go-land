package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*/*/*.gohtml"))
	for _, t := range tpl.Templates() {
		fmt.Println(t.Name())
	}
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("public"))))
	http.ListenAndServe(":8080", nil)
}

func index(res http.ResponseWriter, req *http.Request) {
	err := tpl.ExecuteTemplate(res, "index.gohtml", nil)
	if err != nil {
		log.Fatalln("template didn't execute: ", err)
	}
}

/*func homePage(w http.ResponseWriter, req *http.Request) {
	db, err := config.GetMSSQLDB()
	if err != nil {
		fmt.Println(err)
	} else {
		userConnection := users.UserConnection{
			Db: db,
		}
		fmt.Println("User List")
		users, err2 := userConnection.GetUsers()
		if err2 != nil {
			fmt.Println(err2)
		} else {
			fmt.Print("Users: ", len(users), "\n")
			for _, user := range users {
				fmt.Println("ID:", user.ID)
				fmt.Println("Name:", user.Name)
				fmt.Println("Email:", user.Email)
				fmt.Println("Age:", user.Age)
				fmt.Println("Gender:", user.Gender)
				fmt.Println("----------------------------")
			}
		}
	}
}*/
