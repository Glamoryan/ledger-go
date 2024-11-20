package main

import (
	"Ledger/pkg/db"
	"Ledger/src/factory"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	database := db.ConnectDB()
	appFactory := factory.NewFactory(database)

	userHandler := appFactory.NewUserHandler()

	router := mux.NewRouter()
	router.HandleFunc("/users/add-user", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/users/get-user", userHandler.GetUserByID).Methods("GET")
	router.HandleFunc("/users/add-credit", userHandler.AddCredit).Methods("POST")
	router.HandleFunc("/users/get-credit", userHandler.GetCredit).Methods("GET")
	router.HandleFunc("/users/credits", userHandler.GetAllCredits).Methods("GET")

	http.ListenAndServe(":8080", router)
}
