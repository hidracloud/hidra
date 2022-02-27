// Represent essential entrypoint for hidra
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/namsral/flag"

	"github.com/hidracloud/hidra/src/exporter"
	"github.com/hidracloud/hidra/src/models"
	"github.com/hidracloud/hidra/src/scenarios"
	_ "github.com/hidracloud/hidra/src/scenarios/all"
	"github.com/hidracloud/hidra/src/utils"
	"github.com/joho/godotenv"
)

type flagConfig struct {
	// Which configuration file to use
	testFile string

	// Which conf path to use
	confPath string

	// Buckets
	buckets string

	// maxExecutor is the maximum number of executor to run in parallel
	maxExecutor int

	// port is the port to listen on
	port int

	// duration is the duration of the attack
	duration int

	// workers is the number of workers to use
	workers int
}

// This mode is used for fast checking yaml
func runTestMode(cfg *flagConfig, wg *sync.WaitGroup) {
	if cfg.testFile == "" {
		log.Fatal("testFile expected to be not null")
	}

	var configFiles = []string{cfg.testFile}

	// Check if test file ends with .yaml or .yml
	if !strings.Contains(cfg.testFile, ".yaml") && !strings.Contains(cfg.testFile, ".yml") {
		configFiles, _ = utils.AutoDiscoverYML(cfg.testFile)
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			log.Fatal("testFile does not exists")
		}

		log.Println("Running hidra in test mode")
		log.Println("Running ", configFile)
		data, err := ioutil.ReadFile(configFile)

		if err != nil {
			log.Fatal(err)
		}

		slist, err := models.ReadSampleYAML(data)
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

		scenarios.PrettyPrintScenarioResults(m, slist.Name, slist.Description)
	}
	wg.Done()

}

func runExporter(cfg *flagConfig, wg *sync.WaitGroup) {
	buckets := make([]float64, 0)

	bucketStrArray := strings.Split(cfg.buckets, ",")

	for _, bucketStr := range bucketStrArray {
		bucket, err := strconv.ParseFloat(bucketStr, 64)
		if err != nil {
			panic(err)
		}
		buckets = append(buckets, bucket)
	}

	fmt.Println(buckets)

	exporter.Run(wg, cfg.confPath, cfg.maxExecutor, cfg.port, buckets)
}

func runSyntaxMode(cfg *flagConfig, wg *sync.WaitGroup) {
	var configFiles = []string{cfg.testFile}

	// Check if test file ends with .yaml or .yml
	if !strings.Contains(cfg.testFile, ".yaml") && !strings.Contains(cfg.testFile, ".yml") {
		configFiles, _ = utils.AutoDiscoverYML(cfg.testFile)
	}

	hasError := false

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			log.Fatal("testFile does not exists")
		}

		data, err := ioutil.ReadFile(configFile)

		if err != nil {
			hasError = true
			log.Fatal(err)
		}

		_, err = models.ReadSampleYAML(data)
		if err != nil {
			hasError = true
			log.Println("Syntax error: ", err, " in ", configFile)
		}
	}

	if hasError {
		os.Exit(1)
	}

	wg.Done()
}

func runAttackMode(cfg *flagConfig, wg *sync.WaitGroup) {
	data, err := ioutil.ReadFile(cfg.testFile)
	if err != nil {
		panic(err)
	}

	sample, err := models.ReadSampleYAML(data)
	if err != nil {
		panic(err)
	}

	requestCounter := make([]int, cfg.workers+1)

	tchan := make(chan int)
	go func(c chan<- int) {
		for ti := 1; ti <= cfg.workers; ti++ {
			log.Printf("Thread Count %d", ti)
			requestCounter[ti] = 0
			c <- ti
		}
	}(tchan)

	timeout := time.After(time.Duration(cfg.duration) * time.Second)
	for {
		select { //Select blocks the flow until one of the channels receives a message
		case <-timeout: //receives a msg when execution duration is over
			log.Printf("Execution completed")

			// sum all request counter
			totalRequest := 0
			for _, request := range requestCounter {
				totalRequest += request
			}

			log.Println("Scenarios executed: ", totalRequest, "rate ", float64(totalRequest)/float64(cfg.duration))
			wg.Done()
			return
		case ts := <-tchan: //receives a message when a new user thread has to be initiated
			log.Printf("Thread No %d started", ts)
			go func(t int) {
				for {
					scenarios.RunScenario(sample.Scenario, sample.Name, sample.Description)

					requestCounter[t] = requestCounter[t] + 1
				}
			}(ts)
		}
	}
}

func main() {
	godotenv.Load()

	// Start default configuration
	cfg := flagConfig{}

	// Initialize flags
	var testMode, exporter, syntaxMode, attackMode bool

	// Operating mode
	flag.BoolVar(&testMode, "test", false, "-test enable test mode in given hidra")
	flag.BoolVar(&exporter, "exporter", false, "-exporter enable exporter mode in given hidra")
	flag.BoolVar(&syntaxMode, "syntax", false, "-syntax enable syntax mode in given hidra")
	flag.BoolVar(&attackMode, "attack", false, "-attack enable attack mode in given hidra")

	// Test mode
	flag.StringVar(&cfg.testFile, "file", "", "-file your_test_file_yaml")

	// Exporter mode
	flag.IntVar(&cfg.maxExecutor, "maxExecutor", 1, "-maxExecutor your_max_executor")
	flag.IntVar(&cfg.port, "port", 19090, "-port your_port")
	flag.StringVar(&cfg.buckets, "buckets", "100,200,500,1000,2000,5000", "-buckets your_buckets")
	flag.StringVar(&cfg.confPath, "conf", "", "-conf your_conf_path")
	flag.IntVar(&cfg.duration, "duration", 10, "-duration your_duration")
	flag.IntVar(&cfg.workers, "workers", 10, "-workers your_workers")

	// Attack mode
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

	if syntaxMode {
		wg.Add(1)
		go runSyntaxMode(&cfg, &wg)
	}

	if attackMode {
		wg.Add(1)
		go runAttackMode(&cfg, &wg)
	}

	wg.Wait()
}
