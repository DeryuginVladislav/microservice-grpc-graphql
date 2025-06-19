package main

import (
	"go-microservice/account"
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

	time.Sleep(10 * time.Second)
	
	r, err := account.NewPostgresReposytory(cfg.DatabaseURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Close()

	s := account.NewService(r)

	if err = account.ListenGRPC(s, 50051); err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on port 50051")
}
