package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"dosage_drugs/internal/app/config"
)

// Префикс сервиса нужен, чтобы ключи разных сервисов в одном Redis не пересекались.
const blacklistPrefix = "dosage_drugs.jwt.blacklist."

type Client struct {
	rdb *goredis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password:    cfg.Password,
		DB:          0,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

// BlacklistToken кладёт токен в чёрный список на время, пока он ещё действителен.
// После истечения TTL запись удаляется Redis'ом автоматически.
func (c *Client) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil // токен уже истёк, хранить нечего
	}
	return c.rdb.Set(ctx, blacklistPrefix+token, 1, ttl).Err()
}

// IsBlacklisted сообщает, был ли токен отозван через logout.
func (c *Client) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	n, err := c.rdb.Exists(ctx, blacklistPrefix+token).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
