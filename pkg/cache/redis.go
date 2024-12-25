package cache

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
    "strconv"
    "time"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr string, password string, db int) *RedisCache {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })

    return &RedisCache{
        client: client,
    }
}

func (c *RedisCache) GetUserCredit(ctx context.Context, userID uint) (float64, error) {
    key := fmt.Sprintf("user_credit:%d", userID)
    val, err := c.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return 0, fmt.Errorf("cache miss")
    } else if err != nil {
        return 0, err
    }

    return strconv.ParseFloat(val, 64)
}

func (c *RedisCache) SetUserCredit(ctx context.Context, userID uint, credit float64) error {
    key := fmt.Sprintf("user_credit:%d", userID)
    return c.client.Set(ctx, key, credit, 30*time.Minute).Err()
}

func (c *RedisCache) InvalidateUserCredit(ctx context.Context, userID uint) error {
    key := fmt.Sprintf("user_credit:%d", userID)
    return c.client.Del(ctx, key).Err()
} 