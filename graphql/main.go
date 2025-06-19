package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	AccountURL string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL string `envconfig:"CATALOG_SERVICE_URL"`
	OrderURL   string `envconfig:"ORDER_SERVICE_URL"`
}

func main() {
	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	s, err := NewGraphQlServer(cfg.AccountURL, cfg.CatalogURL, cfg.OrderURL)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/graphql", handler.GraphQL(s.ToExecutableSchema()))
	http.Handle("/playground", handler.Playground("akhil", "/graphql"))

	log.Fatal(http.ListenAndServe(":8080", nil))

	// router := http.NewServeMux()
	// router.Handle("/graphql", handler.New(s.ToExecutableSchema()))
	// router.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))

	// c := cors.New(cors.Options{
	// 	AllowedOrigins: []string{"*"},
	// 	AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	// 	AllowedHeaders: []string{"Content-Type"},
	// })

	// server := &http.Server{
	// 	Addr:    ":8080",
	// 	Handler: c.Handler(router),
	// }

	// log.Println("Server started at http://localhost:8080")
	// log.Fatal(server.ListenAndServe())
}
