package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)

	fmt.Println("Server Starting....")
	http.ListenAndServe(":8080", mux)
}

func homePage(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Home Page"))
}
