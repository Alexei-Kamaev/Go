package main

import (
	"fmt"
	"strings"
	"time"
)

// Функция, которая:
// принимает очищенные данные,
// формирует готовое сообщение для каждого клиента
// и отправляет каждому клиенту в чат.
func prepareMessages(data []Response, client string) error {

	if _, exists := appConfig.Clients[client]; !exists {
		return fmt.Errorf("на клиента %s нет настроек в конфиге", client)
	}

	if len(appConfig.Clients[client].BoxData)+len(appConfig.Clients[client].MonoData) == 0 {
		logging("у клиента [%s] нет складов в конфигурации для отправки сообщений", client)
		return nil
	}

	dataByWarehouse := groupByWarehouse(data)

	for whid, monoCoef := range appConfig.Clients[client].MonoData {

		var (
			message     strings.Builder
			infoMessage string
		)

		warehouseData, ok := dataByWarehouse[whid]

		if !ok || len(warehouseData) == 0 {
			continue
		}

		for _, v := range warehouseData {

			if appConfig.Monos != v.BoxTypeID {
				continue
			}

			if float32(monoCoef) <= v.Coefficient {
				continue
			}

			if !v.AllowUnload {
				continue
			}

			fmt.Fprintf(&message, "<b><u>x%.0f</u></b> ", v.Coefficient)

			if v.IsSortingCenter {
				fmt.Fprintf(&message, "СЦ %s ", v.WarehouseName)
			} else {
				fmt.Fprintf(&message, "%s ", v.WarehouseName)
			}

			normDate, err := time.Parse(time.RFC3339, v.Date)

			if err != nil {
				fmt.Fprintf(&message, "<b><u>%s</u></b>", v.Date)
			} else {
				fmt.Fprintf(&message, "<b><u>%s</u></b>", normDate.Format("02.01"))
			}

			message.WriteString(" Моно\n")

			infoMessage = fmt.Sprintf("\n  логистика: %sр. (+%s)\n  хранение: %sр.\n",
				v.DeliveryBaseLiter, v.DeliveryAdditionalLiter, v.StorageBaseLiter)
		}

		if message.String() != "" {

			message.WriteString(infoMessage)

			if appConfig.DebugMode {
				logging("отправка Mono сообщения клиенту: %s по складу: %d", client, whid)
			}

			if err := sendTextMessage(message.String(), client, 1001); err != nil {

				msg := fmt.Sprintf("ошибка при отправке Mono-сообщения: %v", err)

				if err := sendTextMessage(msg, appConfig.Admin, 0); err != nil {
					logging("не удалось отправить сообщение [%s] админу, ошибка: %v", msg, err)
				}

				logging("%s", msg)
			}
		}
	}

	for whID, boxCoef := range appConfig.Clients[client].BoxData {

		var (
			message     strings.Builder
			infoMessage string
		)

		warehouseData, ok := dataByWarehouse[whID]

		if !ok || len(warehouseData) == 0 {
			continue
		}

		for _, v := range warehouseData {

			if appConfig.Boxes != v.BoxTypeID {
				continue
			}

			if float32(boxCoef) <= v.Coefficient {
				continue
			}

			if !v.AllowUnload {
				continue
			}

			fmt.Fprintf(&message, "<b><u>x%.0f</u></b> ", v.Coefficient)

			if v.IsSortingCenter {
				fmt.Fprintf(&message, "СЦ %s ", v.WarehouseName)
			} else {
				fmt.Fprintf(&message, "%s ", v.WarehouseName)
			}

			normDate, err := time.Parse(time.RFC3339, v.Date)

			if err != nil {
				fmt.Fprintf(&message, "<b><u>%s</u></b>", v.Date)
			} else {
				fmt.Fprintf(&message, "<b><u>%s</u></b>", normDate.Format("02.01"))
			}

			message.WriteString(" Короб\n")

			infoMessage = fmt.Sprintf("\n  логистика: %sр. (+%s)\n  хранение: %sр. (+%s)\n",
				v.DeliveryBaseLiter, v.DeliveryAdditionalLiter, v.StorageBaseLiter, v.StorageAdditionalLiter)
		}

		if message.String() != "" {

			message.WriteString(infoMessage)

			if appConfig.DebugMode {
				logging("отправка Box сообщения клиенту: %s по складу: %d", client, whID)
			}

			if err := sendTextMessage(message.String(), client, whID); err != nil {

				msg := fmt.Sprintf("ошибка при отправке Box-сообщения: %v", err)

				if err := sendTextMessage(msg, appConfig.Admin, 0); err != nil {
					logging("не удалось отправить сообщение [%s] админу, ошибка: %v", msg, err)
				}

				logging("%s", msg)
			}
		}
	}

	return nil
}

// Функция, которая принимает слайс и возращает мапу с ключом ID склада.
func groupByWarehouse(data []Response) map[int][]Response {
	group := make(map[int][]Response)
	for _, item := range data {
		group[item.WarehouseID] = append(group[item.WarehouseID], item)
	}

	return group
}
