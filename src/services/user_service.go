package services

import (
	"Ledger/src/entities"
	"Ledger/src/repository"
)

type UserService interface {
	CreateUser(name, surname string, age int) (*entities.User, error)
	GetUserByID(id uint) (*entities.User, error)
	GetAllUsers() ([]entities.User, error)
	AddCredit(id uint, amount float64) error
	GetCredit(id uint) (float64, error)
	GetAllCredits() ([]map[string]interface{}, error)
	SendCredit(senderId, receiverId uint, amount float64) error
}

type userService struct {
	repo repository.UserRepository
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
	return s.repo.SendCreditToUser(senderId, receiverId, amount)
}
