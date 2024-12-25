package repository

import (
	"Ledger/src/models"
	"gorm.io/gorm"
	"time"
	"errors"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) GetUserCredit(userID uint) (float64, error) {
	var credit float64
	err := r.db.Model(&models.User{}).Where("id = ?", userID).Select("credit").Scan(&credit).Error
	return credit, err
}

func (r *userRepository) UpdateCredit(userID uint, newAmount float64) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("credit", newAmount).Error
}

func (r *userRepository) GetAllCredits() (map[uint]float64, error) {
	var users []models.User
	err := r.db.Select("id, credit").Find(&users).Error
	if err != nil {
		return nil, err
	}

	credits := make(map[uint]float64)
	for _, user := range users {
		credits[user.ID] = user.Credit
	}
	return credits, nil
}

func (r *userRepository) SendCreditToUser(senderID, receiverID uint, amount float64) error {
	var receiver models.User
	if err := r.db.First(&receiver, receiverID).Error; err != nil {
		return errors.New("receiver not found")
	}

	var sender models.User
	if err := r.db.First(&sender, senderID).Error; err != nil {
		return errors.New("sender not found")
	}

	if sender.Credit < amount {
		return errors.New("insufficient balance")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).
			Where("id = ?", senderID).
			Update("credit", gorm.Expr("credit - ?", amount)).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.User{}).
			Where("id = ?", receiverID).
			Update("credit", gorm.Expr("credit + ?", amount)).Error; err != nil {
			return err
		}


		logEntry := models.TransactionLog{
			SenderID:             senderID,
			ReceiverID:           receiverID,
			Amount:               amount,
			SenderCreditBefore:   sender.Credit,
			ReceiverCreditBefore: receiver.Credit,
			TransactionDate:      time.Now(),
		}
		if err := tx.Create(&logEntry).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *userRepository) LogTransaction(senderID, receiverID uint, amount, senderCreditBefore, receiverCreditBefore float64) error {
	return nil
}

func (r *userRepository) GetTransactionLogsBySenderAndDate(senderID uint, date string) ([]models.TransactionLog, error) {
	var logs []models.TransactionLog
	query := r.db.Model(&models.TransactionLog{})

	err := query.
		Where("sender_id = ?", senderID).
		Where("DATE(transaction_date) = ?", date).
		Order("transaction_date DESC").
		Find(&logs).Error

	return logs, err
} 