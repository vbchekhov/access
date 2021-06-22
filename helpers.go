package main

import (
	"html/template"
	"net/http"
	"time"
)

// MessagePage () страничка вывода простых сообщений
func MessagePage(w http.ResponseWriter, message, href, button string) {

	// заголовки
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// заполним
	data := map[string]string{
		"Message": message,
		"Href":    href,
		"Button":  button,
	}
	// парсим html
	t, err := template.ParseFiles("static/message.html")
	if err != nil {
		logger.Printf("Error parse static/message.html - %v ", err)
	}
	// выполняем
	err = t.Execute(w, data)
	if err != nil {
		logger.Printf("Error execute static/message.html - %v ", err)
	}
}

// SetCookie () установка кук
func SetCookie(w http.ResponseWriter, name, value string, ttl time.Duration) error {

	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		Name:       name,
		Value:      value,
		Domain:     "/",
		Path:       "/",
		Expires:    expire,
		RawExpires: expire.Format(time.RFC3339),
	}
	http.SetCookie(w, &cookie)

	return nil
}
