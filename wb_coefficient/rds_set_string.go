package main

import (
	"context"
	"fmt"
	"time"
)

func setStringRedis(key, data string) error {

	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.TimeOut)
	defer cancel()

	var redisExp = 1

	if appConfig != nil {
		redisExp = appConfig.RedisExpiration
	}

	err := redisClient.Set(ctx, key, data, time.Duration(redisExp)*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("ошибка загрузки ключа %s в Redis: %v", key, err)
	} else {
		logging("ключ [%s] успешно загружен в Redis", key)
	}

	return nil
}
