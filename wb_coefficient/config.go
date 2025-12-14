package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Функция, которая проверяет наличие конфига в Редис:
// - если конфига нет, то загружает.
func loadConfigFromJson() error {

	return readConfigFile()
}

func readConfigFile() error {

	cfg, err := os.ReadFile(configFile)
	if err != nil {

		if os.IsNotExist(err) {
			return fmt.Errorf("конфигурация приложения == nil, отсутствует файл [%s] для загрузки конфигурации в Редис! ", configFile)
		}

		return fmt.Errorf("ошибка чтения файла при загрузке конфигурации в Редис: %w", err)
	}

	if len(cfg) == 0 {
		return fmt.Errorf("конфигурационный файл успешно прочитан, но он ПУСТОЙ!")
	}

	return parseAndLoadConfigToRedis(cfg)
}

func parseAndLoadConfigToRedis(cfgApp []byte) error {

	var fullData AppConfig

	if err := json.Unmarshal(cfgApp, &fullData); err != nil {
		return fmt.Errorf("возикла ошибка при парсинге конфигурации приложения: %v", err)
	}

	mainConfig := fullData
	mainConfig.Clients = nil

	mainJSON, err := json.Marshal(mainConfig)
	if err != nil {
		return fmt.Errorf("ошибка при сериализации основного конфига: %v", err)
	}

	if err := setStringRedis(appNameInRedis, string(mainJSON)); err != nil {
		return fmt.Errorf("при загрузке готового конфига в Редис, возникла ошибка: %v", err)
	}

	var warehouseMap = make(map[string]bool)
	var successLoadToRedis int
	for clientID, clientData := range fullData.Clients {
		clientJSON, err := json.Marshal(clientData)
		if err != nil {
			logging("ошибка при сериализации конфига клиента [%s]: %v", clientID, err)
			continue
		}

		clientKey := "coefbot_client_" + clientID
		if err := setStringRedis(clientKey, string(clientJSON)); err != nil {
			logging("при загрузке клиентского [%s] конфига в Редис, возникла ошибка: %v", clientID, err)
			continue
		}

		for box := range clientData.BoxData {
			warehouseMap[fmt.Sprintf("%d", box)] = true
		}

		for mono := range clientData.MonoData {
			warehouseMap[fmt.Sprintf("%d", mono)] = true
		}

		successLoadToRedis++
	}

	logging("конфиг загружен в Редис: основной + %d/%d клиентов",
		successLoadToRedis, len(fullData.Clients))

	if successLoadToRedis < len(fullData.Clients) {
		logging("ВНИМАНИЕ: не все клиенты загружены в Redis!")
	}

	appConfig = &fullData

	if appConfig.AllWarehouses == nil {
		appConfig.AllWarehouses = make([]string, 0, len(warehouseMap))
	} else {
		appConfig.AllWarehouses = appConfig.AllWarehouses[:0]
	}

	for whID := range warehouseMap {
		appConfig.AllWarehouses = append(appConfig.AllWarehouses, whID)
	}

	return nil
}
