package main

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	auth "gopkg.in/korylprince/go-ad-auth.v3"
)

// first () регистрация в нашем сервисе
func first(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	// парсим html
	t, err := template.ParseFiles("static/first.html")
	if err != nil {
		logger.Printf("Error parse static/first.html - %v ", err)
	}

	// выполняем...
	err = t.Execute(w, nil)
	if err != nil {
		logger.Printf("Error execute static/index.html - %v ", err)
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
	config := &auth.Config{
		Server:   "SAD.vmpauto.local",
		Port:     389,
		BaseDN:   "OU=Personal,OU=LOCAL,DC=vmpauto,DC=local",
		Security: auth.SecurityNone,
	}

	// собираем значения
	// сделаем доменной логин маленькими буквами
	username := strings.ToLower(r.FormValue("user"))
	password := r.FormValue("password")

	status, err := auth.Authenticate(config, username, password)

	if !status || err != nil {
		// если ошибка
		if err != nil {
			logger.Printf("Error Authenticate() in domain - %v", err)
			return
		}

		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		MessagePage(w, "Неверный логин или пароль! Повторите настройку сначала!", "/", "На главную")

		return
	}

	//  поставим куки на нужные нам странички
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/qr.png"})
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/2fa/"})
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/verify2fa/"})
	http.SetCookie(w, &http.Cookie{Name: "_username", Value: username, Path: "/manual2fa/"})

	// парсим html
	t, err := template.ParseFiles("static/2fa.html")
	if err != nil {
		logger.Printf("Error parse static/2fa.html - %v ", err)
	}

	// выполняем...
	err = t.Execute(w, nil)
	if err != nil {
		logger.Printf("Error execute static/2fa.html - %v ", err)
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
		t, err := template.ParseFiles("static/manual2fa.html")
		if err != nil {
			logger.Printf("Error parse static/2fa.html - %v ", err)
		}

		// выполняем...
		err = t.Execute(w, data)
		if err != nil {
			logger.Printf("Error execute static/2fa.html - %v ", err)
		}

		return
	}

	// пишем заголовки
	w.WriteHeader(http.StatusForbidden)
	// выдаем страничку ошибки
	MessagePage(w, "Внутренняя ошибка! Обратитесь в IT отдел!", "/", "На главную")
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
		MessagePage(w, "Неверный код! Повторите настройку сначала!", "/", "На главную")

		return
	}

	if _, err := n.NewBasicAuth(); err != nil {
		// пишем заголовки
		w.WriteHeader(http.StatusForbidden)
		// выдаем страничку ошибки
		MessagePage(w, "Внутренняя ошибка! Обратитесь в IT отдел!", "/", "На главную")

		return
	}

	// запишем данные на случай падения
	cache.Save()

	// сохраним на целый месяц, что не забывали свои л
	SetCookie(w, "_username", username.Value, 30*24*time.Hour)

	// если все ок
	MessagePage(w, "Настройка прошла успешно!", "/", "На главную")
}
