// Represent essential entrypoint for hidra
package main

import (
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/namsral/flag"

	"github.com/hidracloud/hidra/agent"
	"github.com/hidracloud/hidra/api"
	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/prometheus"
	"github.com/hidracloud/hidra/scenarios"
	_ "github.com/hidracloud/hidra/scenarios/all"
	"github.com/joho/godotenv"
)

type flagConfig struct {
	// Which configuration file to use
	testFile string
	// Which port to listen
	listenAddr string
	// Which port to listen for metrics
	metricsListenAddr string
	// How often to pull metrics
	metricsPullSeconds int
	// Agent secret
	agentSecret string
	// API endpoint
	apiEndpoint string
	// Data directory
	dataDir string
	// Database technology
	dbType string
	// Database path
	dbPath string
	// Database uri
	dbUri string
}

// This mode is used for fast checking yaml
func runTestMode(cfg *flagConfig, wg *sync.WaitGroup) {
	if cfg.testFile == "" {
		log.Fatal("testFile expected to be not null")
	}

	if _, err := os.Stat(cfg.testFile); os.IsNotExist(err) {
		log.Fatal("testFile does not exists")
	}

	log.Println("Running hidra in test mode")
	data, err := ioutil.ReadFile(cfg.testFile)

	if err != nil {
		log.Fatal(err)
	}

	slist, err := models.ReadScenariosYAML(data)
	if err != nil {
		log.Fatal(err)
	}

	m := scenarios.RunScenario(slist.Scenario, slist.Name, slist.Description)

	if m.Error != nil {
		log.Fatal(m.Error)
	}

	scenarios.PrettyPrintScenarioResults(m, slist.Name, slist.Description)
	wg.Done()
}

func runAgentMode(cfg *flagConfig, wg *sync.WaitGroup) {
	log.Println("Running hidra in agent mode")
	agent.StartAgent(cfg.apiEndpoint, cfg.agentSecret, cfg.dataDir)
	wg.Done()
}

func runApiMode(cfg *flagConfig, wg *sync.WaitGroup) {
	log.Println("Running hidra in api mode")
	api.StartApi(cfg.listenAddr, cfg.dbType)
	wg.Done()
}

func runMetricMode(cfg *flagConfig, wg *sync.WaitGroup) {
	log.Println("Running hidra in metric mode")
	prometheus.StartPrometheus(cfg.metricsListenAddr, cfg.metricsPullSeconds)
	wg.Done()
}

func main() {
	godotenv.Load()

	// Start default configuration
	cfg := flagConfig{}

	// Initialize flags
	var agentMode, apiMode, testMode, metricMode bool

	// Operating mode
	flag.BoolVar(&apiMode, "api", false, "-api enable api mode in given hidra")
	flag.BoolVar(&agentMode, "agent", false, "-agent enable agent mode in given hidra")
	flag.BoolVar(&testMode, "test", false, "-test enable test mode in given hidra")
	flag.BoolVar(&metricMode, "metric", false, "-metric metric mode in given hidra")

	// Test mode
	flag.StringVar(&cfg.testFile, "file", "", "-file your_test_file_yaml")

	// API mode
	flag.StringVar(&cfg.listenAddr, "listen_addr", ":8080", "-listen_addr listen address")
	flag.StringVar(&cfg.dbType, "db_type", "sqlite", "-db_type which type of database you want to use?")
	flag.StringVar(&cfg.dbPath, "db_path", "test.db", "-db_path database path")
	flag.StringVar(&cfg.dbUri, "db_uri", "", "-db_uri database uri")

	// Metric mode
	flag.StringVar(&cfg.metricsListenAddr, "metric_listen_addr", ":9096", "-metric_listen_addr listen address")
	flag.IntVar(&cfg.metricsPullSeconds, "metric_pull_seconds", 15, "-metric_pull_seconds time to pull for new metrics")

	// Agent mode
	flag.StringVar(&cfg.agentSecret, "agent_secret", "", "-agent_secret for registering this agent")
	flag.StringVar(&cfg.apiEndpoint, "api_url", "http://localhost:8080/api", "-api_url where is api url?")
	flag.StringVar(&cfg.dataDir, "data_dir", "/var/lib/hidra/data", "-data_dir where you want to store agent data?")

	flag.Parse()

	var wg sync.WaitGroup

	if apiMode || metricMode {
		models.SetupDB(cfg.dbType, cfg.dbPath, cfg.dbUri)
	}

	if agentMode {
		wg.Add(1)
		go runAgentMode(&cfg, &wg)
	}

	if apiMode {
		wg.Add(1)
		go runApiMode(&cfg, &wg)
	}

	if metricMode {
		wg.Add(1)
		go runMetricMode(&cfg, &wg)

	}

	if testMode {
		wg.Add(1)
		go runTestMode(&cfg, &wg)
	}

	wg.Wait()
}
