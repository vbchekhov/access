package main

import (
	"embed"
	"flag"
	"github.com/gorilla/mux"
	"net/http"
)

//go:embed pages/embed
var pages embed.FS

//go:embed static
var static embed.FS

func main() {

	logger.Printf("Listen and serve on port %s", config.Port)

	router := mux.NewRouter()

	router.Use(auth)

	// ручка парсера картинок
	router.PathPrefix("/static/").Handler(http.FileServer(http.FS(static)))

	// стартовая страница и авторизация
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/login/", login).Methods("POST")

	// регистрация в сервисе
	router.HandleFunc("/first/", first).Methods("GET")
	router.HandleFunc("/qr.png", qrCode)
	router.HandleFunc("/2fa/", new2fa).Methods("POST")
	router.HandleFunc("/manual2fa/", manual2FA).Methods("GET")
	router.HandleFunc("/verify2fa/", verifi2fa).Methods("POST")

	// выход из сервиса
	router.HandleFunc("/exit/", exit).Methods("GET", "POST")

	// ручки переадресации по GET
	router.HandleFunc("/g/{u1}", g).Methods("GET")
	router.HandleFunc("/g/{u1}/{u2}", g).Methods("GET")
	router.HandleFunc("/g/{u1}/{u2}/{u3}", g).Methods("GET")
	router.HandleFunc("/g/{u1}/{u2}/{u3}/{u4}", g).Methods("GET")
	router.HandleFunc("/g/{u1}/{u2}/{u3}/{u4}/{u5}", g).Methods("GET")

	// ручки переадресации по POST
	router.HandleFunc("/g/{u1}", g).Methods("POST")
	router.HandleFunc("/g/{u1}/{u2}", g).Methods("POST")
	router.HandleFunc("/g/{u1}/{u2}/{u3}", g).Methods("POST")
	router.HandleFunc("/g/{u1}/{u2}/{u3}/{u4}", g).Methods("POST")
	router.HandleFunc("/g/{u1}/{u2}/{u3}/{u4}/{u5}", g).Methods("POST")

	// ручки переадресации по GET
	router.HandleFunc("/master/{master_key}/{u1}", master).Methods("GET")
	router.HandleFunc("/master/{master_key}/{u1}/{u2}", master).Methods("GET")
	router.HandleFunc("/master/{master_key}/{u1}/{u2}/{u3}", master).Methods("GET")
	router.HandleFunc("/master/{master_key}/{u1}/{u2}/{u3}/{u4}", master).Methods("GET")
	router.HandleFunc("/master/{master_key}/{u1}/{u2}/{u3}/{u4}/{u5}", master).Methods("GET")

	// ручки переадресации по POST
	router.HandleFunc("/master/{master_key}/{u1}", master).Methods("POST")
	router.HandleFunc("/master/{master_key}/{u1}/{u2}", master).Methods("POST")
	router.HandleFunc("/master/{master_key}/{u1}/{u2}/{u3}", master).Methods("POST")
	router.HandleFunc("/master/{master_key}/{u1}/{u2}/{u3}/{u4}", master).Methods("POST")
	router.HandleFunc("/master/{master_key}/{u1}/{u2}/{u3}/{u4}/{u5}", master).Methods("POST")

	http.ListenAndServe(":"+config.Port, router)
}

func init() {

	configPath = "./config.json"
	// парсим флаги
	flag.StringVar(&configPath, "conf", "./config.json", "a config path")
	flag.Parse()

	// читаем конфигурацию
	config = NewConfig(configPath)

	// загружаем глобальные переменные
	logger = NewLogger()
	cache = NewCache().Read()
	tokens = NewTokens().Read()
	views = newHttpService(config.Service1C, config.Login1C, config.Password1C)

	// открываем соединения
	tokens.OpenConnects()
}
