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

	http.ListenAndServe(":8080", router)
}
