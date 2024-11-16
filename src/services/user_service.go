package services

import (
	"Ledger/src/entities"
	"Ledger/src/repository"
)

type UserService interface {
	CreateUser(name, surname string, age int) (*entities.User, error)
	GetUserByID(id uint) (*entities.User, error)
	GetAllUsers() ([]entities.User, error)
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
