package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/actopolus/project/api"
)

// Обработчик заппросов
type App struct {
	api api.Api
}

// NewApi - Новый обработчик запросов
func NewApp(api api.Api) App {
	return App{api: api}
}

// Info - обработчик запроса info
func (a *App) Info(w http.ResponseWriter, r *http.Request) {
	lst := strings.LastIndexByte(r.URL.Path, '/')
	p := r.URL.Path[lst+1:]

	w.Header().Add("Content-Type", "application/json")

	if len(p) == 0 || lst < 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status": "bad request"}`))
		return
	}

	info := a.api.Info(p)
	if body, err := json.Marshal(info); err == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
	}
}
