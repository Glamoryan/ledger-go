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

func (c *RedisCache) GetMultipleUserCredits(ctx context.Context, userIDs []uint) (map[uint]float64, error) {
    pipe := c.client.Pipeline()
    cmds := make(map[uint]*redis.StringCmd)

    for _, userID := range userIDs {
        key := fmt.Sprintf("user_credit:%d", userID)
        cmds[userID] = pipe.Get(ctx, key)
    }

    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, err
    }

    results := make(map[uint]float64)
    for userID, cmd := range cmds {
        val, err := cmd.Result()
        if err == nil {
            if credit, err := strconv.ParseFloat(val, 64); err == nil {
                results[userID] = credit
            }
        }
    }

    return results, nil
}

func (c *RedisCache) SetMultipleUserCredits(ctx context.Context, credits map[uint]float64) error {
    pipe := c.client.Pipeline()

    for userID, credit := range credits {
        key := fmt.Sprintf("user_credit:%d", userID)
        pipe.Set(ctx, key, credit, 30*time.Minute)
    }

    _, err := pipe.Exec(ctx)
    return err
}

func (c *RedisCache) InvalidateMultipleUserCredits(ctx context.Context, userIDs []uint) error {
    pipe := c.client.Pipeline()

    for _, userID := range userIDs {
        key := fmt.Sprintf("user_credit:%d", userID)
        pipe.Del(ctx, key)
    }

    _, err := pipe.Exec(ctx)
    return err
} 