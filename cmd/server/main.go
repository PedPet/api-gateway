package main

import (
	"context"
	"log"
	"net/http"

	"github.com/PedPet/api-gateway/config"
	"github.com/PedPet/api-gateway/pkg/endpoint"
	"github.com/PedPet/api-gateway/pkg/handler"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	ctx := context.Background()

	r := mux.NewRouter().StrictSlash(false)

	s, err := config.LoadSettings()
	if err != nil {
		log.Fatalf("Failed to load settings: %v", err)
	}

	e := endpoint.MakeEndpoints(ctx, s)
	handler.MakeHandlers(r, e)

	originsOk := handlers.AllowedOrigins([]string{"http://pedpet.local:3000"})
	headersOk := handlers.AllowedHeaders([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	http.ListenAndServeTLS(":443", "./api.pedpet.local/cert.pem", "./api.pedpet.local/key.pem", handlers.CORS(headersOk, originsOk, methodsOk)(r))
}
