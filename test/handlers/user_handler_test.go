package handlers

import (
	"Ledger/src/entities"
	"Ledger/src/handlers"
	"Ledger/src/services"
	"bytes"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(name, surname string, age int) (*entities.User, error) {
	args := m.Called(name, surname, age)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uint) (*map[string]interface{}, error) {
	args := m.Called(id)
	return args.Get(0).(*map[string]interface{}), args.Error(1)
}

func (m *MockUserService) GetAllUsers() ([]map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func TestUserHandler_CreateUser(t *testing.T) {
	mockService := new(services.MockUserService)
	handler := handlers.NewUserHandler(mockService)

	userInput := `{"name": "Jack", "surname": "Black", "Age": 30}`
	req, _ := http.NewRequest("POST", "/users/add-user", bytes.NewBuffer([]byte(userInput)))
	res := httptest.NewRecorder()

	mockService.On("CreateUser", "Jack", "Black", 30).Return(nil)

	handler.CreateUser(res, req)

	if status := res.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	mockService.AssertExpectations(t)
}

func TestUserHandler_GetUserByID(t *testing.T) {
	mockService := new(services.MockUserService)
	handler := handlers.NewUserHandler(mockService)

	user := map[string]interface{}{"id": 1, "name": "Jack", "surname": "Black", "aga": 30}
	mockService.On("GetUserByID", uint(1)).Return(&user, nil)

	req, _ := http.NewRequest("GET", "/users/1", nil)
	res := httptest.NewRecorder()

	handler.GetUserByID(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	mockService.AssertExpectations(t)
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	mockService := new(services.MockUserService)
	handler := handlers.NewUserHandler(mockService)

	users := []map[string]interface{}{
		{"id": 1, "name": "Jack", "surname": "Black", "age": 30},
		{"id": 2, "name": "Bob", "surname": "Black", "age": 25},
	}
	mockService.On("GetAllUsers").Return(users, nil)

	req, _ := http.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	handler.GetAllUsers(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	mockService.AssertExpectations(t)
}
