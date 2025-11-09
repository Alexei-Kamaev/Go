package main

type configApp struct {
	WarehouseList string            `json:"warehouseList"`
	URL           map[string]string `json:"url"`
}
