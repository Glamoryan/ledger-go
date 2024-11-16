package handlers

import (
	"Ledger/src/services"
	"Ledger/src/validation"
	"encoding/json"
	"net/http"
	"strconv"
)

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input validation.UserInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid input format")
	}

	if err := validation.ValidateUserInput(input); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.CreateUser(input.Name, input.Surname, input.Age)

	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Could not create a user: "+err.Error())

		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers()

	if err != nil {
		http.Error(w, "Could not retrieve users", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query().Get("id")

	userID, err := strconv.Atoi(vars)

	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	user, err := h.service.GetUserByID(uint(userID))

	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}
