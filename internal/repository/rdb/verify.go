package rdb

import (
	"context"
	"time"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/sirupsen/logrus"

	"github.com/go-redis/redis"
)

type Verify struct {
	redis *redis.Client

	ttl time.Duration

	log *logrus.Logger
}

func NewVerife(redis *redis.Client, ttl time.Duration, log *logrus.Logger) *Verify {
	return &Verify{
		redis: redis,

		ttl: ttl,

		log: log,
	}
}

func (v *Verify) SetCode(ctx context.Context, id int, code string) error {
	if err := v.redis.Set(code, id, v.ttl).Err(); err != nil {
		return err
	}

	return nil
}

func (v *Verify) Verify(ctx context.Context, code string) (string, error) {
	id, err := v.redis.Get(code).Result()
	if err != nil {
		if err == redis.Nil {
			return "", core.ErrCodeNotFound
		}

		return "", err
	}

	if err := v.redis.Del(code).Err(); err != nil {
		return "", err
	}

	return id, nil
}
