package repository

import (
	"Ledger/src/models"
)

type UserRepository interface {
	Create(user *models.User) error
	GetAll() ([]models.User, error)
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetUserCredit(userID uint) (float64, error)
	UpdateCredit(userID uint, newAmount float64) error
	GetAllCredits() (map[uint]float64, error)
	SendCreditToUser(senderID, receiverID uint, amount float64) error
	LogTransaction(senderID, receiverID uint, amount, senderCreditBefore, receiverCreditBefore float64) error
	GetTransactionLogsBySenderAndDate(senderID uint, date string) ([]models.TransactionLog, error)
	GetMultipleUserCredits(userIDs []uint) (map[uint]float64, error)
	ProcessBatchCreditUpdate(transactions []models.BatchCreditTransaction) []models.BatchTransactionResult
}
