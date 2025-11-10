package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func custom_get_request(url, token string, body any) (*http.Response, error) {

	if url == "" {
		return nil, fmt.Errorf("в функцию custom_get_request пришёл пустой url")
	}

	if token == "" {
		return nil, fmt.Errorf("в функцию custom_get_request пришёл пустой api-токен")
	}

	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("ошибка маршалинга Get запроса: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("GET", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("ошибка в функции custom_get_request при формировании запроса: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	time.Sleep(50 * time.Millisecond)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса в функции custom_get_request: %v", err)
	}
	// defer resp.Body.Close() необходимо закрывать в вызывающей функции, тут нельзя

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET запрос в функции custom_get_request завершился с ошибкой: %d", resp.StatusCode)
	}

	return resp, nil
}
