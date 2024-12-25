package services

import (
	"Ledger/src/models"
	"Ledger/src/repository"
	"errors"
)

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(user *models.User) error {
	return s.repo.Create(user)
}

func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.repo.GetAll()
}

func (s *userService) GetUserByID(id uint) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	return s.repo.GetByEmail(email)
}

func (s *userService) AddCredit(userID uint, amount float64) error {
	currentCredit, err := s.repo.GetUserCredit(userID)
	if err != nil {
		return err
	}
	return s.repo.UpdateCredit(userID, currentCredit+amount)
}

func (s *userService) GetUserCredit(userID uint) (float64, error) {
	return s.repo.GetUserCredit(userID)
}

func (s *userService) GetAllCredits() (map[uint]float64, error) {
	return s.repo.GetAllCredits()
}

func (s *userService) SendCredit(senderID, receiverID uint, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	return s.repo.SendCreditToUser(senderID, receiverID, amount)
}

func (s *userService) GetTransactionLogsBySenderAndDate(senderID uint, date string) ([]models.TransactionLog, error) {
	return s.repo.GetTransactionLogsBySenderAndDate(senderID, date)
}

func (s *userService) GetMultipleUserCredits(userIDs []uint) (map[uint]float64, error) {
	return s.repo.GetMultipleUserCredits(userIDs)
}

func (s *userService) ProcessBatchCreditUpdate(transactions []models.BatchCreditTransaction) []models.BatchTransactionResult {
	for _, txn := range transactions {
		if txn.Amount <= 0 {
			return []models.BatchTransactionResult{{
				Success: false,
				UserID:  txn.UserID,
				Amount:  txn.Amount,
				Error:   "Amount must be positive",
			}}
		}
	}
	
	return s.repo.ProcessBatchCreditUpdate(transactions)
}

func (s *userService) SendCreditAsync(senderID, receiverID uint, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	return s.repo.SendCreditAsync(senderID, receiverID, amount)
} 