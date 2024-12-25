package models

type BatchCreditTransaction struct {
    UserID uint    `json:"user_id"`
    Amount float64 `json:"amount"`
}

type BatchTransactionRequest struct {
    Transactions []BatchCreditTransaction `json:"transactions"`
}

type BatchTransactionResult struct {
    Success    bool   `json:"success"`
    UserID     uint   `json:"user_id"`
    Amount     float64 `json:"amount"`
    Error      string `json:"error,omitempty"`
} 