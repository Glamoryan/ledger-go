package factory

import (
	"Ledger/src/handlers"
	"Ledger/src/repository"
	"Ledger/src/services"
	"gorm.io/gorm"
	"Ledger/pkg/auth"
	"Ledger/pkg/middleware"
)

type Factory interface {
	NewUserHandler() *handlers.UserHandler
	NewUserService() services.UserService
	NewUserRepository() repository.UserRepository
	NewAuthMiddleware() middleware.AuthMiddleware
}

type factory struct {
	db *gorm.DB
	jwtService auth.JWTService
	authMiddleware middleware.AuthMiddleware
}

func NewFactory(db *gorm.DB) Factory {
	jwtService := auth.NewJWTService()
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	
	return &factory{
		db: db,
		jwtService: jwtService,
		authMiddleware: authMiddleware,
	}
}

func (f *factory) NewUserHandler() *handlers.UserHandler {
	return handlers.NewUserHandler(f.NewUserService(), f.jwtService)
}

func (f *factory) NewUserService() services.UserService {
	return services.NewUserService(f.NewUserRepository())
}

func (f *factory) NewUserRepository() repository.UserRepository {
	return repository.NewUserRepository(f.db)
}

func (f *factory) NewAuthMiddleware() middleware.AuthMiddleware {
	return f.authMiddleware
}
