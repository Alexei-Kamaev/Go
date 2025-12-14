package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	configFile          = "config.json"
	minimalPauseRequest = 15
	appNameInRedis      = "public_bot"
)

var (
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownChan = make(chan os.Signal, 1)
	logging      func(string, ...any)
	logs         strings.Builder
	logsCapacity = 2 * 1024
	logMutex     sync.Mutex
	redisClient  *redis.Client
	redisConfig  *RedisConfig
	appConfig    *AppConfig
	httpClient   = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableCompression:    false,
			ResponseHeaderTimeout: 8 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   5,
			IdleConnTimeout:       30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
			MaxConnsPerHost:       2,
		},
	}
)

func main() {

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	startGlobalTime := time.Now()

	logs.Grow(logsCapacity)

	logging = func(data string, args ...any) {

		logMutex.Lock()
		defer logMutex.Unlock()

		timeStamp := time.Now().Format("15:04:05.000")

		fmt.Fprintf(&logs, "[%s] ", timeStamp)
		if len(args) > 0 {
			fmt.Fprintf(&logs, data, args...)
		} else {
			logs.WriteString(data)
		}
		logs.WriteByte('\n')
	}

	go func() {
		sig := <-shutdownChan
		logging("получен сигнал завершения: %v", sig)
		cancel()
		time.Sleep(2 * time.Second)
	}()

	logging("🚀 запускаемся...")

	defer func() {
		if r := recover(); r != nil {
			log.Printf("паника в основном потоке: %v", r)
			debug.PrintStack()
		}
	}()

	defer func() {
		logMutex.Lock()
		defer logMutex.Unlock()

		if logs.Len() == 0 {
			return
		}

		if _, err := fmt.Print(logs.String()); err != nil {
			log.Printf("возникла ошибка при записи логов: %v", err)
		}
	}()

	defer func() {
		logging("приложение завершено [%.3f сек]", time.Since(startGlobalTime).Seconds())
	}()

	redisConfig = &RedisConfig{
		Addr:     os.Getenv("redisAddr"),
		Password: os.Getenv("redisPassword"),
		DB:       0,
		TimeOut:  3 * time.Second}
	var err error
	logging("📡 подключаемся к Redis...")
	if redisClient, err = checkRedisConnection(); err != nil {
		logging("ошибка запуска возникла при проверке подключения к Redis с полученными аргументами запуска приложения: %v", err)
		return
	}

	logging("📋 загружаем конфигурацию...")

	if err := loadConfigFromJson(); err != nil {
		logging("ошибка загрузки конфигурации: %v", err)
		return
	}

	if appConfig == nil {
		logging("КОНФИГ НЕ ЗАГРУЖЕН! appConfig is nil")
		return
	}

	if !appConfig.Working {
		logging("приложение на паузе параметр [working] в config.json")
		return
	}

	logging("приложение успешно запущено")

	var data = make([]Response, 0, 1024)

	for c := 0; ; c++ {

		data = data[:0]

		if ctx.Err() != nil {
			logging("получена команда остановки приложения")
			time.Sleep(100 * time.Millisecond)
			return
		}

		if !appConfig.Working {
			logging("приложение на паузе, ждем 300 секунд")
			for range 300 {
				if ctx.Err() != nil {
					logging("получена команда остановки приложения")
					time.Sleep(1 * time.Second)
					return
				}
				time.Sleep(1 * time.Second)
			}
			continue
		}

		startIterationTime := time.Now()

		if err := getCoefWarehouses(&data, appConfig.Token); err != nil {
			logging("ошибка при получении коэффициентов:\n%v", err)
			continue
		}

		logging("получено сырых данных: %d, capacity: %d", len(data), cap(data))

		if err := clearData(&data); err != nil {
			logging("ошибка очистки данных от КФ [-1]:\n%v", err)
			continue
		}

		for client, clientConfig := range appConfig.Clients {

			if len(appConfig.Clients[client].BoxData)+len(appConfig.Clients[client].MonoData) == 0 {
				logging("пропуск клиента [%s], нет складов в конфигурации", client)
				continue
			}

			if !clientConfig.IsActive {
				logging("пропуск клиента %s статус [%t]", client, clientConfig.IsActive)
				continue
			}

			if clientConfig.Pause > 0 {
				logging("у клиента %s пауза по api %dмс", client, clientConfig.Pause)
				updatedClient := clientConfig
				if updatedClient.Pause > 600 {
					updatedClient.Pause -= 600
				} else {
					updatedClient.Pause = 0
				}
				appConfig.Clients[client] = updatedClient
				logging("обновлена api пауза клиента %s: %dмс", client, updatedClient.Pause)
				continue
			}

			if err := prepareMessages(data, client); err != nil {
				logging("у клиента %s ошибка при формировании или отправке сообщения: %v", client, err)
			}
		}

		reload, err := checkExistsKeyInRedis(appNameInRedis)
		if err != nil {
			logging("%v", err)
		}
		if !reload {
			loadConfigFromJson()
		}

		pause := max(minimalPauseRequest, appConfig.PauseIteration)

		sleep := time.Duration(pause)*time.Second - time.Since(startIterationTime)

		logging("время работы цикла: %.3f, остаток от паузы %d сек: %.3f сек",
			time.Since(startIterationTime).Seconds(),
			pause,
			sleep.Seconds(),
		)

		logMutex.Lock()
		if logs.Len() > 0 {
			fmt.Print(logs.String())
			logs.Reset()
		}
		logMutex.Unlock()

		if sleep <= 0 {

			time.Sleep(100 * time.Millisecond)

		} else {

			seconds := int(sleep.Seconds())

			remainder := sleep - time.Duration(seconds)*time.Second

			for range seconds {
				if ctx.Err() != nil {
					logging("получен сигнал завершения приложения")
					return
				}
				time.Sleep(1 * time.Second)
			}

			if remainder > 0 {
				time.Sleep(remainder)
			}
		}
	}
}
