package main

import (
	"time"
)

// Функция, которая:
// принимает сырые данные,
// очищает от коэффициента -1
// возвращает ошибку.
func clearData(data *[]Response, mapDate map[string]string) error {

	var (
		src = *data
		i   = 0
	)

	for _, v := range src {
		// Приёмка для поставки доступна только при сочетании:
		// coefficient — 0 или 1
		// и allowUnload — true (этот параметр используется при формировании сообщения)
		if v.Coefficient >= 0 {
			// заполняем кэш с датами
			if _, ok := mapDate[v.Date]; !ok {
				if t, err := time.Parse(time.RFC3339, v.Date); err != nil {
					mapDate[v.Date] = v.Date
				} else {
					mapDate[v.Date] = t.Format("02.01")
				}
			}
			src[i] = v
			i++
		}
	}

	if tail := src[i:]; len(tail) > 0 {
		clear(tail) // очистка хвоста с ненужными данными
	}

	*data = src[:i] // укорачиваем слайс

	if appConfig.DebugMode {
		printPrettyJson(*data)
	}

	return nil
}
