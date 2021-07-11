package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/JoseCarlosGarcia95/hidra/models"
	"github.com/JoseCarlosGarcia95/hidra/scenarios"
	_ "github.com/JoseCarlosGarcia95/hidra/scenarios/all"
)

type flagConfig struct {
	hidraMode int
	testFile  string
}

// This mode is used for fast checking yaml
func runTestMode(cfg *flagConfig) {
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

	for _, s := range slist.Scenarios {
		m := scenarios.RunScenario(s)
		scenarios.PrettyPrintScenarioMetrics(m)
	}

}

func main() {
	// Start default configuration
	cfg := flagConfig{
		hidraMode: 0,
	}

	// Initialize flags
	flag.StringVar(&cfg.testFile, "testFile", "", "--testFile your-test-file-yaml")
	flag.Parse()

	// Run hidra in test mode, for testing yaml
	if cfg.hidraMode == 0 {
		runTestMode(&cfg)
	}
}
