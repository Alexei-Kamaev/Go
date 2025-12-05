package main

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisConfig *RedisConfig
	appConfig   *AppConfig
	apiTokenWB  string
	httpClient  = &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DisableCompression:    true,
			ResponseHeaderTimeout: 2 * time.Second,
			TLSHandshakeTimeout:   1 * time.Second,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			MaxIdleConnsPerHost:   5,
			ExpectContinueTimeout: 500 * time.Millisecond,
		},
	}
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TimeOut  time.Duration
}

type AppConfig struct {
	DebugMode       bool                  `json:"debug_mode"`
	Working         bool                  `json:"working"`
	RetryCodes      map[int]bool          `json:"retry_codes"`
	AllWarehouses   []string              `json:"all_warehouses"`
	CountRequests   int                   `json:"count_requests"`
	PauseRequests   int                   `json:"pause_requests"`
	RedisExpiration int                   `json:"redis_expiration"`
	Boxes           int                   `json:"boxes"`
	Monos           int                   `json:"monos"`
	Clients         map[string]ClientData `json:"clients"`
	URL             map[string]string     `json:"url"`
	Token           string                `json:"token"`
}

type ClientData struct {
	IsActive bool              `json:"is_active"`
	Pause    int               `json:"pause,omitempty"`
	BoxData  map[int]int       `json:"box_data"`
	MonoData map[int]int       `json:"mono_data"`
	ChatData map[string]string `json:"chat_data"`
	ApiData  map[string]string `json:"api_data"`
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
