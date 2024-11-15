package main

import (
	"Ledger/pkg/db"
	"Ledger/src/handlers"
	"Ledger/src/repository"
	"Ledger/src/services"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	database := db.ConnectDB()
	userRepo := repository.NewUserRepository(database)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	router := mux.NewRouter()
	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")

	http.ListenAndServe(":8080", router)
}
