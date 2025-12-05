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
		return fmt.Errorf("не найден url [coef_url] для получения коэффициентов в конфигурации приложения")
	}
	if apiTokenWB == "" {
		return fmt.Errorf("пустой api-token в конфигурации приложения")
	}
	if len(appConfig.AllWarehouses) == 0 {
		return fmt.Errorf("пустой список складов в конфигурации приложения")
	}
	url += "?warehouseIDs=" + strings.Join(appConfig.AllWarehouses, ",")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("возникла ошибка при создании запроса: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+apiTokenWB)
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
			appConfig.Working = false
			// остановка приложения
			// тут необходимо реализовать отправку сообщения в ТГ админу
			// пока нет интернета, чтобы узнать chatID && apitoken
			return fmt.Errorf("прекращена работа приложения из-за ответ сервера: %d", resp.StatusCode)
		}
		logging("API WB вернуло ошибку для повтора: [%d]\n%s", resp.StatusCode, string(body))
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
