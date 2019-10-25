package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/actopolus/project/api"
	"github.com/actopolus/project/web"
)

// конфиг
type config struct {
	Listen string
	URL    api.URL
	Http   struct {
		Timeout string
	}
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

	apiReq, err := api.NewApi(cfg.Http.Timeout, cfg.URL)

	if err != nil {
		log.Fatalln(err)
	}

	app := web.NewApp(apiReq)
	http.HandleFunc("/info/", app.Info)

	log.Println("Listen and serve on", cfg.Listen)
	if err := http.ListenAndServe(cfg.Listen, nil); err != nil {
		log.Fatalln(err)
	}
}
