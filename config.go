package main

import (
	"encoding/json"
	"io/ioutil"
)

var config *Config
var configPath string

// Config кофигурация сервиса
type Config struct {
	// параметры подключения к 1С
	Service1C  string `json:"service_1С"`
	Login1C    string `json:"login_1С"`
	Password1C string `json:"password_1С"`
	// параметры приложения
	Debug    bool   `json:"debug"`
	Port     string `json:"port"`
	// хранилища токенов и данных
	Sessions string `json:"sessions"`
	Cache string `json:"cache"`
	Log string `json:"log"`
}

// NewConfig () чтение файла конфигурации
func NewConfig(path string) (c *Config) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Printf("Error ReadFile() config file - %v", err)
		return
	}

	err = json.Unmarshal(f, &c)
	if err != nil {
		logger.Printf("Error Unmarshal() config file - %v", err)
		return
	}

	return c
}
