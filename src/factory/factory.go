package factory

import (
	"Ledger/pkg/auth"
	"Ledger/pkg/cache"
	"Ledger/pkg/middleware"
	"Ledger/pkg/queue"
	"Ledger/src/handlers"
	"Ledger/src/repository"
	"Ledger/src/services"
	"gorm.io/gorm"
)

type Factory interface {
	NewUserHandler() *handlers.UserHandler
	NewUserService() services.UserService
	NewUserRepository() repository.UserRepository
	NewAuthMiddleware() middleware.AuthMiddleware
	NewRedisCache() *cache.RedisCache
	NewRabbitMQ() *queue.RabbitMQ
}

type factory struct {
	db *gorm.DB
	jwtService auth.JWTService
	authMiddleware middleware.AuthMiddleware
	redisCache *cache.RedisCache
	rabbitMQ *queue.RabbitMQ
}

func NewFactory(db *gorm.DB) Factory {
	jwtService := auth.NewJWTService()
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	redisCache := cache.NewRedisCache("localhost:6379", "", 0)
	
	rabbitMQ, err := queue.NewRabbitMQ("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	
	return &factory{
		db: db,
		jwtService: jwtService,
		authMiddleware: authMiddleware,
		redisCache: redisCache,
		rabbitMQ: rabbitMQ,
	}
}

func (f *factory) NewUserHandler() *handlers.UserHandler {
	return handlers.NewUserHandler(f.NewUserService(), f.jwtService)
}

func (f *factory) NewUserService() services.UserService {
	return services.NewUserService(f.NewUserRepository())
}

func (f *factory) NewUserRepository() repository.UserRepository {
	return repository.NewUserRepository(f.db, f.redisCache, f.rabbitMQ)
}

func (f *factory) NewAuthMiddleware() middleware.AuthMiddleware {
	return f.authMiddleware
}

func (f *factory) NewRedisCache() *cache.RedisCache {
	return f.redisCache
}

func (f *factory) NewRabbitMQ() *queue.RabbitMQ {
	return f.rabbitMQ
}
