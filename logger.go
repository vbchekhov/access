package main

import (
	"log"
	"os"
)

// logger вывод логирования
var logger *log.Logger

// NewLogger () создание нового логирования
func NewLogger() *log.Logger {
	// если дебажим или допиливаем
	if config.Debug {
		return log.New(os.Stdout, "", 0)
	}

	// пробуем открыть
	if f, err := os.Open(config.Log); err == nil {
		return log.New(f, "", 0)
	}

	// пробуем создать
	f, err := os.Create(config.Log)
	if err != nil {
		log.Print("Error create logfile")
		return log.New(os.Stdout, "", 0)
	}

	return log.New(f, "", 0)
}
