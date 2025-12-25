package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Функция, которая записывает в Редис ключ:строка.
func setStringRedis(key, data string) error {

	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.TimeOut)
	defer cancel()

	// время жизни ключа по умолчанию, пока не загружен конфиг
	var redisExp = 60

	// меняем время жизни ключа, когда есть это значение в конфиге
	if appConfig != nil {
		redisExp = appConfig.RedisExpiration
	}

	// для ключа склад ВБ ставим сутки
	if strings.HasPrefix(key, "warehouse_") {
		redisExp = 1440
	}

	err := redisClient.Set(ctx, key, data, time.Duration(redisExp)*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("ошибка загрузки ключа %s в Redis: %v", key, err)
	} else {
		logging("%s ключ [%s] успешно загружен в Redis", EmojiSuccess, key)
	}

	return nil
}
