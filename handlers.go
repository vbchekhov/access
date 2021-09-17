package main

import (
	"context"
	"github.com/gorilla/mux"
	authLD "github.com/korylprince/go-ad-auth/v3"
	"github.com/pquerna/otp/totp"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

/* Ручки для работы с rest api 1С
отвечают за переадресацию и обработку
страниц данным прокси. Данные берутся
по ручку из файла config.json
Параметр: service_1С
*/

// g () метод переадресации на сервис 1С под пользователем
// на основании токена из кук и регистрации в cache.json
func g(w http.ResponseWriter, r *http.Request) {

	n := r.Context().Value("note").(*Note)

	// параметры в запросе
	// params := mux.Vars(r)

	// переменные для формирования урла
	var url, p string
	// парсим адрес
	// for _, v := range params {
	// 	url += "/" + v
	// }
	url = strings.Replace(r.URL.Path, "/g", "", 1)

	// собираем параметры запроса
	if r.URL.RawQuery != "" {
		p = "?" + r.URL.RawQuery
	}

	// вызов api из 1С
	h, b, err := n.conn.makeRequest(url+p, r.Method, r.Body, &r.Header)
	if err != nil {
		logger.Printf("Error make request %v", err)
	}

	// подмена кодировки в заголовках
	w.Header().Add("Content-Type", strings.ReplaceAll(h.Get("Content-Type"), "; charset=windows-1251", "; charset=UTF-8"))

	// logger.Printf("URL %s | BODY %s", url+p, b)

	// logger.Printf("url %s \n w headers %+v \n h headers %+v \n method", r.URL, w.Header(), h, r.Method)
	// пишем в ответ то, что получилось
	w.Write(b)
}

// master () метод переадресации на сервис 1С под пользователем
// на основании secret key пользователя в cache.json
// Позваоляем не вводить логин и пароль каждый раз.
// TODO создать механизм генерации мастер ссылок для
// 	пользователей сервиса.
func master(w http.ResponseWriter, r *http.Request) {

	n := r.Context().Value("note").(*Note)

	// параметры в запросе
	params := mux.Vars(r)

	// переменные для формирования урла
	var url, p string
	// парсим адрес
	for k, v := range params {
		if strings.HasPrefix(k, "u") {
			url += "/" + v
		}
	}
	// собираем параметры запроса
	if r.URL.RawQuery != "" {
		p = "?" + r.URL.RawQuery
	}

	// вызов api из 1С
	h, b, err := n.conn.makeRequest(url+p, r.Method, r.Body, &r.Header)
	if err != nil {
		logger.Printf("Error make request %v", err)
	}

	// подмена кодировки в заголовках
	w.Header().Add("Content-Type", strings.ReplaceAll(h.Get("Content-Type"), "; charset=windows-1251", "; charset=UTF-8"))

	// пишем в ответ то, что получилось
	w.Write(b)
}

/* Ручки авторизации на сервисе
отвечат за формирование токенов и открытия сессий
для работы с rest api 1С. Авторизация проходит
по 2fa механизму, ну типа :)
*/

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
	t, err := template.ParseFS(pages, "pages/embed/index.html")
	if err != nil {
		logger.Printf("Error parse pages/embed/index.html - %v ", err)
	}

	// выполняем...
	err = t.Execute(w, data)
	if err != nil {
		logger.Printf("Error execute pages/embed/index.html - %v ", err)
	}
}

// login () проверка данных при входе
func login(w http.ResponseWriter, r *http.Request) {

	// Обрабатываем только POST-запрос
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
	}

	err := r.ParseForm()
	if err != nil {
		logger.Printf("Error ParseForm() - %v", err)
	}

	username := r.FormValue("user")
	passcode := r.FormValue("passcode")
	password := r.FormValue("password")

	// пробуем получить логин
	var n, found = &Note{}, false
	if n, found = cache.Get(username); !found {

		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		messagePage(w, "Неверный логин!", "/", "На главную")

		return
	}

	// проверяем валидацию токена
	if valid := totp.Validate(passcode, n.Secret); !valid && password == "" {

		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		messagePage(w, "Неверный пароль!", "/", "На главную")

		return
	}

	if password != "" && n.Password != password {

		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		messagePage(w, "Неверный пароль!", "/", "На главную")

		return
	}

	// пробуем авторизоваться
	login1C, password1C := n.EncodeBasicAuth()
	n.conn = newHttpService(config.Service1C, login1C, password1C)
	// создаем токен на сессию
	tokens.Create(w, username)
	// сохраним токены
	tokens.Save()

	// сохраним на целый месяц, что не забывали свои л
	SetCookie(w, "_username", username, 30*24*time.Hour)

	ctx := context.WithValue(r.Context(), "note", n)
	r = r.WithContext(ctx)

	// перенаправляем
	http.Redirect(w, r, "/g/MainPage", 302)
}

// first () регистрация в нашем сервисе
func first(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	// парсим html
	t, err := template.ParseFS(pages, "pages/embed/first.html")
	if err != nil {
		logger.Printf("Error parse pages/embed/first.html - %v ", err)
	}

	// выполняем...
	err = t.Execute(w, nil)
	if err != nil {
		logger.Printf("Error execute pages/embed/index.html - %v ", err)
	}
}

