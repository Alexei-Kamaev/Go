package main

import (
	"sync"

	"github.com/redis/go-redis/v9"
)

const (
	configFile = "config.json"
	appName    = "coefbot"
)

var (
	logMutex        sync.Mutex
	logs            string
	cfgApp          *configApp
	redisClient     *redis.Client
	cfgRedis        *redisConfig
	redisExpiration = 24
)

type configApp struct {
	WarehouseList []int             `json:"warehouseList"`
	URL           map[string]string `json:"url"`
}
