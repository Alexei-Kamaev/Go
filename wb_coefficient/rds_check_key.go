package main

import (
	"context"
	"fmt"
)

func checkExistsKeyInRedis(key string) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.TimeOut)
	defer cancel()

	count, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("возникла ошибка при проверке ключа [%s] в Редис: %v", key, err)
	}

	return count > 0, nil
}
