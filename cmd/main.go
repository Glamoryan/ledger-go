package main

import (
	"Ledger/pkg/db"
	"Ledger/src/factory"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	database := db.ConnectDB()
	appFactory := factory.NewFactory(database)

	userHandler := appFactory.NewUserHandler()
	authMiddleware := appFactory.NewAuthMiddleware()

	router := mux.NewRouter()

	// Public endpoints
	router.HandleFunc("/users/add-user", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/login", userHandler.Login).Methods("POST")

	// Protected endpoints
	router.HandleFunc("/users/get-credit", authMiddleware.Authenticate(userHandler.GetCredit)).Methods("GET")
	router.HandleFunc("/users/send-credit", authMiddleware.Authenticate(userHandler.SendCredit)).Methods("POST")
	router.HandleFunc("/users/transaction-logs/sender", authMiddleware.Authenticate(userHandler.GetTransactionLogsBySenderAndDate)).Methods("GET")

	// Admin only endpoints
	router.HandleFunc("/users", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.GetAllUsers))).Methods("GET")
	router.HandleFunc("/users/get-user", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.GetUserByID))).Methods("GET")
	router.HandleFunc("/users/add-credit", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.AddCredit))).Methods("POST")
	router.HandleFunc("/users/credits", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.GetAllCredits))).Methods("GET")
	router.HandleFunc("/users/batch/credits", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.GetMultipleUserCredits))).Methods("POST")
	router.HandleFunc("/users/batch/update-credits", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.ProcessBatchCreditUpdate))).Methods("POST")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
