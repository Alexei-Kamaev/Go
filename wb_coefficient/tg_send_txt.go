package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

// Функция для отправки сообщения в Телеграмм.
func sendTextMessage(text, client string, whid int) error {

	// проверка на сломанный топик в группе
	if client == "public" {
		if _, ok := appConfig.Clients[client].PauseWHID[whid]; ok {
			logging("пропуск отправки сообщения в топик группы, топик сломан: [%d] [%s]",
				whid, appConfig.Clients[client].ChatData[fmt.Sprintf("%d", whid)])
			return nil
		}
	}

	// собираем url по частям
	var fullUrl strings.Builder
	// проверка на наличие базовой части url в конфигурации приложения
	if baseURL, ok := appConfig.URL["base_tlg"]; ok {
		fullUrl.WriteString(baseURL)
	} else {
		return fmt.Errorf("при формировании URL для отправки сообщения в чат, возникла ошибка: отсутствует в конфигурации ключ [appConfig.URL[base_tlg]]")
	}
	// формирование url для отправки служебного сообшения для Админа
	if client == appConfig.Admin && whid == 0 {
		fullUrl.WriteString(appConfig.BotToken)
		fullUrl.WriteString(appConfig.URL["base_chat"])
		fullUrl.WriteString(client)

		// если это не Админ, формируем url для отправки в Public
	} else if client == "public" {
		fullUrl.WriteString(appConfig.Clients[client].ApiData[fmt.Sprintf("%d", whid)])
		fullUrl.WriteString(appConfig.URL["base_chat"])
		fullUrl.WriteString(appConfig.Clients[client].ChatData["public"])
		fullUrl.WriteString(appConfig.Clients[client].ChatData[fmt.Sprintf("%d", whid)])

		// эта часть формирует для обычного клиента url
	} else {
		fullUrl.WriteString(appConfig.Clients[client].TGToken)
		fullUrl.WriteString(appConfig.URL["base_chat"])
		fullUrl.WriteString(client)
	}

	// кодируем сообщение и доклеиваем его в url
	encodedText := url.QueryEscape(text)
	fullUrl.WriteString("&text=" + encodedText)
	fullUrl.WriteString("&parse_mode=HTML")

	// выполняем Гет-запрос
	resp, err := httpClient.Get(fullUrl.String())
	if err != nil {
		return fmt.Errorf("ошибка Telegram запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при разборе ответа от телеграм: %v", err)
	}

	// разбор ответа от Телеграмм
	var result TelegramResponse
	// парсим ответ от ТГ
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("не удалось распарсить ответ от Telegram: %v", err)
	}
	// обработка всех ответов от ТГ
	if !result.Status {
		// логирование ответа и url
		logging("сформированный URL:\n%s", fullUrl.String())
		logging("Telegram ответ [статус %d]:\n%s", resp.StatusCode, string(body))

		// обработка ошибки частого запроса
		if result.ErrorCode == 429 && result.Parameters.RetryAfter > 0 {
			// запись паузы в конфигурацию клиента
			if clientConfig, ok := appConfig.Clients[client]; ok {
				clientConfig.Pause = result.Parameters.RetryAfter
				appConfig.Clients[client] = clientConfig
				// логирование
				logging("%s клиент %s получил паузу от ТГ: %d сек",
					EmojiWarning, client, result.Parameters.RetryAfter)
			}
			// возвращаем ошибку в ответе
			return fmt.Errorf("клиент %s получил паузу от ТГ: %d сек",
				client, result.Parameters.RetryAfter)
		}

		// эта часть обрабатывает ошибку поломки топика в группе
		if result.ErrorCode == 400 && client == "public" {
			// установка паузы для сломанного топика, ожидаем действий от Админа
			appConfig.Clients[client].PauseWHID[whid] = true
			topik := appConfig.Clients[client].ChatData[fmt.Sprintf("%d", whid)]
			logging("пропуск отправки сообщения в топик группы, топик сломан: [%d] [%s]",
				whid, topik)

			return fmt.Errorf("топик сломан: [%d] [%s]", whid, topik)
		}

		return fmt.Errorf("telegram API вернуло ошибку при отправке сообщения клиенту [%s]: %d", client, resp.StatusCode)
	}

	if client != "public" {
		// логируем отправку каждого клиента, кроме группы, её видно
		logging("%s сообщение [%s] успешно отправлено", EmojiSuccess, client)
	}

	// просто спим между отправками сообщений
	time.Sleep(50 * time.Millisecond)

	// счётчик отправок за иттерацию
	appConfig.AllCountSendMessages++

	return nil
}
