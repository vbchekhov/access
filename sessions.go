package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

var tokens *Tokens

// Tokens хранилище ключей
type Tokens struct {
	sync.Mutex
	Storage map[string]string
}

// NewTokens () новый кэш
func NewTokens() *Tokens {
	return &Tokens{Storage: make(map[string]string)}
}

// ExistFile () проверка существования файла
func (t *Tokens) ExistFile() bool {

	if _, err := os.Open(config.Sessions); os.IsNotExist(err) {
		return false
	}

	return true
}

// Create () добавляем токен в хранилище и в куки
func (t *Tokens) Create(w http.ResponseWriter, username string) string {

	t.Lock()
	defer t.Unlock()

	h := md5.New()
	token := fmt.Sprintf("%x", h.Sum([]byte(time.Now().Format("05.999999999Z07:00"))))
	t.Storage[token] = username

	// выдаем токен на 1 месяц
	SetCookie(w, "_token", token, 30*24*time.Hour)

	return token
}

// Delete () стираем токен
func (t *Tokens) Delete(w http.ResponseWriter, username string) {

	t.Lock()
	defer t.Unlock()

	// ставим пустые куки
	SetCookie(w, "_token", "", 0*time.Second)

	delete(t.Storage, username)

	t.Save()
}

// Get () получить запись
func (t *Tokens) Get(username string) (string, bool) {

	t.Lock()
	defer t.Unlock()

	n, ok := t.Storage[username]

	return n, ok
}

// Read () прочитать кэш из файла
func (t *Tokens) Read() *Tokens {

	if !t.ExistFile() {
		t.Save()
	}

	f, _ := ioutil.ReadFile(config.Sessions)
	json.Unmarshal(f, &t.Storage)

	return t
}

// OpenConnects () откроем соединения
func (t *Tokens) OpenConnects() {
	for _, username := range t.Storage {
		// ищем пользователя в кэше
		note, found := cache.Get(username)
		// если не нашли пользователя
		if !found {
			logger.Printf("Error open connect for %s, user not found", username)
			continue
		}
		// открываем соединение
		login, password := note.EncodeBasicAuth()
		note.conn = newHTTPService(config.Service1C, login, password)
	}
}

// Save () сохранить кэш в файла
func (t *Tokens) Save() *Tokens {

	data, _ := json.Marshal(t.Storage)
	ioutil.WriteFile(config.Sessions, data, os.ModePerm)

	t.Read()

	return t
}
