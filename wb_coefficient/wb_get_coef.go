package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Функция, которая:
// получает структуру,
// заполняет данными получаемыми от сервера ВБ,
// возвращает ошибку.
func getCoefWarehouses(data *[]Response) error {

	url, exists := appConfig.URL["coef_url"]
	if !exists {
		return fmt.Errorf("в функции получения коэффициентов не получен url [coef_url] из конфигурации приложения")
	}

	if appConfig.Token == "" {
		return fmt.Errorf("в функции получения коэффициентов не получен api-token из конфигурации приложения")
	}

	if len(appConfig.AllWarehouses) == 0 {
		appConfig.Working = false
		errorMsg := fmt.Sprintf("в функции получения коэффициентов получен len(%d) список складов из конфигурации приложения", len(appConfig.AllWarehouses))
		// если нет складов, то что-то пошло не так, ставим приложение на паузу и зовём админа
		if err := sendTextMessage(errorMsg, appConfig.Admin, 0); err != nil {
			logging("%v", err)
		}
		return fmt.Errorf("%s", errorMsg)
	}

	// доклеиваем склады в url для Гет запроса
	url += "?warehouseIDs=" + strings.Join(appConfig.AllWarehouses, ",")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("возникла ошибка при создании запроса: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+appConfig.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("возникла ошибка при чтении ответа от сервера: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// "retry_codes": {"500", "502", "503", "504"}
		if _, exists := appConfig.RetryCodes[resp.StatusCode]; !exists {

			if err := sendTextMessage(fmt.Sprintf("в приложении по получению коэффициентов приёмки WB получена в ответ на запрос ошибка: %v и сейчас приложение стоит на паузе", resp.StatusCode), appConfig.Admin, 0); err != nil {
				logging("%s возникла ошибка при отправке служебного сообщения: %v", EmojiWarning, err)
			}

			return fmt.Errorf("прекращена работа приложения из-за ответа сервера: %d", resp.StatusCode)
		}

		logging("%s API WB вернуло ошибку для повтора: [%d]\n%s",
			EmojiWarning, resp.StatusCode, string(body))

		time.Sleep(3 * time.Second)

		return fmt.Errorf("API WB вернуло ошибку: [%d]\n%s", resp.StatusCode, string(body))
	}

	if len(body) == 0 {
		time.Sleep(3 * time.Second)
		return fmt.Errorf("сервер вернул пустой ответ")
	}

	if appConfig.DebugMode {
		logging("сырой ответ от сервера:\n%s", string(body))
	}

	if err := json.Unmarshal(body, data); err != nil {
		return fmt.Errorf("ошибка парсинга ответа от сервера [Status: %d]: %v", resp.StatusCode, err)
	}

	if appConfig.DebugMode {
		printPrettyJson(data)
	}

	return nil
}
