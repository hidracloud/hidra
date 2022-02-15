// Represent essential entrypoint for hidra
package main

import (
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/namsral/flag"

	"github.com/hidracloud/hidra/exporter"
	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"
	_ "github.com/hidracloud/hidra/scenarios/all"
	"github.com/joho/godotenv"
)

type flagConfig struct {
	// Which configuration file to use
	testFile string

	// Which conf path to use
	confPath string
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

	if len(slist.Tags) > 0 {
		log.Println("Tags:")
		for key, val := range slist.Tags {
			log.Println(" ", key, "=", val)
		}
	}

	m := scenarios.RunScenario(slist.Scenario, slist.Name, slist.Description)

	if m.Error != nil {
		log.Fatal(m.Error)
	}

	scenarios.PrettyPrintScenarioResults(m, slist.Name, slist.Description)
	wg.Done()
}

func runExporter(cfg *flagConfig, wg *sync.WaitGroup) {
	exporter.Run(wg, cfg.confPath)
}

func main() {
	godotenv.Load()

	// Start default configuration
	cfg := flagConfig{}

	// Initialize flags
	var testMode, exporter bool

	// Operating mode
	flag.BoolVar(&testMode, "test", false, "-test enable test mode in given hidra")
	flag.BoolVar(&exporter, "exporter", false, "-exporter enable exporter mode in given hidra")

	// Test mode
	flag.StringVar(&cfg.testFile, "file", "", "-file your_test_file_yaml")
	flag.StringVar(&cfg.confPath, "conf", "", "-conf your_conf_path")

	flag.Parse()

	var wg sync.WaitGroup

	if testMode {
		wg.Add(1)
		go runTestMode(&cfg, &wg)
	}

	if exporter {
		wg.Add(1)
		go runExporter(&cfg, &wg)
	}

	wg.Wait()
}
