package handlers

import (
	"encoding/json"
	"io"
	"ledger-lambda/pkg/auth"
	"ledger-lambda/src/models"
	"ledger-lambda/src/services"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := &models.User{
		Name:     req.Name,
		Surname:  req.Surname,
		Age:      req.Age,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.service.CreateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := h.service.ValidatePassword(user, req.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID, user.Email, user.Role == "admin")
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *UserHandler) GetCredit(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	credit, err := h.service.GetUserCredit(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]float64{
		"credit": credit,
	})
}

func (h *UserHandler) SendCredit(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	bodyString := string(bodyBytes)
	log.Printf("SendCredit istek gövdesi: %s", bodyString)

	r.Body = io.NopCloser(strings.NewReader(bodyString))

	var req struct {
		SenderID    interface{} `json:"sender_id"`
		ReceiverID  interface{} `json:"receiver_id"`
		Amount      float64     `json:"amount"`
		Description string      `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode hatası: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Ayrıştırılan JSON: %+v", req)

	var senderID uint
	switch v := req.SenderID.(type) {
	case float64:
		senderID = uint(v)
	case int:
		senderID = uint(v)
	case int64:
		senderID = uint(v)
	case json.Number:
		id, err := v.Int64()
		if err != nil {
			log.Printf("JSON Number dönüştürme hatası: %v", err)
			http.Error(w, "Invalid sender ID", http.StatusBadRequest)
			return
		}
		senderID = uint(id)
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			log.Printf("String ID dönüştürme hatası: %v", err)
			http.Error(w, "Invalid sender ID", http.StatusBadRequest)
			return
		}
		senderID = uint(id)
	default:
		log.Printf("Desteklenmeyen Sender ID tipi: %T", v)
		http.Error(w, "Invalid sender ID", http.StatusBadRequest)
		return
	}

	var receiverID uint
	switch v := req.ReceiverID.(type) {
	case float64:
		receiverID = uint(v)
	case int:
		receiverID = uint(v)
	case int64:
		receiverID = uint(v)
	case json.Number:
		id, err := v.Int64()
		if err != nil {
			log.Printf("JSON Number dönüştürme hatası: %v", err)
			http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
			return
		}
		receiverID = uint(id)
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			log.Printf("String ID dönüştürme hatası: %v", err)
			http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
			return
		}
		receiverID = uint(id)
	default:
		log.Printf("Desteklenmeyen Receiver ID tipi: %T", v)
		http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
		return
	}

	log.Printf("İşlem başlatılıyor: senderID=%d, receiverID=%d, amount=%.2f", senderID, receiverID, req.Amount)

	if err := h.service.SendCredit(senderID, receiverID, req.Amount); err != nil {
		log.Printf("Kredi gönderme hatası: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Kredi başarıyla gönderildi: senderID=%d, receiverID=%d, amount=%.2f", senderID, receiverID, req.Amount)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Credit transferred successfully",
	})
}

func (h *UserHandler) GetTransactionLogsBySenderAndDate(w http.ResponseWriter, r *http.Request) {
	senderIDStr := r.URL.Query().Get("senderId")
	dateStr := r.URL.Query().Get("date")

	senderID, err := strconv.ParseUint(senderIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid sender ID", http.StatusBadRequest)
		return
	}

	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	logs, err := h.service.GetTransactionLogsBySenderAndDate(uint(senderID), dateStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(logs)
}

func (h *UserHandler) AddCredit(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	amountStr := r.URL.Query().Get("amount")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	if err := h.service.AddCredit(uint(id), amount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Credit added successfully",
	})
}

func (h *UserHandler) GetAllCredits(w http.ResponseWriter, r *http.Request) {
	credits, err := h.service.GetAllCredits()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(credits)
}

func (h *UserHandler) GetMultipleUserCredits(w http.ResponseWriter, r *http.Request) {
	var userIDs []uint
	if err := json.NewDecoder(r.Body).Decode(&userIDs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	credits, err := h.service.GetMultipleUserCredits(userIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(credits)
}

func (h *UserHandler) ProcessBatchCreditUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Transactions []models.BatchTransaction `json:"transactions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results := h.service.ProcessBatchCreditUpdate(req.Transactions)
	json.NewEncoder(w).Encode(results)
}
