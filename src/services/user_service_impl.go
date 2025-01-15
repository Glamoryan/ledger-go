package services

import (
	"Ledger/src/models"
	"Ledger/src/repository"
)

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
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

func (s *userService) GetUserCredit(userID uint) (float64, error) {
	return s.repo.GetUserCredit(userID)
}

func (s *userService) SendCredit(senderID, receiverID uint, amount float64) error {
	return s.repo.SendCredit(senderID, receiverID, amount)
}

func (s *userService) GetTransactionLogsBySenderAndDate(senderID uint, date string) ([]models.TransactionLog, error) {
	return s.repo.GetTransactionLogsBySenderAndDate(senderID, date)
}

func (s *userService) AddCredit(userID uint, amount float64) error {
	return s.repo.AddCredit(userID, amount)
}

func (s *userService) GetAllCredits() ([]models.User, error) {
	return s.repo.GetAllCredits()
}

func (s *userService) GetMultipleUserCredits(userIDs []uint) ([]models.User, error) {
	return s.repo.GetMultipleUserCredits(userIDs)
}

func (s *userService) ProcessBatchCreditUpdate(transactions []models.BatchTransaction) []models.BatchTransactionResult {
	return s.repo.ProcessBatchCreditUpdate(transactions)
}
