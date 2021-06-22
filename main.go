package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"
)

// g () метод переадресации на сервис 1С под пользователем
// на основании токена из кук и регистрации в cache.json
func g(w http.ResponseWriter, r *http.Request) {

	// проверка на токен в куках
	username, err := r.Cookie("_token")
	if err != nil {
		logger.Printf("Error read cookie session token %v", err)
		http.Redirect(w, r, "/", 302)
		return
	}

	// проверяем токен
	var token, ok = "", false

	if token, ok = tokens.Get(username.Value); !ok {
		logger.Printf("Token for user %s not found ", username.Value)
		http.Redirect(w, r, "/", 302)
		return
	}

	// параметры в запросе
	params := mux.Vars(r)

	// переменные для формирования урла
	var url, p string
	// парсим адрес
	for _, v := range params {
		url += "/" + v
	}
	// собираем параметры запроса
	if r.URL.RawQuery != "" {
		p = "?" + r.URL.RawQuery
	}

	// делаем запрос на страницу под пользователем
	var n *Note
	if n, ok = cache.Get(token); !ok {
		logger.Printf("Cache data for user %s not found ", username.Value)
		http.Redirect(w, r, "/", 302)
		return
	}

	// вызов api из 1С
	h, b, err := n.conn.makeRequest(url+p, r.Method, r.Body, &r.Header)
	if err != nil {
		logger.Printf("Error make request %v", err)
	}

	// подмена кодировки в заголовках
	w.Header().Add("Content-Type", strings.ReplaceAll(h.Get("Content-Type"), "; charset=windows-1251", "; charset=UTF-8"))

	// logger.Printf("url %s \n w headers %+v \n h headers %+v \n method", r.URL, w.Header(), h, r.Method)
	// пишем в ответ то, что получилось
	w.Write(b)
}

// master () метод переадресации на сервис 1С под пользователем
// на основании токена из кук и регистрации в cache.json
func master(w http.ResponseWriter, r *http.Request) {

	// проверка на токен в куках
	username, err := r.Cookie("_token")
	if err != nil {
		logger.Printf("Error read cookie session token %v", err)
		http.Redirect(w, r, "/", 302)
		return
	}

	// проверяем токен
	var token, ok = "", false

	if token, ok = tokens.Get(username.Value); !ok {
		logger.Printf("Token for user %s not found ", username.Value)
		http.Redirect(w, r, "/", 302)
		return
	}

	// параметры в запросе
	params := mux.Vars(r)

	// переменные для формирования урла
	var url, p string
	// парсим адрес
	for _, v := range params {
		url += "/" + v
	}
	// собираем параметры запроса
	if r.URL.RawQuery != "" {
		p = "?" + r.URL.RawQuery
	}

	// делаем запрос на страницу под пользователем
	var n *Note
	if n, ok = cache.Get(token); !ok {
		logger.Printf("Cache data for user %s not found ", username.Value)
		http.Redirect(w, r, "/", 302)
		return
	}

	// вызов api из 1С
	h, b, err := n.conn.makeRequest(url+p, r.Method, r.Body, &r.Header)
	if err != nil {
		logger.Printf("Error make request %v", err)
	}

	// подмена кодировки в заголовках
	w.Header().Add("Content-Type", strings.ReplaceAll(h.Get("Content-Type"), "; charset=windows-1251", "; charset=UTF-8"))

	// logger.Printf("url %s \n w headers %+v \n h headers %+v \n method", r.URL, w.Header(), h, r.Method)
	// пишем в ответ то, что получилось
	w.Write(b)
}

// index () стартовая страница
func index(w http.ResponseWriter, r *http.Request) {

	// заголовки
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	// почитаем куки, может чего есть
	login := ""
	username, err := r.Cookie("_username")
	if err == nil {
		login = username.Value
	}

	// попытка возобновления сессии
	tokenCookie, err := r.Cookie("_token")
	if err == nil {
		// проверяем токен
		if _, ok := tokens.Get(tokenCookie.Value); ok {
			http.Redirect(w, r, "/g/MainPage", 302)
			return
		}
	}

	// если мы в локальной сети, то отображаем кнопку регистрации
	// если нет, то скрываем
	data := map[string]interface{}{
		"IsLocal": strings.HasPrefix(r.RemoteAddr, "127.0.0.1") || strings.HasPrefix(r.RemoteAddr, "192.168."),
		"Login":   login,
	}

	// парсим html
	t, err := template.ParseFiles("static/index.html")
	if err != nil {
		logger.Printf("Error parse static/index.html - %v ", err)
	}

	// выполняем...
	err = t.Execute(w, data)
	if err != nil {
		logger.Printf("Error execute static/index.html - %v ", err)
	}
}

