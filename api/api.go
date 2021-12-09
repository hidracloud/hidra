// Package api Represent an API object
package api

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
	"github.com/rs/cors"
)

// API Represent API object
type API struct {
	dbType string
	router http.Handler
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

func processMetricsQueue() {
	for {
		models.ProcessMetricsQueue()
		time.Sleep(time.Second * 5)
	}
}

// StartAPI Start a new API process
func StartAPI(serverAddr, dbtype string) {
	log.Printf("Starting api at %s\n", serverAddr)

	api := API{}
	api.dbType = dbtype
	r := mux.NewRouter()

	// Public functions
	r.HandleFunc("/api/ping", api.Ping).Methods(http.MethodGet)
	r.HandleFunc("/api/setup_status", api.SetupStatus).Methods(http.MethodGet)
	r.HandleFunc("/api/create_setup", api.CreateSetup).Methods(http.MethodPost)

	r.HandleFunc("/api/login", api.Login).Methods(http.MethodPost)

	// User registered functions
	r.Handle("/api/me", models.AuthMiddleware(http.HandlerFunc(api.GetMe))).Methods(http.MethodGet)
	r.Handle("/api/register_sample", models.AuthMiddleware(http.HandlerFunc(api.RegisterSample))).Methods(http.MethodPost)
	r.Handle("/api/update_sample/{sampleid}", models.AuthMiddleware(http.HandlerFunc(api.UpdateSample))).Methods(http.MethodPut)
	r.Handle("/api/list_samples", models.AuthMiddleware(http.HandlerFunc(api.ListSamples))).Methods(http.MethodGet)
	r.Handle("/api/list_agents", models.AuthMiddleware(http.HandlerFunc(api.ListAgents))).Methods(http.MethodGet)
	r.Handle("/api/verify_token", models.AuthMiddleware(http.HandlerFunc(api.VerifyToken))).Methods(http.MethodGet)
	r.Handle("/api/system_info", models.AuthMiddleware(http.HandlerFunc(api.SystemInfo))).Methods(http.MethodGet)
	r.Handle("/api/list_scenarios_steps", models.AuthMiddleware(http.HandlerFunc(api.ListScenariosSteps))).Methods(http.MethodGet)
	r.Handle("/api/get_sample/{sampleid}", models.AuthMiddleware(http.HandlerFunc(api.AgentGetSample))).Methods(http.MethodGet)
	r.Handle("/api/get_sample_result/{sampleid}", models.AuthMiddleware(http.HandlerFunc(api.GetSampleResult))).Methods(http.MethodGet)
	r.Handle("/api/get_metrics/{sampleid}", models.AuthMiddleware(http.HandlerFunc(api.GetMetrics))).Methods(http.MethodGet)
	r.Handle("/api/get_agent/{agentid}", models.AuthMiddleware(http.HandlerFunc(api.GetAgent))).Methods(http.MethodGet)
	r.Handle("/api/update_agent/{agentid}", models.AuthMiddleware(http.HandlerFunc(api.UpdateAgent))).Methods(http.MethodPut)
	r.Handle("/api/sample_runner", models.AuthMiddleware(http.HandlerFunc(api.SampleRunner))).Methods(http.MethodPost)
	r.Handle("/api/update_password", models.AuthMiddleware(http.HandlerFunc(api.UpdatePassword))).Methods(http.MethodPost)
	r.Handle("/api/2fa_qrcode", models.AuthMiddleware(http.HandlerFunc(api.TwofactorQrcode))).Methods(http.MethodGet)
	r.Handle("/api/configure_2fa", models.AuthMiddleware(http.HandlerFunc(api.TwoFaConfiguration))).Methods(http.MethodPost)
	r.Handle("/api/disable_2fa", models.AuthMiddleware(http.HandlerFunc(api.DisableTwoFaConfiguration))).Methods(http.MethodPost)
	r.Handle("/api/delete_sample/{sampleid}", models.AuthMiddleware(http.HandlerFunc(api.DeleteSample))).Methods(http.MethodDelete)
	// Pre-register agent functions
	r.Handle("/api/register_agent", models.AuthMiddleware(http.HandlerFunc(api.RegisterAgent))).Methods(http.MethodPost)

	// Registered agent functions

	r.Handle("/api/agent_list_samples", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentListSamples))).Methods(http.MethodGet)
	r.Handle("/api/agent_push_metrics/{sampleid}", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentPushMetrics))).Methods(http.MethodPost)
	r.Handle("/api/agent_get_sample/{sampleid}", models.AuthSecretAgentMiddleware(http.HandlerFunc(api.AgentGetSample))).Methods(http.MethodGet)

	// React webapp
	r.PathPrefix("/static").Handler(http.FileServer(getWebApp()))
	r.Handle("/", http.FileServer(getWebApp()))

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webapp := getWebApp()

		file, err := webapp.Open("index.html")

		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		defer file.Close()
		http.ServeContent(w, r, "index.html", time.Now(), file)
	})

	c := cors.AllowAll()

	api.router = c.Handler(r)
	go processMetricsQueue()
	log.Fatal(http.ListenAndServe(serverAddr, c.Handler(api.router)))
}
