package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func getStringRedis(key string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.TimeOut)
	defer cancel()

	val, err := redisClient.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("ошибка при работе с Redis: %v", err)
	}

	return val, nil
}
