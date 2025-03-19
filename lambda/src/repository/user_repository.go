package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"ledger-lambda/src/models"
	"log"
	"time"
)

type UserRepository interface {
	Create(user *models.User) error
	GetAll() ([]models.User, error)
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetUserCredit(userID uint) (float64, error)
	SendCredit(senderID, receiverID uint, amount float64) error
	GetTransactionLogsBySenderAndDate(senderID uint, date string) ([]models.TransactionLog, error)
	AddCredit(userID uint, amount float64) error
	GetAllCredits() ([]models.User, error)
	GetMultipleUserCredits(userIDs []uint) ([]models.User, error)
	ProcessBatchCreditUpdate(transactions []models.BatchTransaction) []models.BatchTransactionResult
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(user *models.User) error {
	if r.db == nil {
		return errors.New("veritabanı bağlantısı yok")
	}

	log.Printf("Yeni kullanıcı oluşturuluyor, şifre düz metin olarak saklanacak")

	query := `
		INSERT INTO users (name, surname, age, email, password, credit)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var credit float64 = 0

	err := r.db.QueryRow(
		query,
		user.Name,
		user.Surname,
		user.Age,
		user.Email,
		user.Password, // Düz metin şifre
		credit,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("kullanıcı oluşturma hatası: %v", err)
	}

	user.Role = "user" // Varsayılan rol veritabanında yok, kod içinde ayarlıyoruz
	user.Credit = credit

	return nil
}

func (r *PostgresUserRepository) GetAll() ([]models.User, error) {
	if r.db == nil {
		return nil, errors.New("veritabanı bağlantısı yok")
	}

	query := `
		SELECT id, name, surname, age, email, credit
		FROM users
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("kullanıcıları getirme hatası: %v", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Surname,
			&user.Age,
			&user.Email,
			&user.Credit,
		); err != nil {
			return nil, fmt.Errorf("kullanıcı verisi okuma hatası: %v", err)
		}

		user.Role = "user"
		if user.Email == "admin@ledger.com" {
			user.Role = "admin"
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("veri akışı hatası: %v", err)
	}

	return users, nil
}

func (r *PostgresUserRepository) GetByID(id uint) (*models.User, error) {
	if r.db == nil {
		return nil, errors.New("veritabanı bağlantısı yok")
	}

	query := `
		SELECT id, name, surname, age, email, credit
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Age,
		&user.Email,
		&user.Credit,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("kullanıcı bulunamadı: %d", id)
	} else if err != nil {
		return nil, fmt.Errorf("kullanıcı getirme hatası: %v", err)
	}

	user.Role = "user"
	if user.Email == "admin@ledger.com" {
		user.Role = "admin"
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetByEmail(email string) (*models.User, error) {
	if r.db == nil {
		return nil, errors.New("veritabanı bağlantısı yok")
	}

	query := `
		SELECT id, name, surname, age, email, password, credit
		FROM users
		WHERE email = $1
	`

	log.Printf("GetByEmail çağrılıyor, email: %s", email)

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Age,
		&user.Email,
		&user.PasswordHash,
		&user.Credit,
	)

	if err == sql.ErrNoRows {
		log.Printf("Kullanıcı bulunamadı: %s", email)
		return nil, fmt.Errorf("kullanıcı bulunamadı: %s", email)
	} else if err != nil {
		log.Printf("Kullanıcı getirme hatası: %v", err)
		return nil, fmt.Errorf("kullanıcı getirme hatası: %v", err)
	}

	user.Role = "user"
	if user.Email == "admin@ledger.com" {
		user.Role = "admin"
	}

	log.Printf("Kullanıcı başarıyla bulundu: %s (ID: %d)", email, user.ID)
	return &user, nil
}

func (r *PostgresUserRepository) GetUserCredit(userID uint) (float64, error) {
	if r.db == nil {
		return 0, errors.New("veritabanı bağlantısı yok")
	}

	query := `SELECT credit FROM users WHERE id = $1`

	var credit float64
	err := r.db.QueryRow(query, userID).Scan(&credit)

	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("kullanıcı bulunamadı: %d", userID)
	} else if err != nil {
		return 0, fmt.Errorf("kredi getirme hatası: %v", err)
	}

	return credit, nil
}

func (r *PostgresUserRepository) SendCredit(senderID, receiverID uint, amount float64) error {
	if r.db == nil {
		return errors.New("veritabanı bağlantısı yok")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("transaction başlatma hatası: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic
		} else if err != nil {
			tx.Rollback() // hata varsa rollback
		}
	}()

	var senderCredit, receiverCredit float64
	err = tx.QueryRow("SELECT credit FROM users WHERE id = $1 FOR UPDATE", senderID).Scan(&senderCredit)
	if err != nil {
		return fmt.Errorf("gönderen kredi bilgisi alınamadı: %v", err)
	}

	err = tx.QueryRow("SELECT credit FROM users WHERE id = $1 FOR UPDATE", receiverID).Scan(&receiverCredit)
	if err != nil {
		return fmt.Errorf("alıcı bulunamadı: %v", err)
	}

	if senderCredit < amount {
		return errors.New("yetersiz kredi")
	}

	senderCreditAfter := senderCredit - amount
	receiverCreditAfter := receiverCredit + amount

	_, err = tx.Exec("UPDATE users SET credit = $1 WHERE id = $2", senderCreditAfter, senderID)
	if err != nil {
		return fmt.Errorf("gönderen kredi güncellemesi başarısız: %v", err)
	}

	_, err = tx.Exec("UPDATE users SET credit = $1 WHERE id = $2", receiverCreditAfter, receiverID)
	if err != nil {
		return fmt.Errorf("alıcı kredi güncellemesi başarısız: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO transaction_logs 
		(sender_id, receiver_id, amount, description, sender_credit_before, receiver_credit_before, 
		sender_credit_after, receiver_credit_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		senderID, receiverID, amount, "Kredi transferi", senderCredit, receiverCredit,
		senderCreditAfter, receiverCreditAfter, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("işlem kaydı oluşturma hatası: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit hatası: %v", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetTransactionLogsBySenderAndDate(senderID uint, dateStr string) ([]models.TransactionLog, error) {
	if r.db == nil {
		return nil, errors.New("veritabanı bağlantısı yok")
	}

	query := `
		SELECT id, sender_id, receiver_id, amount, description, sender_credit_before, receiver_credit_before, 
		sender_credit_after, receiver_credit_after, created_at
		FROM transaction_logs
		WHERE sender_id = $1 AND DATE(created_at) = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, senderID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("işlem kayıtları getirme hatası: %v", err)
	}
	defer rows.Close()

	var logs []models.TransactionLog
	for rows.Next() {
		var log models.TransactionLog
		if err := rows.Scan(
			&log.ID,
			&log.SenderID,
			&log.ReceiverID,
			&log.Amount,
			&log.Description,
			&log.SenderCreditBefore,
			&log.ReceiverCreditBefore,
			&log.SenderCreditAfter,
			&log.ReceiverCreditAfter,
			&log.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("işlem kaydı okuma hatası: %v", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("veri akışı hatası: %v", err)
	}

	return logs, nil
}

func (r *PostgresUserRepository) AddCredit(userID uint, amount float64) error {
	if r.db == nil {
		return errors.New("veritabanı bağlantısı yok")
	}

	if amount <= 0 {
		return errors.New("eklenen miktar pozitif olmalıdır")
	}

	query := `UPDATE users SET credit = credit + $1 WHERE id = $2`

	result, err := r.db.Exec(query, amount, userID)
	if err != nil {
		return fmt.Errorf("kredi ekleme hatası: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("sonuç kontrol hatası: %v", err)
	}

	if rows == 0 {
		return fmt.Errorf("kullanıcı bulunamadı: %d", userID)
	}

	return nil
}

func (r *PostgresUserRepository) GetAllCredits() ([]models.User, error) {
	if r.db == nil {
		return nil, errors.New("veritabanı bağlantısı yok")
	}

	query := `
		SELECT id, name, surname, email, credit
		FROM users
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("kredi bilgilerini getirme hatası: %v", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Surname,
			&user.Email,
			&user.Credit,
		); err != nil {
			return nil, fmt.Errorf("kullanıcı verisi okuma hatası: %v", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("veri akışı hatası: %v", err)
	}

	return users, nil
}

func (r *PostgresUserRepository) GetMultipleUserCredits(userIDs []uint) ([]models.User, error) {
	if r.db == nil {
		return nil, errors.New("veritabanı bağlantısı yok")
	}

	if len(userIDs) == 0 {
		return []models.User{}, nil
	}

	placeholders := ""
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, name, surname, email, credit
		FROM users
		WHERE id IN (%s)
		ORDER BY id
	`, placeholders)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("kredi bilgilerini getirme hatası: %v", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Surname,
			&user.Email,
			&user.Credit,
		); err != nil {
			return nil, fmt.Errorf("kullanıcı verisi okuma hatası: %v", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("veri akışı hatası: %v", err)
	}

	return users, nil
}

func (r *PostgresUserRepository) ProcessBatchCreditUpdate(transactions []models.BatchTransaction) []models.BatchTransactionResult {
	if r.db == nil {
		return []models.BatchTransactionResult{{
			Success: false,
			Error:   "veritabanı bağlantısı yok",
		}}
	}

	results := make([]models.BatchTransactionResult, 0, len(transactions))

	for _, tx := range transactions {
		result := models.BatchTransactionResult{
			UserID: tx.UserID,
			Amount: tx.Amount,
		}

		var err error
		if tx.Amount >= 0 {
			err = r.AddCredit(tx.UserID, tx.Amount)
		} else {
			err = r.AddCredit(tx.UserID, tx.Amount) // Amount zaten negatif
		}

		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results
}
