package main

import (
	"go-microservice/order"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	AccountURL  string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL  string `envconfig:"CATALOG_SERVICE_URL"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	r, err := order.NewPostgresReposytory(cfg.DatabaseURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Close()

	s := order.NewService(r)

	if err = order.ListenGRPC(s, cfg.AccountURL, cfg.CatalogURL, 50051); err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on port 50051")
}
