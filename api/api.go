// Represent an API object
package api

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
	_ "github.com/hidracloud/hidra/prometheus"
	"github.com/rs/cors"
)

// Represent API object
type API struct {
	router http.Handler
}

// Run a cron for API process
func cron() {
	log.Println("Starting cron for API")

	for {
		models.DeleteExpiredMetrics()
		time.Sleep(time.Second)
	}
}

//go:embed external/hidra-frontend/app/build
var webapp embed.FS

func getWebApp() http.FileSystem {
	fsys, err := fs.Sub(webapp, "external/hidra-frontend/app/build")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
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
	r.Handle("/api/update_sample/{sampleid}", models.AuthMiddleware(http.HandlerFunc(api.UpdateSample))).Methods(http.MethodPut)
	r.Handle("/api/list_samples", models.AuthMiddleware(http.HandlerFunc(api.ListSamples))).Methods(http.MethodGet)
	r.Handle("/api/list_agents", models.AuthMiddleware(http.HandlerFunc(api.ListAgents))).Methods(http.MethodGet)
	r.Handle("/api/verify_token", models.AuthMiddleware(http.HandlerFunc(api.VerifyToken))).Methods(http.MethodGet)
	r.Handle("/api/get_sample/{sampleid}", models.AuthMiddleware(http.HandlerFunc(api.AgentGetSample))).Methods(http.MethodGet)
	r.Handle("/api/get_agent/{agentid}", models.AuthMiddleware(http.HandlerFunc(api.GetAgent))).Methods(http.MethodGet)
	r.Handle("/api/update_agent/{agentid}", models.AuthMiddleware(http.HandlerFunc(api.UpdateAgent))).Methods(http.MethodPut)

	// Pre-register agent functions
	r.Handle("/api/register_agent", models.AuthMiddleware(http.HandlerFunc(api.RegisterAgent))).Methods(http.MethodPost)

	// Registered agent functions
	r.Handle("/api/agent_list_samples", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentListSamples))).Methods(http.MethodGet)
	r.Handle("/api/agent_push_metrics/{sampleid}", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentPushMetrics))).Methods(http.MethodPost)
	r.Handle("/api/agent_get_sample/{sampleid}", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentGetSample))).Methods(http.MethodGet)

	// React webapp
	r.PathPrefix("/static").Handler(http.FileServer(getWebApp()))
	r.Handle("/", http.FileServer(getWebApp()))

	c := cors.AllowAll()

	api.router = c.Handler(r)

	go cron()

	log.Fatal(http.ListenAndServe(":8080", c.Handler(api.router)))
}
