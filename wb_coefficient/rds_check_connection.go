package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Функция для проверки соединения с Redis, возвращает Redis-client
func checkRedisConnection() (*redis.Client, error) {

	// создаём клиент с необходимыми параметрами
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.TimeOut)
	defer cancel()

	// пинг Редиса
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %v", err)
	}

	return client, nil
}
