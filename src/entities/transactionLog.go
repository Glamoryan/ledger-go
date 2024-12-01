package entities

import "time"

type TransactionLog struct {
	ID                   uint      `gorm:"primaryKey"`
	SenderID             uint      `gorm:"not null"`
	ReceiverID           uint      `gorm:"not null"`
	Amount               float64   `gorm:"not null"`
	SenderCreditBefore   float64   `gorm:"not null"`
	ReceiverCreditBefore float64   `gorm:"not null"`
	TransactionDate      time.Time `gorm:"autoCreateTime"`
}
