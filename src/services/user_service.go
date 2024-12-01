package services

import (
	"Ledger/src/entities"
	"Ledger/src/repository"
	"errors"
)

type UserService interface {
	CreateUser(name, surname string, age int) (*entities.User, error)
	GetUserByID(id uint) (*entities.User, error)
	GetAllUsers() ([]entities.User, error)
	AddCredit(id uint, amount float64) error
	GetCredit(id uint) (float64, error)
	GetAllCredits() ([]map[string]interface{}, error)
	SendCredit(senderId, receiverId uint, amount float64) error
	GetTransactionLogsBySenderAndDate(senderId uint, date string) ([]entities.TransactionLog, error)
}

type userService struct {
	repo repository.UserRepository
}

func (s *userService) GetTransactionLogsBySenderAndDate(senderId uint, date string) ([]entities.TransactionLog, error) {
	return s.repo.GetTransactionLogsBySenderAndDate(senderId, date)
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(name, surname string, age int) (*entities.User, error) {
	user := &entities.User{
		Name:    name,
		Surname: surname,
		Age:     age,
	}
	err := s.repo.Create(user)
	return user, err
}

func (s *userService) GetUserByID(id uint) (*entities.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetAllUsers() ([]entities.User, error) {
	return s.repo.GetAll()
}

func (s *userService) AddCredit(id uint, amount float64) error {
	currentCredit, err := s.repo.GetUserCredit(id)
	if err != nil {
		return err
	}
	return s.repo.UpdateCredit(id, currentCredit+amount)
}

func (s *userService) GetCredit(id uint) (float64, error) {
	return s.repo.GetUserCredit(id)
}

func (s *userService) GetAllCredits() ([]map[string]interface{}, error) {
	return s.repo.GetAllCredits()
}

func (s *userService) SendCredit(senderId, receiverId uint, amount float64) error {
	senderCredit, err := s.repo.GetUserCredit(senderId)
	if err != nil {
		return err
	}

	if senderCredit < amount {
		return errors.New("insufficient credit")
	}

	receiverCredit, err := s.repo.GetUserCredit(receiverId)
	if err != nil {
		return err
	}

	if err := s.repo.SendCreditToUser(senderId, receiverId, amount); err != nil {
		return err
	}

	if err := s.repo.LogTransaction(senderId, receiverId, amount, senderCredit, receiverCredit); err != nil {
		return err
	}

	return nil
}
