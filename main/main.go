package main

import (
	"fmt"
	"net/http"

	"github.com/Trigve-Hagen/rlayouts/config"
	users "github.com/Trigve-Hagen/rlayouts/models"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)

	fmt.Println("Server Starting....")
	http.ListenAndServe(":8080", mux)
}

func homePage(w http.ResponseWriter, req *http.Request) {
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
}
