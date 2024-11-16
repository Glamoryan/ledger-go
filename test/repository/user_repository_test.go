package repository

import (
	"Ledger/src/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (*entities.User, error) {
	args := m.Called(id)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetAll() ([]entities.User, error) {
	args := m.Called()
	return args.Get(0).([]entities.User), args.Error(1)
}

func TestUserRepository_Create(t *testing.T) {
	mockRepo := new(MockUserRepository)

	user := &entities.User{Name: "Jack", Surname: "Black", Age: 40}
	mockRepo.On("Create", user).Return(nil)

	err := mockRepo.Create(user)

	assert.Nil(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_GetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)

	user := &entities.User{ID: 1, Name: "Jack", Surname: "Black", Age: 40}
	mockRepo.On("GetByID", uint(1)).Return(user, nil)

	result, err := mockRepo.GetByID(1)

	assert.Nil(t, err)
	assert.Equal(t, user, result)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_GetAll(t *testing.T) {
	mockRepo := new(MockUserRepository)

	users := []entities.User{
		{ID: 1, Name: "Jack", Surname: "Black", Age: 50},
		{ID: 2, Name: "Bob", Surname: "Black", Age: 20},
	}
	mockRepo.On("GetAll").Return(users, nil)

	result, err := mockRepo.GetAll()

	assert.Nil(t, err)
	assert.Equal(t, users, result)
	mockRepo.AssertExpectations(t)
}
