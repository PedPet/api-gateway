package handler

import (
	"github.com/PedPet/api-gateway/pkg/endpoint"
	"github.com/gorilla/mux"
)

func makeUserHandlers(r *mux.Router, e endpoint.User) {
	ur := r.PathPrefix("/user").Subrouter()
	ur.MethodNotAllowedHandler = methodNotAllowedHandler()

	ur.HandleFunc("/register", e.RegisterEndpoint).Methods("POST")
	ur.HandleFunc("/confirm", e.ConfirmUserEndpoint).Methods("POST")
	ur.HandleFunc("/resend-confirmation", e.ResendConfirmationEndpoint).Methods("POST")
	ur.HandleFunc("/check-username-taken", e.CheckUsernameEndpoint).Methods("GET")
	ur.HandleFunc("/login", e.LoginEndpoint).Methods("POST")
}
