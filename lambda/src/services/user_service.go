package services

import (
	"errors"
	"ledger-lambda/src/models"
	"ledger-lambda/src/repository"

	"log"
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
	ValidatePassword(user *models.User, password string) error
}

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

func (s *userService) ValidatePassword(user *models.User, password string) error {
	log.Printf("ValidatePassword çağrıldı: %s", user.Email)

	if user.PasswordHash == "" {
		log.Printf("Şifre hash'i boş: %s", user.Email)
		return errors.New("password is empty")
	}

	log.Printf("Şifre doğrulaması yapılıyor. Düz metin karşılaştırması.")

	if user.PasswordHash != password {
		log.Printf("Şifre doğrulama hatası: Şifreler eşleşmiyor")
		return errors.New("invalid password")
	}

	log.Printf("Şifre başarıyla doğrulandı: %s", user.Email)
	return nil
}
