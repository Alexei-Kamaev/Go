package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func checkConfigInRedis() error {
	// проверяем наличие ключа с конфигурацией в Redis
	if data, err := getStringRedis(configKey); err != nil {
		// ошибка соединения с Redis, возвращаем ошибку
		return err
	} else if len(data) == 0 || checkAgeConfigFile() {
		// если вдруг переменная пустая или конфигурация была обновлена
		// парсим конфигурацию из фала
		// заружаем в Redis конфигурацию
		config, err := readConfigFile()
		if err != nil {
			return err
		}
		// парсим данные из config.json
		if err := json.Unmarshal(config, &appConfig); err != nil {
			return fmt.Errorf("возникла ошибка при парсинге конфигурации приложения: %v", err)
		}
		// загружаем строку в Redis
		if err := setStringRedis(configKey, string(config)); err != nil {
			return fmt.Errorf("возникла ошибка при записи конфигурации приложения в Redis: %v", err)
		}
	} else {
		if err := json.Unmarshal([]byte(data), &appConfig); err != nil {
			return fmt.Errorf("ошибка парсинга данных из Redis: %v", err)
		}
	}
	return nil
}

func checkAgeConfigFile() bool {
	info, err := os.Stat(configFile)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			logging("конфигурационный файл не найден: %q", configFile)
		case os.IsPermission(err):
			logging("нет прав на чтение файла: %q", configFile)
		default:
			logging("ошибка при проверке файла %q: %v", configFile, err)
		}
		return false
	}
	modTime := info.ModTime()
	age := time.Since(modTime)
	return age < maxAgeConfigFile
}

func readConfigFile() ([]byte, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("невозможно прочитать отсутствующий файл %s", configFile)
		}
		return nil, fmt.Errorf("ошибка при чтении файла %s: %w", configFile, err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("пустой конфигурационный файл: %s", configFile)
	}
	return data, nil
}
