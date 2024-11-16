package factory

import (
	"Ledger/src/handlers"
	"Ledger/src/repository"
	"Ledger/src/services"
	"gorm.io/gorm"
)

type Factory interface {
	NewUserHandler() *handlers.UserHandler
	NewUserService() services.UserService
	NewUserRepository() repository.UserRepository
}

type factory struct {
	db *gorm.DB
}

func NewFactory(db *gorm.DB) Factory {
	return &factory{db: db}
}

func (f *factory) NewUserHandler() *handlers.UserHandler {
	return handlers.NewUserHandler(f.NewUserService())
}

func (f *factory) NewUserService() services.UserService {
	return services.NewUserService(f.NewUserRepository())
}

func (f *factory) NewUserRepository() repository.UserRepository {
	return repository.NewUserRepository(f.db)
}
