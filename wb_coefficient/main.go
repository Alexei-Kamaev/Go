package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"
)

func main() {
	startTime := time.Now()
	// fmt.Printf("[SYSTEM] Процесс запущен: %s\n", startTime.Format("15:04:05.000"))
	defer func() {
		if r := recover(); r != nil {
			log.Printf("паника в основном потоке: %v", r)
			debug.PrintStack()
		}
	}()
	defer func() {
		if logs != "" {
			if _, err := fmt.Print(logs); err != nil {
				log.Printf("возникла ошибка при записи логов: %v", err)
			}
			logs = ""
		}
	}()
	logging("🚀 запускаемся...")
	// проверяем аргументы запуска
	if len(os.Args) < 3 {
		log.Fatalf("Для запуска приложения необходимо 3 аргумента: адрес Redis, Redis Password, API token WB!")
	}
	// получаем Redis адрес и пароль
	redisConfig = &RedisConfig{
		Addr:     os.Args[1],
		Password: os.Args[2],
		DB:       0,
		TimeOut:  3 * time.Second}
	// проверка соединения с Redis
	var err error
	logging("📡 подключаемся к Redis...")
	if redisClient, err = checkRedisConnection(); err != nil {
		log.Printf("ошибка запуска возникла в функции [init] при проверке соединения с Redis с полученными аргументами заупска приложения: %v", err)
		return
	}
	// если имеется аргумент токена, то получаем и api-токен
	if len(os.Args) > 3 {
		apiTokenWB = os.Args[3]
		if appConfig.DebugMode {
			token := apiTokenWB[:6] + "..."
			log.Printf("получен API токен WB в качестве аргумента запуска приложения: %s", token)
		}
	}
	// проверка и загрузка конфигурации в Redis
	logging("📋 загружаем конфигурацию...")
	if err := checkConfigInRedis(); err != nil {
		log.Printf("ошибка при проверке конфигурации в Redis: %v", err)
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	logging("приложение успешно запущено")
	if appConfig == nil {
		log.Println("конфигурация приложения не загружена!")
		return
	} else if appConfig.DebugMode {
		if data, err := json.MarshalIndent(appConfig, "", "  "); err == nil {
			logging("загруженная конфигурация:\n%s", string(data))
		} else {
			logging("%v", err)
		}
	}

	if apiTokenWB == "" {
		apiTokenWB = appConfig.Token
	}

	if !appConfig.Working {
		logging("приложение стоит на паузе параметр [working] в config.json")
	}

	for c := range appConfig.CountRequests {

		if !appConfig.Working {
			logging("приложение остановлено по ошибке ответа сервера")
			return
		}

		var data []Response

		start := time.Now()

		if err := getCoefWarehouses(&data); err != nil {
			logging("ошибка при получении коэффициентов:\n%v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		if err := clearData(&data); err != nil {
			logging("ошибка очистки данных от КФ [-1]:\n%v", err)
			continue
		}

		for client := range appConfig.Clients {
			if err := prepareMessages(data, client); err != nil {
				logging("у клиента %s ошибка при формировании сообщений: %v", client, err)
			}
		}

		sleep := time.Duration(appConfig.PauseRequests)*time.Second - time.Since(start)
		if sleep > 0 && c < appConfig.CountRequests-1 {
			time.Sleep(sleep)
		}
	}

	logging("приложение завершено [%.3f сек]", time.Since(startTime).Seconds())
}
