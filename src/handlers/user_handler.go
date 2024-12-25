package handlers

import (
	"Ledger/src/services"
	"encoding/json"
	"net/http"
	"strconv"
	"Ledger/pkg/middleware"
	"Ledger/src/models"
	"Ledger/pkg/auth"
	"Ledger/pkg/response"
	"golang.org/x/crypto/bcrypt"
	"fmt"
)

type UserHandler struct {
	service    services.UserService
	jwtService auth.JWTService
}

func NewUserHandler(service services.UserService, jwtService auth.JWTService) *UserHandler {
	return &UserHandler{
		service:    service,
		jwtService: jwtService,
	}
}

func (h *UserHandler) GetTransactionLogsBySenderAndDate(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	senderIdStr := vars.Get("senderId")
	date := vars.Get("date")

	senderId, err := strconv.Atoi(senderIdStr)
	if err != nil || senderId <= 0 {
		response.WriteError(w, http.StatusBadRequest, "Invalid or missing senderId")
		return
	}

	logs, err := h.service.GetTransactionLogsBySenderAndDate(uint(senderId), date)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Failed to fetch transaction logs: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	user := &models.User{
		Name:         req.Name,
		Surname:      req.Surname,
		Age:         req.Age,
		Email:       req.Email,
		PasswordHash: string(hashedPassword),
		Role:        "user",
	}

	err = h.service.CreateUser(user)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Error creating user: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
	})
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers()

	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Could not retrieve users: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query().Get("id")

	userID, err := strconv.Atoi(vars)

	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	user, err := h.service.GetUserByID(uint(userID))

	if err != nil {
		response.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) AddCredit(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	id, err := strconv.Atoi(vars.Get("id"))
	if err != nil || id <= 0 {
		response.WriteError(w, http.StatusBadRequest, "Invalid or missing user ID")
		return
	}

	amountStr := vars.Get("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		response.WriteError(w, http.StatusBadRequest, "Invalid or missing amount. Amount must be a positive number.")
		return
	}

	err = h.service.AddCredit(uint(id), amount)
	if err != nil {
		if err.Error() == "user not found" {
			response.WriteError(w, http.StatusNotFound, fmt.Sprintf("User with ID %d not found", id))
		} else {
			response.WriteError(w, http.StatusInternalServerError, "Failed to add credit: "+err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Credit added successfully"})
}

func (h *UserHandler) GetCredit(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	if !h.jwtService.IsAdmin(claims) && uint(userID) != claims.UserID {
		response.WriteError(w, http.StatusForbidden, "You can only view your own credit balance")
		return
	}

	credit, err := h.service.GetUserCredit(uint(userID))
	if err != nil {
		if err.Error() == "user not found" {
			response.WriteError(w, http.StatusNotFound, fmt.Sprintf("User with ID %d not found", userID))
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Error getting credit: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{
		"credit": credit,
	})
}

func (h *UserHandler) GetAllCredits(w http.ResponseWriter, r *http.Request) {
	credits, err := h.service.GetAllCredits()
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Failed to retrieve user credits: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(credits); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Failed to encode response: "+err.Error())
	}
}

func (h *UserHandler) SendCredit(w http.ResponseWriter, r *http.Request) {
	senderIDStr := r.URL.Query().Get("senderId")
	receiverIDStr := r.URL.Query().Get("receiverId")
	amountStr := r.URL.Query().Get("amount")

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	senderID, _ := strconv.Atoi(senderIDStr)

	if uint(senderID) != claims.UserID {
		response.WriteError(w, http.StatusForbidden, "You can only send credit from your own account")
		return
	}

	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid receiver ID format")
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid amount format")
		return
	}

	err = h.service.SendCredit(uint(senderID), uint(receiverID), amount)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Error sending credit: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Credit transferred successfully",
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	user, err := h.service.GetUserByEmail(req.Email)
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *UserHandler) GetMultipleUserCredits(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	if !h.jwtService.IsAdmin(claims) {
		response.WriteError(w, http.StatusForbidden, "Admin access required")
		return
	}

	var userIDs []uint
	if err := json.NewDecoder(r.Body).Decode(&userIDs); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	credits, err := h.service.GetMultipleUserCredits(userIDs)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Error getting credits: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(credits)
}

func (h *UserHandler) ProcessBatchCreditUpdate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	if !h.jwtService.IsAdmin(claims) {
		response.WriteError(w, http.StatusForbidden, "Admin access required")
		return
	}

	var req models.BatchTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if len(req.Transactions) == 0 {
		response.WriteError(w, http.StatusBadRequest, "No transactions provided")
		return
	}

	if len(req.Transactions) > 1000 {
		response.WriteError(w, http.StatusBadRequest, "Too many transactions. Maximum 1000 allowed")
		return
	}

	results := h.service.ProcessBatchCreditUpdate(req.Transactions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
