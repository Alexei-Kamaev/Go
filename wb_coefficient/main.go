package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

func init() {

	argsCount := len(os.Args)

	if argsCount < 3 {
		log.Fatalf("не хватает аргументов для запуска приложения, получено %d аргументов, для запуска необходимо 2", argsCount-1)
	}

	cfgRedis = &redisConfig{
		Addr:     os.Args[1],
		Password: os.Args[2],
		DB:       0,
		TimeOut:  5 * time.Second,
	}

	var err error

	if redisClient, err = checkRedisConnection(); err != nil {
		log.Fatalln(err)
	}

	if err := checkConfigAppRedis(); err != nil {
		log.Fatalln(err)
	}

}

func main() {

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

	if redisClient != nil {
		defer redisClient.Close()
	}

	addLog("приложение успешно запущено")

	var maxWorkers int

	if cfgApp.MaxWorkers == 0 {
		maxWorkers = 3
	} else {
		maxWorkers = cfgApp.MaxWorkers
	}

	if len(cfgApp.Clients) == 0 {
		addLog("нет клиентов для обработки")
		return
	}

	if maxWorkers > len(cfgApp.Clients) {
		maxWorkers = len(cfgApp.Clients)
	}

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(maxWorkers)

	for client := range cfgApp.Clients {
		client := client
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				processing(client)
				return nil
			}
		})
	}

	if err := g.Wait(); err != nil {
		addLog("ошибка при выполнении: %v", err)
	}

	addLog("приложение успешно завершено")

}

func addLog(logData string, args ...any) {

	logMutex.Lock()
	defer logMutex.Unlock()

	var buffer strings.Builder

	timestamp := time.Now().Format("15:04:05.000")
	buffer.WriteString("[")
	buffer.WriteString(timestamp)
	buffer.WriteString("] ")

	if len(args) > 0 {
		fmt.Fprintf(&buffer, logData, args...)
	} else {
		buffer.WriteString(logData)
	}

	buffer.WriteString("\n")

	logs += buffer.String()

}
