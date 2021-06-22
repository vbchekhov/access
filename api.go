package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/html/charset"
)

var views *httpService

type httpService struct {
	server, login, password string
	client                  *http.Client
}

func newHTTPService(server, login, password string) *httpService {
	return &httpService{
		server:   server,
		login:    login,
		password: password,
		client: &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		},
	}
}

func (api *httpService) makeRequest(point, method string, body io.ReadCloser, headers *http.Header) (http.Header ,[]byte, error) {

	req, _ := http.NewRequest(method, api.server+point, body)
	if headers != nil { req.Header = *headers }
	req.SetBasicAuth(api.login, api.password)
	
	resp, err := api.client.Do(req)
	if err != nil {
		return req.Header,[]byte{}, fmt.Errorf("error response to %s: %s", point, err)
	}

	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.Header,[]byte{}, fmt.Errorf("error read all body: %s", err)
	}
	rbody = bytes.TrimPrefix(rbody, []byte("\xef\xbb\xbf"))
	utfbody, err := charset.NewReader(bytes.NewReader(rbody), resp.Header.Get("Content-Type"))
	if err != nil {
		return resp.Header,[]byte{}, fmt.Errorf("error read content: %s", err)
	}

	defer resp.Body.Close()

	reader, err := ioutil.ReadAll(utfbody)

	return resp.Header, reader, err
}
