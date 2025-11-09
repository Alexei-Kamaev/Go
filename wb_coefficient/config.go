package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func checkConfigAppRedis() error {
	key := "config_" + appName
	if data, err := getStringValueRedis(key); err != nil {
		return err
	} else if len(data) == 0 {
		return loadConfigFile(key)
	} else {
		return parseConfigApp([]byte(data))
	}
}

func loadConfigAppRedis(data, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfgRedis.TimeOut)
	defer cancel()
	err := redisClient.Set(ctx, key, data, time.Duration(redisExpiration)*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации приложения в Redis: %v", err)
	}
	return parseConfigApp([]byte(data))
}

func parseConfigApp(data []byte) error {
	err := json.Unmarshal(data, &cfgApp)
	if err != nil {
		return fmt.Errorf("ошибка JSON парсинга конфигурации приложения: %v", err)
	}
	return nil
}

func loadConfigFile(key string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("ошибка: файла %s не существует", configFile)
		}
		return fmt.Errorf("ошибка при чтении файла: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("ошибка: файл %s пустой", configFile)
	}
	return loadConfigAppRedis(string(data), key)
}
