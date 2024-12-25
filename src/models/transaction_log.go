package models

import "time"

type TransactionLog struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	SenderID             uint      `json:"sender_id"`
	ReceiverID           uint      `json:"receiver_id"`
	Amount               float64   `json:"amount"`
	SenderCreditBefore   float64   `json:"sender_credit_before"`
	ReceiverCreditBefore float64   `json:"receiver_credit_before"`
	TransactionDate      time.Time `json:"transaction_date"`
} 