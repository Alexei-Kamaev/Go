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
	// получаем длинну среза, который пришёл в функцию в режиме отладки
	if appConfig.DebugMode {
		logging("len(данных которые пришли) в функции [prepareMessages]: %d", len(data))
	}
	// ищем настройки клиента в конфигурации
	if _, exists := appConfig.Clients[client]; !exists {
		return fmt.Errorf("на клиента %s нет настроек в конфиге", client)
	}
	// обработка сообщений для типа поставки Моно
	for whid, monoCoef := range appConfig.Clients[client].MonoData {
		var (
			message     strings.Builder
			infoMessage string
		)
		// итерация по данным которые пришли в функцию
		for _, v := range data {
			// пропуск неподходящего склада
			if whid != v.WarehouseID {
				continue
			}
			// пропуск если не монопаллета
			if appConfig.Monos != v.BoxTypeID {
				continue
			}
			// пропуск если не устраивает коэффициент
			if float32(monoCoef) <= v.Coefficient {
				continue
			}
			// пропуск сочетания
			if !v.AllowUnload {
				continue
			}
			// неподсредственно склейка сообшения
			message.WriteString(fmt.Sprintf("<b><u>x%.0f</u></b> %s ", v.Coefficient, v.WarehouseName))
			normDate, err := time.Parse(time.RFC3339, v.Date)
			if err != nil {
				message.WriteString(fmt.Sprintf("<b><u>%s</u></b>", v.Date))
			} else {
				message.WriteString(fmt.Sprintf("<b><u>%s</u></b>", normDate.Format("02.01")))
			}
			message.WriteString(" Моно\n")
			infoMessage = fmt.Sprintf("\nДля информации:\n  логистика: %sр. (+%s)\n  хранение: %sр.\n", v.DeliveryBaseLiter, v.DeliveryAdditionalLiter, v.StorageBaseLiter)
		}
		// отправка непустого сообщения в чат клиенту
		if message.String() != "" {
			message.WriteString(infoMessage)
			if appConfig.DebugMode {
				logging("отправка Mono сообщения клиенту: %s по складу: %d", client, whid)
			}
			if err := sendTextMessage(message.String(), client, 1001); err != nil {
				logging("у клиента %s проблемы с отправкой Mono сообщения: %v", client, err)
			}
		}
		// чтение сообщения в режиме отладки
		if appConfig.DebugMode {
			logging("Monos:\n%s", message.String())
		}
	}
	// обработка сообщений для типа поставки Короб
	for whID, boxCoef := range appConfig.Clients[client].BoxData {
		var (
			message     strings.Builder
			infoMessage string
		)
		// итерация по полученным данным
		for _, v := range data {
			// пропуск ненужного склада
			if whID != v.WarehouseID {
				continue
			}
			// пропуск если тип поставки не Короб
			if appConfig.Boxes != v.BoxTypeID {
				continue
			}
			// пропуск неподходящего коэффициента
			if float32(boxCoef) <= v.Coefficient {
				continue
			}
			// пропуск сочетания
			if !v.AllowUnload {
				continue
			}
			// непосредственно склейка сообщения
			message.WriteString(fmt.Sprintf("<b><u>x%.0f</u></b> %s ", v.Coefficient, v.WarehouseName))
			normDate, err := time.Parse(time.RFC3339, v.Date)
			if err != nil {
				message.WriteString(fmt.Sprintf("<b><u>%s</u></b>", v.Date))
			} else {
				message.WriteString(fmt.Sprintf("<b><u>%s</u></b>", normDate.Format("02.01")))
			}
			message.WriteString(" Короб\n")
			infoMessage = fmt.Sprintf("\nДля информации:\n  логистика: %sр. (+%s)\n  хранение: %sр. (+%s)\n", v.DeliveryBaseLiter, v.DeliveryAdditionalLiter, v.StorageBaseLiter, v.StorageAdditionalLiter)
		}
		// отправка сообщения клиенту в чат
		if message.String() != "" {
			message.WriteString(infoMessage)
			if appConfig.DebugMode {
				logging("отправка Box сообщения клиенту: %s по складу: %d", client, whID)
			}
			if err := sendTextMessage(message.String(), client, whID); err != nil {
				logging("у клиента %s проблемы с отправкой Box сообщения: %v", client, err)
			}
		}
		// чтение сообщения в режиме отладки
		if appConfig.DebugMode {
			logging("Boxes:\n%s", message.String())
		}
	}
	return nil
}
