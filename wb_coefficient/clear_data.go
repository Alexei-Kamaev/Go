package main

// Функция, которая:
// принимает сырые данные,
// очищает от коэффициента -1
// возвращает ошибку.
func clearData(data *[]Response) error {
	var cleanData []Response
	for _, v := range *data {
		// Приёмка для поставки доступна только при сочетании:
		// coefficient — 0 или 1
		// и allowUnload — true (этот параметр используется при формировании сообщения)
		if v.Coefficient >= 0 {
			cleanData = append(cleanData, v)
		}
	}
	*data = cleanData
	if appConfig.DebugMode {
		printPrettyJson(*data)
	}
	return nil
}
