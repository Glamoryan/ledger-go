package repository

import (
	"Ledger/pkg/cache"
	"Ledger/src/models"
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type userRepository struct {
	db    *gorm.DB
	cache *cache.RedisCache
}

func NewUserRepository(db *gorm.DB, cache *cache.RedisCache) UserRepository {
	return &userRepository{
		db:    db,
		cache: cache,
	}
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
	ctx := context.Background()

	credit, err := r.cache.GetUserCredit(ctx, userID)
	if err == nil {
		fmt.Printf("Cache HIT for user %d: %f\n", userID, credit)
		return credit, nil
	}
	fmt.Printf("Cache MISS for user %d\n", userID)

	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errors.New("user not found")
		}
		return 0, err
	}

	var dbCredit float64
	err = r.db.Model(&models.User{}).Where("id = ?", userID).Select("credit").Scan(&dbCredit).Error
	if err != nil {
		return 0, err
	}

	err = r.cache.SetUserCredit(ctx, userID, dbCredit)
	if err != nil {
		fmt.Printf("Failed to set cache for user %d: %v\n", userID, err)
	} else {
		fmt.Printf("Cache SET for user %d: %f\n", userID, dbCredit)
	}

	return dbCredit, nil
}

func (r *userRepository) UpdateCredit(userID uint, newAmount float64) error {
	ctx := context.Background()

	err := r.db.Model(&models.User{}).Where("id = ?", userID).Update("credit", newAmount).Error
	if err != nil {
		return err
	}

	_ = r.cache.InvalidateUserCredit(ctx, userID)

	return nil
}

func (r *userRepository) GetAllCredits() ([]models.User, error) {
	var users []models.User
	err := r.db.Select("id, credit").Find(&users).Error
	return users, err
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

func (r *userRepository) GetMultipleUserCredits(userIDs []uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Select("id, credit").Where("id IN ?", userIDs).Find(&users).Error
	return users, err
}

func (r *userRepository) ProcessBatchCreditUpdate(transactions []models.BatchTransaction) []models.BatchTransactionResult {
	results := make([]models.BatchTransactionResult, len(transactions))
	updatedUserIDs := make([]uint, 0)

	err := r.db.Transaction(func(tx *gorm.DB) error {
		for i, txn := range transactions {
			var user models.User
			if err := tx.First(&user, txn.UserID).Error; err != nil {
				results[i] = models.BatchTransactionResult{
					Success: false,
					UserID:  txn.UserID,
					Amount:  txn.Amount,
					Error:   "User not found",
				}
				continue
			}

			if err := tx.Model(&user).Update("credit", user.Credit+txn.Amount).Error; err != nil {
				results[i] = models.BatchTransactionResult{
					Success: false,
					UserID:  txn.UserID,
					Amount:  txn.Amount,
					Error:   "Update failed",
				}
			} else {
				results[i] = models.BatchTransactionResult{
					Success: true,
					UserID:  txn.UserID,
					Amount:  txn.Amount,
				}
				updatedUserIDs = append(updatedUserIDs, txn.UserID)
			}
		}

		if len(updatedUserIDs) > 0 {
			ctx := context.Background()
			go func() {
				_ = r.cache.InvalidateMultipleUserCredits(ctx, updatedUserIDs)
			}()
		}

		return nil
	})

	if err != nil {
		return results
	}

	return results
}

func (r *userRepository) SendCredit(senderID, receiverID uint, amount float64) error {
	var sender, receiver models.User
	if err := r.db.First(&sender, senderID).Error; err != nil {
		return errors.New("sender not found")
	}
	if err := r.db.First(&receiver, receiverID).Error; err != nil {
		return errors.New("receiver not found")
	}

	if sender.Credit < amount {
		return errors.New("insufficient balance")
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		senderCreditBefore := sender.Credit
		receiverCreditBefore := receiver.Credit

		// Update sender's credit
		if err := tx.Model(&sender).Update("credit", sender.Credit-amount).Error; err != nil {
			return err
		}

		// Update receiver's credit
		if err := tx.Model(&receiver).Update("credit", receiver.Credit+amount).Error; err != nil {
			return err
		}

		// Log the transaction
		transactionLog := models.TransactionLog{
			SenderID:             senderID,
			ReceiverID:           receiverID,
			Amount:               amount,
			SenderCreditBefore:   senderCreditBefore,
			ReceiverCreditBefore: receiverCreditBefore,
			TransactionDate:      time.Now(),
		}
		if err := tx.Create(&transactionLog).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Invalidate cache for both users
	ctx := context.Background()
	go func() {
		_ = r.cache.InvalidateUserCredit(ctx, senderID)
		_ = r.cache.InvalidateUserCredit(ctx, receiverID)
	}()

	return nil
}

func (r *userRepository) AddCredit(userID uint, amount float64) error {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	if err := r.db.Model(&user).Update("credit", user.Credit+amount).Error; err != nil {
		return err
	}

	ctx := context.Background()
	go func() {
		_ = r.cache.InvalidateUserCredit(ctx, userID)
	}()

	return nil
}
