package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func prepareMessages(data []Response, client string) error {

	if appConfig.DebugMode {
		logging("len data в функции [prepareMessages]: %d", len(data))
	}

	if _, exists := appConfig.Clients[client]; !exists {
		return fmt.Errorf("на клиента %s нет настроек в конфиге", client)
	}

	for whID, boxCoef := range appConfig.Clients[client].WhData {
		var message string
		for _, v := range data {
			if whID != v.WarehouseID {
				continue
			}
			clientData := strings.Split(boxCoef, "_")
			if len(clientData) < 2 {
				logging("неправильный формат boxCoef у клиента %s: %s", client, boxCoef)
				continue
			}
			box := clientData[0]
			coef := clientData[1]
			if box != fmt.Sprintf("%d", v.BoxTypeID) {
				continue
			}
			maxCoef, err := strconv.Atoi(coef)
			if err != nil {
				logging("ошибка при конвертации максимального коэффициента у клиента %s в функции [prepareMessages]: %v", client, err)
			}
			if float32(maxCoef) <= v.Coefficient {
				continue
			}
			message += fmt.Sprintf("<b><u>x%.0f</u></b> %s ", v.Coefficient, v.WarehouseName)
			normDate, err := time.Parse(time.RFC3339, v.Date)
			if err != nil {
				message += fmt.Sprintf("<b><u>%s</u></b>", v.Date)
			} else {
				message += fmt.Sprintf("<b><u>%s</u></b>", normDate.Format("02.01"))
			}
			message += " Короб\n"
			message += "     лог: " + v.DeliveryBaseLiter + "+" + v.DeliveryAdditionalLiter + " || "
			message += " хр: " + v.StorageBaseLiter + "+" + v.StorageAdditionalLiter + "\n"
		}
		if appConfig.DebugMode {
			logging("%s", message)
		}
		if message != "" {
			if appConfig.DebugMode {
				logging("отправка сообщения клиенту: %s по складу: %d", client, whID)
			}
			if err := sendTextMessage(message, client, whID); err != nil {
				logging("у клиента %s проблемы с отправкой сообщений: %v", client, err)
			}
		}
	}

	return nil
}
