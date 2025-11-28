package main

func clearData(data *[]Response) error {
	var cleanData []Response
	for _, v := range *data {
		// Приёмка для поставки доступна только при сочетании:
		// coefficient — 0 или 1
		// и allowUnload — true
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
