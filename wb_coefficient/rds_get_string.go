package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Функция, которая получает значение по ключу из Редис.
// Возвращает значение и ошибку.
func getStringRedis(key string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.TimeOut)
	defer cancel()

	val, err := redisClient.Get(ctx, key).Result()

	// если нет ключа и ошибки, возвращаем пустую строку
	if err == redis.Nil {
		return "", nil
	}

	// обработка ошибки при работе с Редис
	if err != nil {
		return "", fmt.Errorf("ошибка при работе с Redis: %v", err)
	}

	return val, nil
}
