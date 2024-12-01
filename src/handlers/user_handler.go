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

func (h *UserHandler) GetTransactionLogsBySenderAndDate(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	senderIdStr := vars.Get("senderId")
	date := vars.Get("date")

	senderId, err := strconv.Atoi(senderIdStr)
	if err != nil || senderId <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid or missing senderId")
		return
	}

	logs, err := h.service.GetTransactionLogsBySenderAndDate(uint(senderId), date)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to fetch transaction logs: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input validation.UserInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid input format")
		return
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

func (h *UserHandler) AddCredit(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	id, err := strconv.Atoi(vars.Get("id"))
	if err != nil || id <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid or missing user ID")
		return
	}

	amountStr := vars.Get("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid or missing amount. Amount must be a positive number.")
		return
	}

	err = h.service.AddCredit(uint(id), amount)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to add credit: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Credit added successfully"}); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) GetCredit(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	id, _ := strconv.Atoi(vars.Get("id"))
	credit, err := h.service.GetCredit(uint(id))
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get credit: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]float64{"credit": credit}); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) GetAllCredits(w http.ResponseWriter, r *http.Request) {
	credits, err := h.service.GetAllCredits()
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve user credits: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(credits); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) SendCredit(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	senderId, err := strconv.Atoi(vars.Get("senderId"))
	if err != nil || senderId <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid or missing sender ID")
		return
	}

	receiverId, err := strconv.Atoi(vars.Get("receiverId"))
	if err != nil || receiverId <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid or missing receiver ID")
		return
	}

	amountStr := vars.Get("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid or missing amount. Amount must be a positive number.")
		return
	}

	err = h.service.SendCredit(uint(senderId), uint(receiverId), amount)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to send credit: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Credit sent successfully"}); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}
