package services

import (
	"Ledger/src/entities"
	"Ledger/src/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockRepository) GetByID(id uint) (*entities.User, error) {
	args := m.Called(id)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockRepository) GetAll() ([]entities.User, error) {
	args := m.Called()
	return args.Get(0).([]entities.User), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
	mockRepo := new(MockRepository)
	service := services.NewUserService(mockRepo)

	user := &entities.User{Name: "Jack", Surname: "Black", Age: 30}
	mockRepo.On("Create", user).Return(nil)

	createdUser, err := service.CreateUser("Jack", "Black", 30)

	assert.Nil(t, err)
	assert.Equal(t, user.Name, createdUser.Name)
	assert.Equal(t, user.Surname, createdUser.Surname)
	assert.Equal(t, user.Age, createdUser.Age)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID(t *testing.T) {
	mockRepo := new(MockRepository)
	service := services.NewUserService(mockRepo)

	user := &entities.User{ID: 1, Name: "Jack", Surname: "Black", Age: 30}
	mockRepo.On("GetByID", uint(1)).Return(user, nil)

	fetchedUser, err := service.GetUserByID(1)

	assert.Nil(t, err)
	assert.Equal(t, user, fetchedUser)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetAllUsers(t *testing.T) {
	mockRepo := new(MockRepository)
	service := services.NewUserService(mockRepo)

	users := []entities.User{
		{ID: 1, Name: "Jack", Surname: "Black", Age: 30},
		{ID: 2, Name: "Bob", Surname: "Black", Age: 20},
	}
	mockRepo.On("GetAll").Return(users, nil)

	fetchedUsers, err := service.GetAllUsers()

	assert.Nil(t, err)
	assert.Equal(t, users, fetchedUsers)
	mockRepo.AssertExpectations(t)
}
