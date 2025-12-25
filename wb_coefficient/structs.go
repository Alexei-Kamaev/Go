package main

import (
	"time"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TimeOut  time.Duration
}

type AppConfig struct {
	Admin                string `json:"admin"`
	BotToken             string `json:"bot_token"`
	DebugMode            bool   `json:"debug_mode"`
	Working              bool   `json:"working"`
	AllActiveClients     []string
	AllCountSendMessages int
	RetryCodes           map[int]bool          `json:"retry_codes"`
	AllWarehouses        []string              `json:"all_warehouses,omitempty"`
	PauseIteration       int                   `json:"pause"`
	RedisExpiration      int                   `json:"redis_expiration"`
	Boxes                int                   `json:"boxes"`
	Monos                int                   `json:"monos"`
	Clients              map[string]ClientData `json:"clients"`
	URL                  map[string]string     `json:"url"`
	Token                string                `json:"token"`
}

type ClientData struct {
	IsActive  bool              `json:"is_active"`
	Pause     int               `json:"pause,omitempty"`
	PauseWHID map[int]bool      `json:"pause_whid,omitempty"`
	BoxData   map[int]int       `json:"box_data"`
	MonoData  map[int]int       `json:"mono_data"`
	ChatData  map[string]string `json:"chat_data"`
	ApiData   map[string]string `json:"api_data"`
	TGToken   string            `json:"tg_token"`
}

type Response struct {
	Date string `json:"date"` // string
	// Дата начала действия коэффициента
	Coefficient float32 `json:"coefficient"` // number
	// Коэффициент приёмки:
	// -1 — приёмка недоступна, вне зависимости от значения поля allowUnload
	// 0 — бесплатная приёмка
	// от 1 — множитель стоимости приёмки
	WarehouseID int `json:"warehouseID"` // integer
	// ID склада
	WarehouseName string `json:"warehouseName"` // string
	// Название склада
	AllowUnload bool `json:"allowUnload"` // boolean
	// Доступность приёмки для поставок данного типа, смотри значение поля boxTypeID:
	// true — приёмка доступна
	// false — приёмка не доступна
	BoxTypeID int `json:"boxTypeID"` // integer
	// ID типа поставки:
	// 2 — Короба
	// 5 — Монопаллеты
	// 6 — Суперсейф
	// Для типа поставки QR-поставка с коробами поле не возвращается
	StorageCoef string `json:"storageCoef"` // string or null
	// Коэффициент хранения
	DeliveryCoef string `json:"deliveryCoef"` // string or null
	// Коэффициент логистики
	DeliveryBaseLiter string `json:"deliveryBaseLiter"` // string or null
	// Стоимость логистики первого литра
	DeliveryAdditionalLiter string `json:"deliveryAdditionalLiter"` //string or null
	// Стоимость логистики каждого следующего литра
	StorageBaseLiter string `json:"storageBaseLiter"` // string or null
	// Стоимость хранения:
	// для паллет — стоимость за одну паллету
	// для коробов — стоимость хранения за первый литр
	StorageAdditionalLiter string `json:"storageAdditionalLiter"` // string or null
	// Стоимость хранения каждого последующего литра:
	// для паллет — всегда будет null, т.к. стоимость хранения за единицу паллеты определяется в StorageBaseLiter
	// для коробов — стоимость хранения за каждый последующий литр
	IsSortingCenter bool `json:"isSortingCenter"` // boolean
	// Тип склада:
	// true — сортировочный центр (СЦ)
	// false — обычный
}

type WarehouseListID struct {
	ID int64 `json:"ID"` // integer
	// ID склада
	Name string `json:"name"` // string
	// Название склада
	Address string `json:"address"` // string
	// Адрес склада
	WorkTime string `json:"workTime"` // string
	// Режим работы склада
	IsActive bool `json:"isActive"` // boolean
	// Доступен ли в качестве склада назначения:
	// true — да
	// false — нет
	IsTransitActive bool `json:"isTransitActive"` // boolean
	// Доступен ли в качестве транзитного склада:
	// true — да
	// false — нет
}

type TelegramResponse struct {
	Status      bool   `json:"ok"`
	Description string `json:"description,omitempty"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Parameters  struct {
		RetryAfter int `json:"retry_after,omitempty"`
	} `json:"parameters,omitempty"`
	Result struct {
		MessageID int    `json:"message_id"`
		Text      string `json:"text"`
		Date      int    `json:"date"`
		Chat      struct {
			ID       int64  `json:"id"`
			Title    string `json:"title"`
			Username string `json:"username"`
		} `json:"chat"`
	} `json:"result,omitempty"`
}
