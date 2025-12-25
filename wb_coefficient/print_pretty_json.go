package main

import "encoding/json"

// Функция, которая выводит полученный json от api WB в красивом читаемом виде.
func printPrettyJson(data any) {
	if prettyJson, err := json.MarshalIndent(data, "", "  "); err == nil {
		logging("prettyJson data [len: %d]:\n%s", len(prettyJson), string(prettyJson))
	} else {
		logging("%s ошибка парсинга данных в prettyJson: %v", EmojiWarning, err)
	}
}
