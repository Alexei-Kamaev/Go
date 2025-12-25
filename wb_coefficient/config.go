package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Функция, которая проверяет наличие конфига в Редис:
// - если конфига нет, то загружает.
func loadConfigFromJson() error {

	return readConfigFile()
}

// Функция, которая читает файл и возвращает ошибку парсинга этого файла.
func readConfigFile() error {

	// читаем файл
	cfg, err := os.ReadFile(configFile)
	if err != nil {
		// если нет файла -> ошибка
		if os.IsNotExist(err) {
			return fmt.Errorf("конфигурация приложения == nil, отсутствует файл [%s] для загрузки конфигурации в Редис! ", configFile)
		}
		// ошибка работы с файлом
		return fmt.Errorf("ошибка чтения файла при загрузке конфигурации в Редис: %w", err)
	}

	// на случай, если прочитанный файл был пустой
	if len(cfg) == 0 {
		return fmt.Errorf("конфигурационный файл успешно прочитан, но он ПУСТОЙ!")
	}

	// возвращаем не ошибку функции, а результат функции парсинга прочитанного файла
	return parseAndLoadConfigToRedis(cfg)
}

// Функция, которая парсит данные и загружает конфигурацию приложение и клиента
func parseAndLoadConfigToRedis(cfgApp []byte) error {

	// для начала парсим все данные, и клиента, и приложения
	var fullData AppConfig
	if err := json.Unmarshal(cfgApp, &fullData); err != nil {
		return fmt.Errorf("возикла ошибка при парсинге конфигурации приложения: %v", err)
	}

	// копируем структуру с данными и обнуляем клиентскую часть
	mainConfig := fullData
	mainConfig.Clients = nil

	// маршалим данные для загрузки строкой в Редис
	mainJSON, err := json.Marshal(mainConfig)
	if err != nil {
		return fmt.Errorf("ошибка при сериализации основного конфига: %v", err)
	}
	// загрузка конфигурации приложения в Редис
	if err := setStringRedis(appNameInRedis, string(mainJSON)); err != nil {
		return fmt.Errorf("при загрузке готового конфига в Редис, возникла ошибка: %v", err)
	}

	// загрузка клиентов в Редис
	// формирование из данных клиентов списка опрашиваемых складов
	// формирование сортированного слайса списка клиентов, чтобы public точно был последним
	var (
		warehouseMap       = make(map[string]bool)
		successLoadToRedis int
	)

	for clientID, clientData := range fullData.Clients {

		// маршалим данные для загрузки в Редис
		clientJSON, err := json.Marshal(clientData)
		if err != nil {
			logging("%s ошибка при сериализации конфига клиента [%s]: %v",
				EmojiError, clientID, err)
			continue
		}

		// загрузка в Редис клиентской конфигурации с определённым ключом
		clientKey := "coefbot_client_" + clientID
		if err := setStringRedis(clientKey, string(clientJSON)); err != nil {
			logging("%s при загрузке клиентского [%s] конфига в Редис, возникла ошибка: %v",
				EmojiError, clientID, err)
			continue
		}

		// перебираем все клиентские склады в мапу для уникальности
		for box := range clientData.BoxData {
			warehouseMap[fmt.Sprintf("%d", box)] = true
		}

		// перебираем все клиентские склады в мапу для уникальности
		for mono := range clientData.MonoData {
			warehouseMap[fmt.Sprintf("%d", mono)] = true
		}

		successLoadToRedis++
	}

	logging("%s конфиг загружен в Редис: основной + %d/%d клиентов",
		EmojiSuccess, successLoadToRedis, len(fullData.Clients))

	if successLoadToRedis < len(fullData.Clients) {
		logging("%s ВНИМАНИЕ: не все клиенты загружены в Redis!", EmojiWarning)
	}

	appConfig = &fullData

	// создаём или обнуляем срез с опрашиваемыми складами
	if appConfig.AllWarehouses == nil {
		appConfig.AllWarehouses = make([]string, 0, len(warehouseMap))
	} else {
		appConfig.AllWarehouses = appConfig.AllWarehouses[:0]
	}

	// создаём сортированный срез с клиентами, чтобы сдвинуть public на последнее место
	appConfig.AllActiveClients = nil

	// перебираем клиентов, автивных добавляем в срез для рассылки
	for client, clientData := range appConfig.Clients {
		if !clientData.IsActive {
			continue
		}
		appConfig.AllActiveClients = append(appConfig.AllActiveClients, client)
	}

	// сортируем клиентов и записываем их в конфигурацию приложения
	sort.Strings(appConfig.AllActiveClients)

	// записываем список складов в конфигурацию приложения
	appConfig.AllWarehouses = nil
	for whID := range warehouseMap {
		appConfig.AllWarehouses = append(appConfig.AllWarehouses, whID)
	}

	// сортируем список складов по алфавиту
	sort.Strings(appConfig.AllWarehouses)

	return nil
}
