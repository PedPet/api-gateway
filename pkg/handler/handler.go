package handler

import (
	"fmt"
	"net/http"

	"github.com/PedPet/api-gateway/pkg/endpoint"
	"github.com/gorilla/mux"
)

// MakeHandlers creates the route handlers for the api routes
func MakeHandlers(r *mux.Router, e endpoint.Endpoints) {
	makeUserHandlers(r, e.User)
}

func methodNotAllowedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := r.Method
		http.Error(w, fmt.Sprintf("%s method not allowed", m), http.StatusMethodNotAllowed)
	}
}
