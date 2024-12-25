package repository

import (
	"Ledger/pkg/cache"
	"Ledger/src/models"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type userRepository struct {
	db *gorm.DB
	cache *cache.RedisCache
}

func NewUserRepository(db *gorm.DB, cache *cache.RedisCache) UserRepository {
	return &userRepository{
		db: db,
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

func (r *userRepository) GetMultipleUserCredits(userIDs []uint) (map[uint]float64, error) {
	ctx := context.Background()
	
	cachedCredits, err := r.cache.GetMultipleUserCredits(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	missingUserIDs := make([]uint, 0)
	for _, userID := range userIDs {
		if _, exists := cachedCredits[userID]; !exists {
			missingUserIDs = append(missingUserIDs, userID)
		}
	}

	if len(missingUserIDs) > 0 {
		var users []models.User
		err := r.db.Where("id IN ?", missingUserIDs).Select("id, credit").Find(&users).Error
		if err != nil {
			return nil, err
		}

		dbCredits := make(map[uint]float64)
		for _, user := range users {
			dbCredits[user.ID] = user.Credit
			cachedCredits[user.ID] = user.Credit
		}

		go func() {
			_ = r.cache.SetMultipleUserCredits(ctx, dbCredits)
		}()
	}

	return cachedCredits, nil
}

func (r *userRepository) ProcessBatchCreditUpdate(transactions []models.BatchCreditTransaction) []models.BatchTransactionResult {
	results := make([]models.BatchTransactionResult, len(transactions))
	userIDs := make([]uint, len(transactions))
	
	for i, txn := range transactions {
		userIDs[i] = txn.UserID
	}

	var existingUsers []models.User
	err := r.db.Where("id IN ?", userIDs).Select("id, credit").Find(&existingUsers).Error
	if err != nil {
		for i := range results {
			results[i] = models.BatchTransactionResult{
				Success: false,
				UserID:  transactions[i].UserID,
				Amount:  transactions[i].Amount,
				Error:   "Database error",
			}
		}
		return results
	}

	existingUserMap := make(map[uint]bool)
	for _, user := range existingUsers {
		existingUserMap[user.ID] = true
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		ctx := context.Background()
		updatedUserIDs := make([]uint, 0)

		for i, txn := range transactions {
			if !existingUserMap[txn.UserID] {
				results[i] = models.BatchTransactionResult{
					Success: false,
					UserID:  txn.UserID,
					Amount:  txn.Amount,
					Error:   "User not found",
				}
				continue
			}

			err := tx.Model(&models.User{}).
				Where("id = ?", txn.UserID).
				Update("credit", gorm.Expr("credit + ?", txn.Amount)).Error

			if err != nil {
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
			go func() {
				_ = r.cache.InvalidateMultipleUserCredits(ctx, updatedUserIDs)
			}()
		}

		return nil
	})

	return results
} 