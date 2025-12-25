package main

import (
	"context"
	"fmt"
)

// Функция, которая проверяет время жизни ключа.
func checkTTLRedisKey(key string) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.TimeOut)
	defer cancel()

	ttlRedis, err := redisClient.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("возникла ошибка при проверке ttl ключа в Редис: %v", err)
	}

	return int(ttlRedis.Seconds()), nil
}
