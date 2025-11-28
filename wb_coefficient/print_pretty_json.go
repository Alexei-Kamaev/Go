package main

import "encoding/json"

func printPrettyJson(data any) {
	if prettyJson, err := json.MarshalIndent(data, "", "  "); err == nil {
		logging("prettyJson data [len: %d]:\n%s", len(prettyJson), string(prettyJson))
	} else {
		logging("ошибка парсинга данных в prettyJson: %v", err)
	}
}
