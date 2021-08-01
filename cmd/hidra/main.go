// Represent essential entrypoint for hidra
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/JoseCarlosGarcia95/hidra/agent"
	"github.com/JoseCarlosGarcia95/hidra/api"
	"github.com/JoseCarlosGarcia95/hidra/models"
	"github.com/JoseCarlosGarcia95/hidra/prometheus"
	"github.com/JoseCarlosGarcia95/hidra/scenarios"
	_ "github.com/JoseCarlosGarcia95/hidra/scenarios/all"
	"github.com/joho/godotenv"
)

type flagConfig struct {
	testFile           string
	listenAddr         string
	metricsListenAddr  string
	metricsPullSeconds int
	configFile         string
	agentSecret        string
	apiEndpoint        string
	dataDir            string
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

	scenarios.PrettyPrintScenarioMetrics(m, slist.Name, slist.Description)
	wg.Done()
}

func runAgentMode(cfg *flagConfig, wg *sync.WaitGroup) {
	log.Println("Running hidra in agent mode")
	agent.StartAgent(cfg.apiEndpoint, cfg.agentSecret, cfg.dataDir)
	wg.Done()
}

func runApiMode(cfg *flagConfig, wg *sync.WaitGroup) {
	log.Println("Running hidra in api mode")
	api.StartApi(cfg.listenAddr)
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
	flag.BoolVar(&apiMode, "api", false, "-api enable api mode in given hidra")
	flag.BoolVar(&agentMode, "agent", false, "-agent enable agent mode in given hidra")
	flag.BoolVar(&testMode, "test", false, "-test enable test mode in given hidra")
	flag.BoolVar(&metricMode, "metric", false, "-metric metric mode in given hidra")

	flag.StringVar(&cfg.configFile, "config", "", "-config your configuration")
	flag.StringVar(&cfg.testFile, "file", "", "-file your-test-file-yaml")
	flag.StringVar(&cfg.listenAddr, "listen-addr", ":8080", "-listen-addr listen address")
	flag.StringVar(&cfg.metricsListenAddr, "metric-listen-addr", ":9096", "-metric-listen-addr listen address")
	flag.IntVar(&cfg.metricsPullSeconds, "metric-pull-seconds", 1, "-metric-pull-seconds time to pull for new metrics")

	flag.StringVar(&cfg.agentSecret, "agent-secret", "", "-agent-secret for registering this agent")
	flag.StringVar(&cfg.apiEndpoint, "api-url", "", "-api-url where is api url?")
	flag.StringVar(&cfg.dataDir, "data-dir", "/tmp", "-data-dir where you want to store agent data?")

	flag.Parse()

	var wg sync.WaitGroup

	if apiMode || metricMode {
		models.SetupDB()
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
