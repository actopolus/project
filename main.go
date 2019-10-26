package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Конфигурация
type config struct {
	Listen    string
	Timeout   time.Duration
	Countries string
}

// Обработчик запросов
type Web struct {
	cfg    config
	client http.Client
}

// Структура ответа
type Response struct {
	Code  int         `xml:"code"`
	Error string      `xml:"error,omitempty"`
	Data  interface{} `xml:"data,omitempty"`
}

// Информация о стране
type Country struct {
	Name    string   `json:"nativeName" xml:"name"`
	Flag    string   `json:"flag" xml:"flag"`
	Borders []string `json:"borders" xml:"borders"`
}

// NewWeb - создаёт новый обработчик запросов
func NewWeb(cfg config) Web {
	return Web{cfg: cfg, client: http.Client{Timeout: cfg.Timeout * time.Millisecond}}
}

// main - точка входа
func main() {
	path := flag.String("config", "", "Path to config file")
	flag.Parse()

	if len(*path) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var cfg config
	if _, err := toml.DecodeFile(*path, &cfg); err != nil {
		log.Fatalln(err)
	}

	web := NewWeb(cfg)
	http.HandleFunc("/", web.Handle)

	log.Println("Listen and serve on", cfg.Listen)
	if err := http.ListenAndServe(cfg.Listen, nil); err != nil {
		log.Fatalln(err)
	}
}

// getCountry - получает название страны из URL
func (w *Web) getCountry(url string) (country string, err error) {
	pos := strings.LastIndexByte(url, '/')

	if pos == -1 || pos == len(url) {
		return country, errors.New(`Param "country" isn't valid`)
	}

	return url[pos+1:], nil
}

// sendError - выводит ошибку
func (w *Web) sendError(rw http.ResponseWriter, statusCode int, err error) {
	body, _ := xml.Marshal(Response{
		Code:  1,
		Error: err.Error(),
	})

	rw.WriteHeader(statusCode)
	_, _ = rw.Write(body)
}

// Handle - отдаёт информацию о стране
func (w *Web) Handle(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/xml")

	country, err := w.getCountry(req.URL.Path)
	if err != nil {
		w.sendError(rw, http.StatusBadRequest, err)
		return
	}

	resp, err := w.client.Get(w.cfg.Countries + country)
	if err != nil {
		w.sendError(rw, http.StatusInternalServerError, err)
		return
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.sendError(rw, http.StatusInternalServerError, err)
		return
	}

	var response []Country
	if err := json.Unmarshal(b, &response); err != nil {
		w.sendError(rw, http.StatusInternalServerError, err)
		return
	}

	body, _ := xml.Marshal(response[0])

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(body)
}
