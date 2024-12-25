package main

import (
	"Ledger/pkg/db"
	"Ledger/pkg/queue"
	"Ledger/src/factory"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	database := db.ConnectDB()
	appFactory := factory.NewFactory(database)
	
	rabbitMQ := appFactory.NewRabbitMQ()
	defer rabbitMQ.Close()

	userService := appFactory.NewUserService()

	transactionHandler := func(msg queue.TransactionMessage) error {
		return userService.SendCredit(msg.SenderID, msg.ReceiverID, msg.Amount)
	}

	rabbitMQ.ConsumeTransactions(transactionHandler)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Worker started. Press CTRL+C to exit.")
	<-sigChan
	log.Println("Shutting down worker...")
} 