package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	configKey        = "public_bot"
	configFile       = "config.json"
	maxAgeConfigFile = 2 * time.Minute
)

var (
	logging      func(string, ...any)
	logs         strings.Builder
	logsCapacity = 100 * 1024 // 100KB
	logMutex     sync.Mutex
)

func main() {
	// замеряем время работы приложения
	startGlobalTime := time.Now()
	// выделяем память под логи
	logs.Grow(logsCapacity)
	// функция для логирования
	logging = func(data string, args ...any) {
		// блокировка записи, для однопоточного процесса не нужно,
		// при масштабировании будет полезно от гонки данных
		logMutex.Lock()
		defer logMutex.Unlock()
		// формат таймштампа
		timeStamp := time.Now().Format("15:04:05.000")
		// формирование и склейка лог-сообщений
		fmt.Fprintf(&logs, "[%s] ", timeStamp)
		if len(args) > 0 {
			fmt.Fprintf(&logs, data, args...)
		} else {
			logs.WriteString(data)
		}
		logs.WriteByte('\n')
	}
	// первый лог
	logging("🚀 запускаемся...")
	// отлавливаем паники и падения приложения
	defer func() {
		if r := recover(); r != nil {
			log.Printf("паника в основном потоке: %v", r)
			debug.PrintStack()
		}
	}()
	// сброс всех логов при завершении приложения
	defer func() {
		logMutex.Lock()
		defer logMutex.Unlock()
		// проверка на пустые логи
		if logs.Len() == 0 {
			return
		}
		// дописываем в лог статистику использования [logs strings.Builder]
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Fprintf(&logs, "[%s] [STATS] Capacity: %d, Length: %d\n",
			timestamp, logs.Cap(), logs.Len())
		// сброс всех логов в StdOut
		if _, err := fmt.Print(logs.String()); err != nil {
			log.Printf("возникла ошибка при записи логов: %v", err)
		}
		// обнуляем логи для дальнейшего использования
		logs.Reset()
	}()
	// логируем завершение работы приложения
	defer func() {
		logging("приложение завершено [%.3f сек]", time.Since(startGlobalTime).Seconds())
	}()
	// получаем необходимые аргументы для запуска приложения
	if len(os.Args) < 2 {
		logging("для запуска приложения необходимо минимум 2 аргумента: адрес Redis и Redis Password!")
		return
	}
	// инициализация Redis клиента
	redisConfig = &RedisConfig{
		Addr:     os.Args[1],
		Password: os.Args[2],
		DB:       0,
		TimeOut:  3 * time.Second}
	var err error
	logging("📡 подключаемся к Redis...")
	if redisClient, err = checkRedisConnection(); err != nil {
		logging("ошибка запуска возникла при проверке подключения к Redis с полученными аргументами запуска приложения: %v", err)
		return
	}
	if len(os.Args) > 3 {
		apiTokenWB = os.Args[3]
		if appConfig.DebugMode {
			token := apiTokenWB[:6] + "..."
			logging("получен API токен WB в качестве аргумента запуска приложения: %s", token)
		}
	}
	logging("📋 загружаем конфигурацию...")
	if err := checkConfigInRedis(); err != nil {
		logging("ошибка при проверке конфигурации в Redis: %v", err)
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
		logging("приложение на паузе параметр [working] в config.json")
	}
	for c := range appConfig.CountRequests {
		if !appConfig.Working {
			// надо подумать над логикой остановки приложения во время работы
			// пока неверная логика, приложение полностью останавливается из-за одной ошибки
			logging("приложение было остановлено в процессе работы, по ошибке ответа от сервера")
			return
		}
		var data []Response
		startIterationTime := time.Now()
		logging("%d круг", c+1)
		if err := getCoefWarehouses(&data); err != nil {
			logging("ошибка при получении коэффициентов:\n%v", err)
			continue
		}
		if err := clearData(&data); err != nil {
			logging("ошибка очистки данных от КФ [-1]:\n%v", err)
			continue
		}
		for client := range appConfig.Clients {
			if !appConfig.Clients[client].IsActive {
				logging("у клиента %s установлен статус [%t]", client, appConfig.Clients[client].IsActive)
				continue
			}
			pause := appConfig.Clients[client].Pause
			if pause > 0 {
				logging("у клиента %s пауза по api %dмс", client, pause)
				if pause-600 > 0 {
					pause -= 600
				} else {
					pause = 0
				}
				continue
			}
			if err := prepareMessages(data, client); err != nil {
				logging("у клиента %s ошибка при формировании сообщений: %v", client, err)
			}
		}
		sleep := time.Duration(appConfig.PauseRequests)*time.Second - time.Since(startIterationTime)
		logging("общее время: %.3f, текущее время: %.3f, остаток паузы в данной итерации: %v",
			time.Since(startGlobalTime).Seconds(), time.Since(startIterationTime).Seconds(), sleep)
		if sleep > 0 && c < appConfig.CountRequests-1 {
			time.Sleep(sleep)
		}
	}
}
