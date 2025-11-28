package main

import (
	"fmt"
	"strings"
	"time"
)

func logging(data string, args ...any) {

	logMutex.Lock()
	defer logMutex.Unlock()

	var buffer strings.Builder

	timestamp := time.Now().Format("15:04:05.000")

	buffer.WriteString("[")
	buffer.WriteString(timestamp)
	buffer.WriteString("] ")

	if len(args) > 0 {
		fmt.Fprintf(&buffer, data, args...)
	} else {
		buffer.WriteString(data)
	}

	buffer.WriteString("\n")

	logs += buffer.String()
}
