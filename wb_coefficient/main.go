package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	logMutex sync.Mutex
	logs     string
)

func init() {
	argsCount := len(os.Args)
	if argsCount < 3 {
		log.Fatalf("не хватает аргументов для запуска приложения, получено %d аргументов, для запуска необходимо 2", argsCount-1)
	}
	cfg := &redisConfig{
		Addr:     os.Args[1],
		Password: os.Args[2],
		DB:       0,
		TimeOut:  5 * time.Second,
	}
	var err error
	if redisClient, err = checkRedisConnection(cfg); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	defer func() {
		if logs != "" {
			if _, err := fmt.Print(logs); err != nil {
				log.Printf("возникла ошибка при записи логов: %v", err)
			}
			logs = ""
		}
	}()
	defer redisClient.Close()
	addLog("приложение успешно запущено")
}

func addLog(logData string, args ...any) {
	logMutex.Lock()
	defer logMutex.Unlock()
	var buffer strings.Builder
	timestamp := time.Now().Format("15:04:05.000")
	buffer.WriteString("[")
	buffer.WriteString(timestamp)
	buffer.WriteString("]")
	if len(args) > 0 {
		fmt.Fprintf(&buffer, logData, args...)
	} else {
		buffer.WriteString(logData)
	}
	buffer.WriteString("\n")
	logs += buffer.String()
}
