package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// Валюта банка
type valute struct {
	ID       string  `json:"ID"`
	NumCode  string  `json:"NumCode"`
	CharCode string  `json:"CharCode"`
	Nominal  int     `json:"Nominal"`
	Name     string  `json:"Name"`
	Value    float32 `json:"Value"`
	Previous float32 `json:"Previous"`
}

// Список валют из банка
type bankCurrencies struct {
	Valutes map[string]valute `json:"Valute"`
}

// Описание валюты
type countryCurrency struct {
	Code string `json:"code"`
}

// Описание страны
type country struct {
	NativeName string            `json:"nativeName"`
	Flag       string            `json:"flag"`
	Currencies []countryCurrency `json:"bankCurrencies"`
}

// Информация о курсе валют
type Info struct {
	Name   string
	Flag   string
	Values map[string]float32
}

// Адреса API
type URL struct {
	Countries  string
	Currencies string
}

// Обёртка для запросов
type Api struct {
	client http.Client
	urls   URL
}

// NewApi - создать новое Api
func NewApi(timeout string, urls URL) (Api, error) {
	t, err := time.ParseDuration(timeout)

	if err != nil {
		return Api{}, err
	}

	client := http.Client{Timeout: t}

	return Api{client: client, urls: urls}, nil
}

// request - сделать запрос по url
func (a *Api) request(url string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, out)
}

// Info - получить информацию о курсе
func (a *Api) Info(name string) Info {
	var (
		wg    sync.WaitGroup
		currs bankCurrencies
		cntry country
	)

	wg.Add(2)
	go func() {
		if err := a.request(a.urls.Currencies, &currs); err != nil {
			log.Println(err)
		}
		wg.Done()
	}()

	go func() {
		var tmp []country
		if err := a.request(a.urls.Countries+name, &tmp); err != nil {
			log.Println(err)
		} else if len(tmp) > 0 {
			cntry = tmp[0]
		}
		wg.Done()
	}()
	wg.Wait()

	info := Info{Values: make(map[string]float32)}

	for _, cur := range cntry.Currencies {
		if v, ok := currs.Valutes[cur.Code]; ok {
			info.Values[cur.Code] = float32(v.Nominal) / v.Value
		}
	}

	info.Name = cntry.NativeName
	info.Flag = cntry.Flag

	return info
}
