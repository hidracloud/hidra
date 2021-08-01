// Represent an API object
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

// Start a new API process
func StartApi(serverAddr string) {
	log.Printf("Starting api at %s\n", serverAddr)

	api := API{}
	r := mux.NewRouter()

	// Public functions
	r.HandleFunc("/ping", api.Ping).Methods(http.MethodGet)
	r.HandleFunc("/login", api.Login).Methods(http.MethodPost)

	// User registered functions
	r.Handle("/register_sample", models.AuthMiddleware(http.HandlerFunc(api.RegisterSample))).Methods(http.MethodPost)

	// Pre-register agent functions
	r.Handle("/register_agent", models.AuthMiddleware(http.HandlerFunc(api.RegisterAgent))).Methods(http.MethodPost)

	// Registered agent functions
	r.Handle("/agent_list_samples", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentListSamples))).Methods(http.MethodGet)
	r.Handle("/agent_push_metrics/{sampleid}", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentPushMetrics))).Methods(http.MethodPost)
	r.Handle("/agent_get_sample/{sampleid}", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentGetSample))).Methods(http.MethodGet)

	api.router = r

	log.Fatal(http.ListenAndServe(":8080", api.router))
}
