package main

import (
	"encoding/json"
	"io"
)

func checkToken(token string, log func(data string, args ...any)) bool {

	type responseCheckToken struct {
		TS     string `json:"TS,omitempty"`
		Status string `json:"Status,omitempty"`
	}

	ok := false

	if url, exists := cfgApp.URL["ping"]; !exists {
		log("при проверке токена возникла ошибка с получением url из конфига в функции [checkToken]")
		return false
	} else {
		resp, err := custom_get_request(url, token, nil)
		if err != nil {
			log("%v", err)
			return false
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log("ошибка чтения ответа в функции [checkToken]: %v", err)
			return false
		}

		var result responseCheckToken

		if err = json.Unmarshal(bodyBytes, &result); err != nil {
			log("ошибка обработки ответа в функции [checkToken]: %v", err)
			return false
		}

		if cfgApp.DebugMode {
			prettyResult, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				log("ошибка форматирования красивого ответа в функции [checkToken]: %v", err)
			}
			log("\n=== ДАННЫЕ ОТВЕТА функции [checkToken] ===\n%s\n", string(prettyResult))
		}

		ok = result.Status == "OK"
	}

	return ok
}
