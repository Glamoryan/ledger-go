package services

import (
	"Ledger/src/entities"
	"Ledger/src/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(name string) (*entities.User, error) {
	user := &entities.User{Name: name}
	err := s.repo.Create(user)

	return user, err
}

func (s *UserService) GetUserByID(id uint) (*entities.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) GetAllUsers() ([]entities.User, error) {
	return s.repo.GetAll()
}
