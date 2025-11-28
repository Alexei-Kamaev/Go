package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func sendTextMessage(text, client string, whid int) error {

	if _, ok := appConfig.Clients[client]; !ok {
		return fmt.Errorf("клиента %s нет в конфигурации", client)
	}

	var fullUrl string

	encodedText := url.QueryEscape(text)

	if client == "public" {
		fullUrl += appConfig.Clients[client].ApiData["base_data"]
		fullUrl += appConfig.Clients[client].ApiData[fmt.Sprintf("%d", whid)]
		fullUrl += appConfig.Clients[client].ChatData["base_data"]
		fullUrl += appConfig.Clients[client].ChatData[fmt.Sprintf("%d", whid)]
		fullUrl += "&text=" + encodedText
		fullUrl += "&parse_mode=HTML"
	}

	if appConfig.DebugMode {
		logging("%s", fullUrl)
	}

	resp, err := http.Get(fullUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: %d", resp.StatusCode)
	}

	time.Sleep(10 * time.Millisecond)

	return nil
}
