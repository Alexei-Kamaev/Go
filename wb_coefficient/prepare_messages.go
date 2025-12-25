package main

import (
	"fmt"
	"strings"
)

// Функция, которая:
// принимает очищенные данные,
// формирует готовое сообщение для каждого клиента
// и отправляет каждому клиенту в чат.
func prepareMessages(data []Response, mapDate map[string]string, client string) error {

	// формируем мапу из слайса по ID склада
	dataByWarehouse := groupByWarehouse(data)

	var (
		message     strings.Builder
		infoMessage string
		previousLen int
	)

	// изначально прогоняем Монопаллеты, их меньше
	// итерируемся по складам из конфигурации клиента
	for whid, monoCoef := range appConfig.Clients[client].MonoData {

		// смотрим есть ли в мапе склад, нет -> пропускаем итерацию
		warehouseData, ok := dataByWarehouse[whid]
		if !ok || len(warehouseData) == 0 {
			continue
		}

		// если есть такой склад, формируем сообщение
		for _, v := range warehouseData {

			// пропуск данных, если это не монопаллета
			if appConfig.Monos != v.BoxTypeID {
				continue
			}

			// пропуск неподходящих коэффициентов
			if float32(monoCoef) <= v.Coefficient {
				continue
			}

			// пропуск дополнительной настройки
			if !v.AllowUnload {
				continue
			}

			// формирование сообщения
			message.WriteString("<b><u>x")
			fmt.Fprintf(&message, "%.0f", v.Coefficient)
			message.WriteString("</u></b> ")
			// если это СЦ, то добавляем к строке СЦ
			if v.IsSortingCenter {
				message.WriteString("СЦ ")
				fmt.Fprintf(&message, "%s ", v.WarehouseName)
			} else {
				fmt.Fprintf(&message, "%s ", v.WarehouseName)
			}
			message.WriteString("<b><u>")
			fmt.Fprintf(&message, "%s", mapDate[v.Date])
			message.WriteString("</u></b>")
			message.WriteString(" Моно\n")
			// добавляем информационное сообщение с тарифами на логистику и хранение
			infoMessage = fmt.Sprintf("\n  логистика: %sр. (+%s)\n  хранение: %sр.\n\n",
				v.DeliveryBaseLiter, v.DeliveryAdditionalLiter, v.StorageBaseLiter)
		}

		// по окончанию цикла, если не было данных то не приклеиваем информационное сообщение
		// сделано для склейки сообщений по одному типу (Моно или Короб)
		if previousLen != message.Len() {
			message.WriteString(infoMessage)
			previousLen = message.Len()
		}
	}

	// если сообщение не пустое, то отправляем клиенту в чат
	if message.String() != "" {
		if client == "public" {
			// отправляем в public-чат
			sendPrepareMessage(message.String(), client, "Mono", 1001)
		} else {
			// отправляем клиенту
			sendPrepareMessage(message.String(), client, "Mono", 1001)
		}
		// обнуляем для использованию в коробах
		message.Reset()
		previousLen = 0
	}

	// работаем с типом поставки Короб
	for whID, boxCoef := range appConfig.Clients[client].BoxData {

		// взяли данные склада из мапы, нет склада -> пропуск итерации
		warehouseData, ok := dataByWarehouse[whID]
		if !ok || len(warehouseData) == 0 {
			continue
		}

		// если есть данные, итерация по данным
		for _, v := range warehouseData {

			// пропуск неподходящего склада
			if appConfig.Boxes != v.BoxTypeID {
				continue
			}
			// пропуск неподходящего коэффициента
			if float32(boxCoef) <= v.Coefficient {
				continue
			}
			// пропуск, если не совпадает параметр
			if !v.AllowUnload {
				continue
			}

			// формирование сообщения
			message.WriteString("<b><u>x")
			fmt.Fprintf(&message, "%.0f", v.Coefficient)
			message.WriteString("</u></b> ")
			// если это СЦ, то добавляем приставку
			if v.IsSortingCenter {
				message.WriteString("СЦ ")
				fmt.Fprintf(&message, "%s ", v.WarehouseName)
			} else {
				fmt.Fprintf(&message, "%s ", v.WarehouseName)
			}
			message.WriteString("<b><u>")
			fmt.Fprintf(&message, "%s", mapDate[v.Date])
			message.WriteString("</u></b>")
			message.WriteString(" Короб\n")

			infoMessage = fmt.Sprintf("\n  логистика: %sр. (+%s)\n  хранение: %sр. (+%s)\n\n",
				v.DeliveryBaseLiter, v.DeliveryAdditionalLiter, v.StorageBaseLiter, v.StorageAdditionalLiter)
		}

		if previousLen != message.Len() {
			message.WriteString(infoMessage)
			message.WriteString("\n")
			previousLen = message.Len()
		}

		if message.String() != "" && client == "public" {
			sendPrepareMessage(message.String(), client, "Box", whID)
			message.Reset()
			previousLen = 0
		}
	}

	if message.String() != "" && client != "public" {
		sendPrepareMessage(message.String(), client, "Box", 100)
		message.Reset()
		previousLen = 0
	}

	return nil
}

// Функция, которая принимает слайс и возращает мапу с ключом ID склада.
func groupByWarehouse(data []Response) map[int][]Response {

	var group = make(map[int][]Response, len(data))

	for _, item := range data {
		group[item.WarehouseID] = append(group[item.WarehouseID], item)
	}

	return group
}

// Функция для отправки подготовленного сообщения
func sendPrepareMessage(message, client, boxType string, whid int) {

	// логирование в режиме отладки
	if appConfig.DebugMode {
		logging("отправка %s сообщения клиенту: %s по складу: %d", boxType, client, whid)
	}

	var id int = whid

	// замена входных данных для public Моно-сообщений
	if client == "public" && boxType == "Mono" {
		id = 1001
	}

	// отправка сообщения в телеграм-чат
	if err := sendTextMessage(message, client, id); err != nil {
		msg := fmt.Sprintf("ошибка отправки %s-сообщения: %v", boxType, err)
		// отправка сообщения с ошибкой Админу в чат
		if err := sendTextMessage(msg, appConfig.Admin, 0); err != nil {
			logging("%s не удалось отправить сообщение [%s] админу, ошибка: %v",
				EmojiWarning, msg, err)
		}
		// логирование ошибки
		logging("%s %s", EmojiWarning, msg)
	}
}
