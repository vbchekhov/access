package main

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
)

type notAuth struct {
	Mask *regexp.Regexp
}

var gHandler = regexp.MustCompile("^/g/*")
var masterHandler = regexp.MustCompile("^/master/*")

var notAuthPath = []notAuth{
	{Mask: regexp.MustCompile("^/$")},
	{Mask: regexp.MustCompile("/first/.*")},
	{Mask: regexp.MustCompile("/qr.png")},
	{Mask: regexp.MustCompile("/login/.*")},
	{Mask: regexp.MustCompile("/2fa/.*")},
	{Mask: regexp.MustCompile("/verify2fa/.*")},
	{Mask: regexp.MustCompile("/manual2fa/.*")},
	{Mask: regexp.MustCompile("/static/.*")},
}

func auth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestPath := r.URL.Path

		for i := range notAuthPath {

			search := notAuthPath[i].Mask.FindStringSubmatch(requestPath)

			if len(search) != 0 {
				next.ServeHTTP(w, r)
				return
			}
		}

		gMask := gHandler.FindStringSubmatch(requestPath)
		if len(gMask) != 0 {
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

			// делаем запрос на страницу под пользователем
			var n *Note
			if n, ok = cache.Get(token); !ok {
				logger.Printf("Cache data for user %s not found ", username.Value)
				http.Redirect(w, r, "/", 302)
				return
			}

			ctx := context.WithValue(r.Context(), "note", n)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}

		masterMask := masterHandler.FindStringSubmatch(requestPath)
		if len(masterMask) != 0 {

			vars := mux.Vars(r)
			key := vars["master_key"]

			// делаем запрос на страницу под пользователем
			var n *Note
			var ok bool
			if n, ok = cache.GetBySecret(key); !ok {
				logger.Printf("Cache data for user %s not found ", key)
				http.Redirect(w, r, "/", 302)
				return
			}

			ctx := context.WithValue(r.Context(), "note", n)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}

	})
}
