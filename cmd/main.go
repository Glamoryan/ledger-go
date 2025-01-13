package main

import (
	"Ledger/pkg/db"
	"Ledger/src/factory"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	database := db.ConnectDB()
	appFactory := factory.NewFactory(database)

	userHandler := appFactory.NewUserHandler()
	authMiddleware := appFactory.NewAuthMiddleware()

	router := mux.NewRouter()

	router.HandleFunc("/users/add-user", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/login", userHandler.Login).Methods("POST")

	router.HandleFunc("/users", authMiddleware.Authenticate(userHandler.GetAllUsers)).Methods("GET")
	router.HandleFunc("/users/get-user", authMiddleware.Authenticate(userHandler.GetUserByID)).Methods("GET")
	router.HandleFunc("/users/get-credit", authMiddleware.Authenticate(userHandler.GetCredit)).Methods("GET")
	router.HandleFunc("/users/send-credit", authMiddleware.Authenticate(userHandler.SendCredit)).Methods("POST")
	router.HandleFunc("/users/send-credit-async", authMiddleware.Authenticate(userHandler.SendCreditAsync)).Methods("POST")
	router.HandleFunc("/users/transaction-logs/sender", authMiddleware.Authenticate(userHandler.GetTransactionLogsBySenderAndDate)).Methods("GET")

	router.HandleFunc("/users/add-credit", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.AddCredit))).Methods("POST")
	router.HandleFunc("/users/credits", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.GetAllCredits))).Methods("GET")

	router.HandleFunc("/users/batch/credits", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.GetMultipleUserCredits))).Methods("POST")
	router.HandleFunc("/users/batch/update-credits", authMiddleware.Authenticate(authMiddleware.AdminOnly(userHandler.ProcessBatchCreditUpdate))).Methods("POST")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		return
	}
}
