package api

import (
	"log"
	"net/http"

	"github.com/JoseCarlosGarcia95/hidra/models"
	"github.com/gorilla/mux"
)

// Represent API object
type API struct {
	router http.Handler
}

func StartApi(serverAddr string) {
	log.Printf("Starting api at %s\n", serverAddr)

	api := API{}
	r := mux.NewRouter()

	// Public functions
	r.HandleFunc("/ping", api.Ping).Methods(http.MethodGet)
	r.HandleFunc("/login", api.Login).Methods(http.MethodPost)

	// User registered functions
	r.Handle("/agent_token", models.AuthMiddleware(http.HandlerFunc(api.AgentToken))).Methods(http.MethodGet)

	// Pre-register agent functions
	r.Handle("/register_agent", models.AuthRegisterAgentMiddleware(http.HandlerFunc(api.RegisterAgent))).Methods(http.MethodPost)

	api.router = r

	log.Fatal(http.ListenAndServe(":8080", api.router))
}
