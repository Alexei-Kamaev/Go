package main

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

func processing(client string) {

	startTime := time.Now()

	var (
		workerLog strings.Builder
		okToken   []string
	)

	log := func(data string, args ...any) {
		timestamp := time.Now().Format("15:04:05.000")
		workerLog.WriteString(fmt.Sprintf("[%s] %s: ", timestamp, client))
		if len(args) > 0 {
			fmt.Fprintf(&workerLog, data, args...)
		} else {
			workerLog.WriteString(data)
		}
		workerLog.WriteString("\n")
	}

	defer func() {

		var panicMsg string

		if r := recover(); r != nil {
			panicMsg = fmt.Sprintf("паника в горутине: %v", r)
			debug.PrintStack()
		}

		elapsed := time.Since(startTime)

		if panicMsg != "" {
			log(panicMsg)
		}

		log("затрачено времени: %.3f сек", elapsed.Seconds())

		if workerLog.Len() > 0 {
			logMutex.Lock()
			logs += workerLog.String()
			logMutex.Unlock()
		}

	}()

	log("начало обработки")

	if len(cfgApp.Clients[client].APIToken) > 1 {
		for i, v := range cfgApp.Clients[client].APIToken {
			if checkToken(v, log) {
				okToken = append(okToken, v)
			} else {
				log("у клиента не пингуется токен: %d", i)
			}
		}
	} else {
		if checkToken(cfgApp.Clients[client].APIToken[0], log) {
			okToken = append(okToken, cfgApp.Clients[client].APIToken[0])
		} else {
			log("у клиента всего один токен и он не пингуется")
		}
	}

	if len(okToken) == 0 {
		return
	}

	log("обработка завершена")
}
