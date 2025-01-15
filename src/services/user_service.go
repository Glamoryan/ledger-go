package services

import (
	"Ledger/src/models"
)

type UserService interface {
	CreateUser(user *models.User) error
	GetAllUsers() ([]models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserCredit(userID uint) (float64, error)
	SendCredit(senderID, receiverID uint, amount float64) error
	GetTransactionLogsBySenderAndDate(senderID uint, date string) ([]models.TransactionLog, error)
	AddCredit(userID uint, amount float64) error
	GetAllCredits() ([]models.User, error)
	GetMultipleUserCredits(userIDs []uint) ([]models.User, error)
	ProcessBatchCreditUpdate(transactions []models.BatchTransaction) []models.BatchTransactionResult
}
