package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	configFile = "config.json"
)

var (
	configApp  config
	list       []warehouse
	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}
)

type config struct {
	Key          string `json:"key"`
	Mode         bool   `json:"debug"`
	UrlPingToken string `json:"url_ping_token"`
	UrlGetList   string `json:"url_get_list"`
}

type warehouse struct {
	ID              int    `json:"ID"`
	Name            string `json:"name"`
	Address         string `json:"address"`
	WorkTime        string `json:"workTime"`
	IsActive        bool   `json:"isActive"`
	IsTransitActive bool   `json:"isTransitActive"`
}

func main() {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("паника восстановлена: %v", r)
		}
	}()

	var err error

	if configApp, err = readConfig(); err != nil {
		log.Println(err)
		return
	}

	status, err := checkToken()

	if err != nil {
		log.Println(err)
		return
	} else if !status {
		log.Println("api-token не проходит валидацию")
		return
	}

	list, err = getListWarehouse()

	if err != nil {
		log.Println(err)
		return
	}

	if len(list) == 0 {
		log.Println("запрос выполнен без ошибок, но список складов пустой")
		return
	}

	if len(os.Args) > 1 {
		findWarehouse(os.Args[1])
	} else {
		for _, w := range list {
			fmt.Printf("Name: %s\nID: %d\nActive: %t\n", w.Name, w.ID, w.IsActive)
		}
	}

}

func findWarehouse(search string) {

	search = strings.ToLower(search)
	status := true

	for _, v := range list {

		if strings.Contains(strings.ToLower(v.Name), search) || strings.Contains(strings.ToLower(v.Address), search) {
			status = false
			fmt.Printf("Name: %s\nAddress: %s\nID: %d\nActive: %t\n",
				v.Name, v.Address, v.ID, v.IsActive)
		}

	}

	if status {
		fmt.Println("по Вашему запросу ничего не найдено, измените параметры поиска и повторите")
	}

}

func getListWarehouse() (w []warehouse, e error) {

	r, err := getRequest(configApp.UrlGetList)

	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка складов: %v", err)
	}

	debugPrint("ответ сервера для получения списка складов", r)

	err = json.Unmarshal(r, &w)

	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа при получении списка складов: %w", err)
	}

	return w, nil

}

func checkToken() (s bool, e error) {

	type pingResponse struct {
		Status string `json:"Status"`
	}

	r, err := getRequest(configApp.UrlPingToken)

	if err != nil {
		return false, fmt.Errorf("ошибка при проверке токена: %w", err)
	}

	debugPrint("ответ сервера при проверке токена", r)

	var response pingResponse

	err = json.Unmarshal(r, &response)

	if err != nil {
		return false, fmt.Errorf("ошибка парсинга ответа при проверке токена: %w", err)
	}

	return response.Status == "OK", nil

}

func getRequest(url string) (b []byte, e error) {

	r, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании GET запроса: %w", err)
	}

	r.Header.Set("Authorization", "Bearer "+configApp.Key)

	resp, err := httpClient.Do(r)

	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении GET запроса: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP статус %d", resp.StatusCode)
	}

	b, err = io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	return b, nil

}

func debugPrint(operation string, data []byte) {

	if !configApp.Mode {
		return
	}

	var buf bytes.Buffer

	if json.Indent(&buf, data, "  ", "    ") == nil {
		fmt.Printf("%s:\n%s\n", operation, buf.String())
	} else {
		fmt.Printf("%s:\n%s\n", operation, string(data))
	}

}

func readConfig() (cfg config, err error) {

	data, err := os.ReadFile(configFile)

	if err != nil {
		return cfg, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	err = json.Unmarshal(data, &cfg)

	if err != nil {
		return cfg, fmt.Errorf("во время парсинга файла json: %s возникла ошибка: %w",
			configFile, err,
		)
	}

	return cfg, nil

}