// login () проверка данных при входе
func login(w http.ResponseWriter, r *http.Request) {

	//Обрабатываем только POST-запрос
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
	}

	err := r.ParseForm()
	if err != nil {
		logger.Printf("Error ParseForm() - %v", err)
	}

	username := r.FormValue("user")
	passcode := r.FormValue("passcode")

	// пробуем получить логин
	var n, found = &Note{}, false
	if n, found = cache.Get(username); !found {

		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		MessagePage(w, "Неверный логин!", "/", "На главную")

		return
	}

	// проверяем валидацию токена
	if valid := totp.Validate(passcode, n.Secret); !valid {

		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		MessagePage(w, "Неверный пароль!", "/", "На главную")

		return
	}

	// пробуем авторизоваться
	login, password := n.EncodeBasicAuth()
	n.conn = newHTTPService(config.Service1C, login, password)
	// создаем токен на сессию
	tokens.Create(w, username)
	// сохраним токены
	tokens.Save()

	// сохраним на целый месяц, что не забывали свои л
	SetCookie(w, "_username", username, 30*24*time.Hour)

	// перенаправляем
	http.Redirect(w, r, "/g/MainPage", 302)
}

// exit () разлогин
func exit(w http.ResponseWriter, r *http.Request) {

	username, err := r.Cookie("_token")
	if err != nil {
		logger.Print(err)
		http.Redirect(w, r, "/", 302)
		return
	}
	// стираем токены
	tokens.Delete(w, username.Value)
	// стираем токены и тут
	// http.SetCookie(w, &http.Cookie{Name: "_token", Value: "", Path: "/", Domain: "/"})
	// отправляем обратно на главную
	http.Redirect(w, r, "/", 302)
}

// static() ручка чтения картинок из /static/icons
func static(w http.ResponseWriter, r *http.Request) {
	// только png картинки
	w.Header().Set("Content-Type", "image/png; charset=utf-8")

	// пробуем прочитать
	file, err := ioutil.ReadFile(r.URL.Path[1:])
	if err != nil {
		logger.Printf("Error parse /static/ %v", err)
		return
	}

	// если получилось - вернем картинку
	w.Write(file)
}

// index () стартовая страница
func admin(w http.ResponseWriter, r *http.Request) {

	// заголовки
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	// // почитаем куки, может чего есть
	// login := ""
	// username, err := r.Cookie("_username")
	// if err == nil {
	// 	login = username.Value
	// }

	// // если мы в локальной сети, то отображаем кнопку регистрации
	// // если нет, то скрываем
	// data := map[string]interface{}{
	// 	"IsLocal": strings.HasPrefix(r.RemoteAddr, "127.0.0.1") || strings.HasPrefix(r.RemoteAddr, "192.168."),
	// 	"Login":   login,
	// }

	// парсим html
	t, err := template.ParseFiles("static/admin.html")
	if err != nil {
		logger.Printf("Error parse static/admin.html - %v ", err)
	}

	// выполняем...
	err = t.Execute(w, map[string]interface{}{
		"Users": cache.Table(),
	})
	if err != nil {
		logger.Printf("Error execute static/admin.html - %v ", err)
	}
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func main() {

	logger.Printf("Listen and serve on port %s", config.Port)

	router := mux.NewRouter()

	// router.HandleFunc("/admin", admin).Methods("GET")

	// ручка парсера картинок
	router.HandleFunc("/static/"+`{path:\S+}`, static)

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
	router.HandleFunc("/exit/", exit)

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
	router.HandleFunc("/master/{u1}", master).Methods("GET")
	router.HandleFunc("/master/{u1}/{u2}", master).Methods("GET")
	router.HandleFunc("/master/{u1}/{u2}/{u3}", master).Methods("GET")
	router.HandleFunc("/master/{u1}/{u2}/{u3}/{u4}", master).Methods("GET")
	router.HandleFunc("/master/{u1}/{u2}/{u3}/{u4}/{u5}", master).Methods("GET")

	// ручки переадресации по POST
	router.HandleFunc("/master/{u1}", master).Methods("POST")
	router.HandleFunc("/master/{u1}/{u2}", master).Methods("POST")
	router.HandleFunc("/master/{u1}/{u2}/{u3}", master).Methods("POST")
	router.HandleFunc("/master/{u1}/{u2}/{u3}/{u4}", master).Methods("POST")
	router.HandleFunc("/master/{u1}/{u2}/{u3}/{u4}/{u5}", master).Methods("POST")

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
	views = newHTTPService(config.Service1C, config.Login1C, config.Password1C)

	// открываем соединения
	tokens.OpenConnects()
}
