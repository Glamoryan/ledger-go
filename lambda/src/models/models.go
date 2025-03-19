package models

import "time"

type User struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Surname      string  `json:"surname"`
	Age          int     `json:"age"`
	Email        string  `json:"email"`
	Password     string  `json:"password,omitempty"`
	PasswordHash string  `json:"-"`
	Role         string  `json:"role"`
	Credit       float64 `json:"credit"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Age      int    `json:"age"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TransactionLog struct {
	ID                   uint      `json:"id"`
	UserID               uint      `json:"user_id,omitempty"`
	SenderID             uint      `json:"sender_id"`
	ReceiverID           uint      `json:"receiver_id"`
	Amount               float64   `json:"amount"`
	Description          string    `json:"description,omitempty"`
	SenderCreditBefore   float64   `json:"sender_credit_before"`
	ReceiverCreditBefore float64   `json:"receiver_credit_before"`
	SenderCreditAfter    float64   `json:"sender_credit_after"`
	ReceiverCreditAfter  float64   `json:"receiver_credit_after"`
	CreatedAt            time.Time `json:"created_at"`
}

type BatchTransaction struct {
	UserID uint    `json:"user_id"`
	Amount float64 `json:"amount"`
}

type BatchTransactionResult struct {
	Success bool    `json:"success"`
	UserID  uint    `json:"user_id"`
	Amount  float64 `json:"amount"`
	Error   string  `json:"error"`
}
