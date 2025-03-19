package factory

import (
	"database/sql"
	"ledger-lambda/pkg/auth"
	"ledger-lambda/src/handlers"
	"ledger-lambda/src/repository"
	"ledger-lambda/src/services"
)

type Factory interface {
	NewUserHandler() *handlers.UserHandler
	NewUserService() services.UserService
	NewUserRepository() repository.UserRepository
	NewJWTService() auth.JWTService
}

type factory struct {
	db         *sql.DB
	jwtService auth.JWTService
}

func NewFactory(db *sql.DB) Factory {
	jwtService := auth.NewJWTService()
	return &factory{
		db:         db,
		jwtService: jwtService,
	}
}

func (f *factory) NewUserHandler() *handlers.UserHandler {
	return handlers.NewUserHandler(f.NewUserService(), f.jwtService)
}

func (f *factory) NewUserService() services.UserService {
	return services.NewUserService(f.NewUserRepository())
}

func (f *factory) NewUserRepository() repository.UserRepository {
	return repository.NewPostgresUserRepository(f.db)
}

func (f *factory) NewJWTService() auth.JWTService {
	return f.jwtService
}
