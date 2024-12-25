package factory

import (
	"Ledger/src/handlers"
	"Ledger/src/repository"
	"Ledger/src/services"
	"gorm.io/gorm"
	"Ledger/pkg/auth"
	"Ledger/pkg/middleware"
	"Ledger/pkg/cache"
)

type Factory interface {
	NewUserHandler() *handlers.UserHandler
	NewUserService() services.UserService
	NewUserRepository() repository.UserRepository
	NewAuthMiddleware() middleware.AuthMiddleware
	NewRedisCache() *cache.RedisCache
}

type factory struct {
	db *gorm.DB
	jwtService auth.JWTService
	authMiddleware middleware.AuthMiddleware
	redisCache *cache.RedisCache
}

func NewFactory(db *gorm.DB) Factory {
	jwtService := auth.NewJWTService()
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	redisCache := cache.NewRedisCache("localhost:6379", "", 0)
	
	return &factory{
		db: db,
		jwtService: jwtService,
		authMiddleware: authMiddleware,
		redisCache: redisCache,
	}
}

func (f *factory) NewUserHandler() *handlers.UserHandler {
	return handlers.NewUserHandler(f.NewUserService(), f.jwtService)
}

func (f *factory) NewUserService() services.UserService {
	return services.NewUserService(f.NewUserRepository())
}

func (f *factory) NewUserRepository() repository.UserRepository {
	return repository.NewUserRepository(f.db, f.redisCache)
}

func (f *factory) NewAuthMiddleware() middleware.AuthMiddleware {
	return f.authMiddleware
}

func (f *factory) NewRedisCache() *cache.RedisCache {
	return f.redisCache
}
