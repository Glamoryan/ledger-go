package repository

import (
	"Ledger/src/entities"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *entities.User) error
	GetByID(id uint) (*entities.User, error)
	GetAll() ([]entities.User, error)
	UpdateCredit(id uint, credit float64) error
	GetUserCredit(id uint) (float64, error)
	GetAllCredits() ([]map[string]interface{}, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *entities.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*entities.User, error) {
	var user entities.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *userRepository) GetAll() ([]entities.User, error) {
	var users []entities.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateCredit(id uint, credit float64) error {
	return r.db.Model(&entities.User{}).Where("id = ?", id).Update("credit", credit).Error
}

func (r *userRepository) GetUserCredit(id uint) (float64, error) {
	var credit float64
	err := r.db.Model(&entities.User{}).Where("id = ?", id).Select("credit").Scan(&credit).Error
	return credit, err
}

func (r *userRepository) GetAllCredits() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Model(&entities.User{}).
		Select("id, name, credit").
		Find(&results).Error
	return results, err
}
