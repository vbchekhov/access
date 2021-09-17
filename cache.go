package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var cache *Cache

// Note запись
type Note struct {
	UserName  string
	Password  string
	BasicAuth string
	Secret    string
	Timestamp uint64
	conn      *httpService
	key       *otp.Key
}

// EncodeBasicAuth () фомрирование логина и пароля
func (n *Note) EncodeBasicAuth() (login, password string) {

	// убираем переносы строк
	// по средствам n.BasicAuth[:len(n.BasicAuth)-1]
	// то есть сокращаем последний символ
	encode, _ := b64.StdEncoding.DecodeString(n.BasicAuth)
	encodeStr := strings.TrimRight(string(encode), "\r\n")
	split := strings.Split(encodeStr, ":")

	if len(split) < 2 {
		return "", ""
	}

	return split[0], split[1]
}

// NewKey () генерация ключа
func (n *Note) NewKey() (key *otp.Key, err error) {

	c := totp.GenerateOpts{
		Issuer:      "vmpauto.local",
		AccountName: n.UserName + "@vmpauto.local",
	}

	if key, err = totp.Generate(c); err != nil {
		logger.Printf("Error create 2fa key %v", err)
		return nil, err
	}

	n.Secret = key.Secret()
	n.Timestamp = key.Period()
	n.key = key

	return key, err
}

// NewQRCode ()
func (n *Note) NewQRCode() (pic []byte, err error) {

	var buf bytes.Buffer

	img, err := n.key.Image(300, 300)
	if err != nil {
		logger.Printf("Error create Image() - %v", err)
		return []byte{}, nil
	}

	err = png.Encode(&buf, img)
	if err != nil {
		logger.Printf("Error create png Encode() - %v", err)
		return []byte{}, nil
	}

	return buf.Bytes(), nil
}

// NewBasicAuth () генерация ключа
func (n *Note) NewBasicAuth() (key string, err error) {

	var b64 map[string]string
	_, b, err := views.makeRequest("/service/token?domainuser="+n.UserName, "GET", nil, nil)
	if err != nil {
		logger.Printf("Error NewBasicAuth() - %v", err)
		return "", err
	}
	if err = json.Unmarshal(b, &b64); err != nil {
		logger.Printf("Error NewBasicAuth() unmarhasl - %v", err)
		return "", err
	}
	n.BasicAuth = b64[n.UserName]

	return n.BasicAuth, err
}

// Cache хранилище ключей
type Cache struct {
	sync.Mutex
	Storage map[string]*Note
}

// NewCache () новый кэш
func NewCache() *Cache {
	return &Cache{Storage: make(map[string]*Note)}
}

// ExistFile () проверка существования файла
func (c *Cache) ExistFile() bool {

	if _, err := os.Open(config.Cache); os.IsNotExist(err) {
		return false
	}

	return true
}

// Add () добавить запись
func (c *Cache) Add(n *Note) {

	c.Lock()
	defer c.Unlock()

	n.UserName = strings.ToLower(n.UserName)
	c.Storage[n.UserName] = n
}

// Get () получить запись
func (c *Cache) Get(username string) (*Note, bool) {

	c.Lock()
	defer c.Unlock()

	n, ok := c.Storage[strings.ToLower(username)]

	return n, ok
}

func (c *Cache) GetBySecret(key string) (*Note, bool) {

	c.Lock()
	defer c.Unlock()

	for _, note := range c.Storage {
		if note.Secret == key {
			return note, true
		}
	}

	return &Note{}, false
}

// Read () прочитать кэш из файла
func (c *Cache) Read() *Cache {

	if !c.ExistFile() {
		c.Save()
	}

	f, _ := ioutil.ReadFile(config.Cache)
	json.Unmarshal(f, &c.Storage)

	return c
}

// Save () сохранить кэш в файла
func (c *Cache) Save() *Cache {

	data, _ := json.Marshal(c.Storage)
	ioutil.WriteFile(config.Cache, data, os.ModePerm)

	c.Read()

	return c
}

type CacheTable []CacheTableSting

type CacheTableSting struct {
	Num       int
	UserName  string
	BasicAuth string
	Secret    string
	Sesions   bool
}

func (n *Note) Active() bool {
	return true
}

func (c *Cache) Table() CacheTable {

	var ct CacheTable
	var num int
	for index := range cache.Storage {
		num++
		ct = append(ct, CacheTableSting{
			num,
			cache.Storage[index].UserName,
			cache.Storage[index].BasicAuth,
			cache.Storage[index].Secret,
			cache.Storage[index].Active(),
		})
	}

	return ct
}
