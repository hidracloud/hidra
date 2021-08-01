// Represent an API object
package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
	_ "github.com/hidracloud/hidra/prometheus"
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
	r.HandleFunc("/api/ping", api.Ping).Methods(http.MethodGet)
	r.HandleFunc("/api/login", api.Login).Methods(http.MethodPost)

	// User registered functions
	r.Handle("/api/register_sample", models.AuthMiddleware(http.HandlerFunc(api.RegisterSample))).Methods(http.MethodPost)

	// Pre-register agent functions
	r.Handle("/api/register_agent", models.AuthMiddleware(http.HandlerFunc(api.RegisterAgent))).Methods(http.MethodPost)

	// Registered agent functions
	r.Handle("/api/agent_list_samples", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentListSamples))).Methods(http.MethodGet)
	r.Handle("/api/agent_push_metrics/{sampleid}", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentPushMetrics))).Methods(http.MethodPost)
	r.Handle("/api/agent_get_sample/{sampleid}", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentGetSample))).Methods(http.MethodGet)

	api.router = r

	log.Fatal(http.ListenAndServe(":8080", api.router))
}
