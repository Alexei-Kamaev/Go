package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func sendTextMessage(text, client string, whid int) error {

	if client != appConfig.Admin {
		if _, ok := appConfig.Clients[client]; !ok {
			return fmt.Errorf("ошибка при отправке сообщения в ТГ, клиента %s нет в конфигурации", client)
		}
	}

	var fullUrl strings.Builder

	encodedText := url.QueryEscape(text)

	if baseURL, ok := appConfig.URL["base_tlg"]; ok {
		fullUrl.WriteString(baseURL)
	} else {
		return fmt.Errorf("при формировании URL для отправки сообщения в чат, возникла ошибка: отсутствует в конфигурации ключ [appConfig.URL[base_tlg]]")
	}

	if client == appConfig.Admin && whid == 0 {
		fullUrl.WriteString(appConfig.BotToken)
		fullUrl.WriteString(appConfig.URL["base_chat"])
		fullUrl.WriteString(client)
	} else if client == "public" {
		fullUrl.WriteString(appConfig.Clients[client].ApiData[fmt.Sprintf("%d", whid)])
		fullUrl.WriteString(appConfig.URL["base_chat"])
		fullUrl.WriteString(appConfig.Clients[client].ChatData["public"])
		fullUrl.WriteString(appConfig.Clients[client].ChatData[fmt.Sprintf("%d", whid)])
	} else {
		fullUrl.WriteString(appConfig.Clients[client].TGToken)
		fullUrl.WriteString(appConfig.URL["base_chat"])
		fullUrl.WriteString(client)
	}

	fullUrl.WriteString("&text=" + encodedText)
	fullUrl.WriteString("&parse_mode=HTML")

	resp, err := httpClient.Get(fullUrl.String())
	if err != nil {
		return fmt.Errorf("ошибка Telegram запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logging("сформированный URL:\n%s", fullUrl.String())
		return fmt.Errorf("telegram API вернуло ошибку при отправке сообщения клиенту [%s] склад [%d]: %d", client, whid, resp.StatusCode)
	}

	if client != "public" {
		logging("сообщение клиенту [%s] по складу [%d] успешно отправлено", client, whid)
	}

	time.Sleep(20 * time.Millisecond)

	return nil
}
