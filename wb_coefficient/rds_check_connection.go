package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func checkRedisConnection() (*redis.Client, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.TimeOut)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %v", err)
	}

	return client, nil
}