// new2fa () страничка авторизации в домене.
//  проверяем - наш ли это?
//  если наш, то рисуем страничку с qr кодом
func new2fa(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	// обрабатываем только POST-запрос
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
	}

	// парсим форму
	err := r.ParseForm()
	if err != nil {
		logger.Printf("Error ParseForm() - %v", err)
	}

	// настройки соединения с доменом
	config := &authLD.Config{
		Server:   "SAD.vmpauto.local",
		Port:     389,
		BaseDN:   "OU=Personal,OU=LOCAL,DC=vmpauto,DC=local",
		Security: authLD.SecurityNone,
	}

	// собираем значения
	// сделаем доменной логин маленькими буквами
	username := strings.ToLower(r.FormValue("user"))
	password := r.FormValue("password")

	status, err := authLD.Authenticate(config, username, password)

	if !status || err != nil {
		// если ошибка
		if err != nil {
			logger.Printf("Error Authenticate() in domain - %v", err)
			return
		}

		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		messagePage(w, "Неверный логин или пароль! Повторите настройку сначала!", "/", "На главную")

		return
	}

	//  поставим куки на нужные нам странички
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/qr.png"})
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/2fa/"})
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/verify2fa/"})
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/manual2fa/"})

	// парсим html
	t, err := template.ParseFS(pages, "pages/embed/2fa.html")
	if err != nil {
		logger.Printf("Error parse pages/embed/2fa.html - %v ", err)
	}

	// выполняем...
	err = t.Execute(w, nil)
	if err != nil {
		logger.Printf("Error execute pages/embed/2fa.html - %v ", err)
	}
}

// qrCode () выдача картинки
func qrCode(w http.ResponseWriter, r *http.Request) {

	// читаем куки
	username, err := r.Cookie("_username")
	if err != nil {
		logger.Printf("Error read cookie _username() - %v", err)
		// return
	}

	// создаем новую запись
	note := &Note{
		UserName: username.Value,
	}

	// собираем ключ
	if _, err := note.NewKey(); err != nil {
		logger.Printf("Error generate key() - %v", err)
		return
	}

	// собираем qr код
	qr, err := note.NewQRCode()
	if err != nil {
		logger.Printf("Error generate qrcode() - %v", err)
		return
	}

	// пишем кэш
	cache.Add(note)

	// отправляем qr
	w.Header().Set("Content-Type", "image/png")
	w.Write(qr)
}

// manual2FA () если вдруг нет компа под рукой,
//  чтобы отсканировать qr код
func manual2FA(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	// читаем куки
	username, err := r.Cookie("_username")
	if err != nil {
		logger.Printf("Error read cookie _username() - %v", err)
		// return
	}

	// пробуем получить логин
	if n, found := cache.Get(username.Value); found {

		// структура для отрисовки страницы
		data := map[string]string{
			"AccountName":   username.Value + "@vmpauto.local",
			"AccountSecret": n.Secret,
		}

		// парсим html
		t, err := template.ParseFS(pages, "pages/embed/manual2fa.html")
		if err != nil {
			logger.Printf("Error parse pages/embed/2fa.html - %v ", err)
		}

		// выполняем...
		err = t.Execute(w, data)
		if err != nil {
			logger.Printf("Error execute pages/embed/2fa.html - %v ", err)
		}

		return
	}

	// пишем заголовки
	w.WriteHeader(http.StatusForbidden)
	// выдаем страничку ошибки
	messagePage(w, "Внутренняя ошибка! Обратитесь в IT отдел!", "/", "На главную")
}

// verifi2fa () проверка кода на соответствие и запись
//  в нашу базу данных
func verifi2fa(w http.ResponseWriter, r *http.Request) {

	// обрабатываем только POST-запрос
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
	}

	// читаем куки
	username, err := r.Cookie("_username")
	if err != nil {
		logger.Printf("Error read cookie _username() - %v", err)
		// return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Printf("Error ParseForm() - %v", err)
	}

	// читаем кэщ
	n, _ := cache.Get(username.Value)

	// проверяем код
	passcode := r.FormValue("passcode")
	if valid := totp.Validate(passcode, n.Secret); !valid {

		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		messagePage(w, "Неверный код! Повторите настройку сначала!", "/", "На главную")

		return
	}

	if _, err := n.NewBasicAuth(); err != nil {
		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		messagePage(w, "Внутренняя ошибка! Обратитесь в IT отдел!", "/", "На главную")

		return
	}

	// запишем данные на случай падения
	cache.Save()

	// сохраним на целый месяц, что не забывали свои л
	SetCookie(w, "_username", username.Value, 30*24*time.Hour)

	// если все ок
	messagePage(w, "Настройка прошла успешно!", "/", "На главную")
}

// exit () разлогин
func exit(w http.ResponseWriter, r *http.Request) {

	log.Print("op!")

	username, err := r.Cookie("_token")
	log.Printf("%v %v", username, err)
	if err != nil {
		logger.Printf("Error exit %v", err)
		http.Redirect(w, r, "/", 302)
		return
	}
	// стираем токены
	tokens.Delete(w, username.Value)

	// отправляем обратно на главную
	http.Redirect(w, r, "/", 302)
}
