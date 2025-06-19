package main

import (
	"go-microservice/catalog"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(20 * time.Second)

	r, err := catalog.NewElasticReposytory(cfg.DatabaseURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Close()

	s := catalog.NewService(r)

	if err = catalog.ListenGRPC(s, 50051); err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on port 50051")
}
