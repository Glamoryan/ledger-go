package services

import (
	"Ledger/src/entities"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(name, surname string, age int) (*entities.User, error) {
	args := m.Called(name, surname, age)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uint) (*entities.User, error) {
	args := m.Called(id)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers() ([]entities.User, error) {
	args := m.Called()
	return args.Get(0).([]entities.User), args.Error(1)
}
