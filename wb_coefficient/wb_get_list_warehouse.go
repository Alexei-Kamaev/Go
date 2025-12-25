package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func getListWarehouseWB(list *map[int64]string) error {

	if appConfig.Token == "" {
		return fmt.Errorf("ошибка при получении списка складов ВБ, api-token пустой")
	}

	url, ok := appConfig.URL["get_list_whid"]

	if !ok {
		return fmt.Errorf("ошибка при получении списка складов, url пустой")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("возникла ошибка при создании запроса для получения списка складов ВБ: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+appConfig.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка в запросе при получении списка складов: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при чтении тела ответа в функции получения списка складов: %v", err)
	}

	if resp.StatusCode == http.StatusOK {

		if len(body) == 0 {
			return fmt.Errorf("при получении списка складов вернулся пустой ответ от сервера: [%d]", resp.StatusCode)
		}

		var result []WarehouseListID

		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("при получении списка складов возникла ошибка при распарсивании данных: %v", err)
		}

		var data strings.Builder

		for _, v := range result {

			if !v.IsActive {
				continue
			}

			fmt.Fprintf(&data, "%s\nАдрес: %s\nРаботает: %s", v.Name, v.Address, v.WorkTime)

			(*list)[v.ID] = data.String()

			data.Reset()
		}

	} else {

		if len(body) == 0 {
			return fmt.Errorf("при получении списка складов вернулся пустой ответ от сервера: [%d]", resp.StatusCode)
		}

		return fmt.Errorf("при получении списка складов, API WB вернуло ошибку: [%d]\n%s",
			resp.StatusCode, string(body))
	}

	return nil
}
