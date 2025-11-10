package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisConfig struct {
	Addr     string
	Password string
	DB       int
	TimeOut  time.Duration
}

func getStringValueRedis(key string) (value string, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), cfgRedis.TimeOut)
	defer cancel()

	value, err = redisClient.Get(ctx, key).Result()

	switch err {
	case redis.Nil:
		addLog("ключ %s не найден в Redis", key)
		return "", nil
	case nil:
		return value, nil
	default:
		return "", fmt.Errorf("ошибка при работе с Redis: %w", err)
	}

}

func checkRedisConnection() (*redis.Client, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     cfgRedis.Addr,
		Password: cfgRedis.Password,
		DB:       cfgRedis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfgRedis.TimeOut)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %v", err)
	}

	return client, nil

}
