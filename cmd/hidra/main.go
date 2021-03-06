// Represent essential entrypoint for hidra
package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/namsral/flag"
	"go.opentelemetry.io/otel/baggage"

	"github.com/hidracloud/hidra/v2/pkg/attack"
	"github.com/hidracloud/hidra/v2/pkg/exporter"
	"github.com/hidracloud/hidra/v2/pkg/models"
	"github.com/hidracloud/hidra/v2/pkg/otel"
	"github.com/hidracloud/hidra/v2/pkg/scenarios"
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/all"
	"github.com/hidracloud/hidra/v2/pkg/utils"
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

	// maxScreenshotExecutor is the maximum number of executor to run in parallel for screenshot
	maxScreenshotExecutor int

	// port is the port to listen on
	port int

	// duration is the duration of the attack
	duration int

	// workers is the number of workers to use
	workers int

	// resultFile is the file to write the results to
	resultFile string

	// exitOnError is a flag to exit on error
	exitOnError bool

	// screenshotOnError is a flag to take a screenshot on error
	screenshotOnError bool

	// screenshotPath is the path to save the screenshot
	screenshotPath string

	// screenshotS3Bucket is the s3 bucket to save the screenshot
	screenshotS3Bucket string

	// screenshotS3Endpoint is the s3 endpoint to save the screenshot
	screenshotS3Endpoint string

	// screenshotS3Region is the s3 region to save the screenshot
	screenshotS3Region string

	// screenshotS3AccessKey is the s3 access key to save the screenshot
	screenshotS3AccessKey string

	// screenshotS3SecretKey is the s3 secret key to save the screenshot
	screenshotS3SecretKey string

	// screenshotS3Prefix is the s3 prefix to save the screenshot
	screenshotS3Prefix string

	// screenshotS3TLS is the s3 tls to save the screenshot
	screenshotS3TLS bool

	// jaegerEndpoint is the jaeger endpoint to save the trace
	jaegerEndpoint string

	// quiet is a flag to quiet the output
	quiet bool
}

func runOneTestConfig(ctx context.Context, configFile string, cfg *flagConfig) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatal("testFile does not exists")
	}

	utils.LogDebug("Running hidra in test mode")
	utils.LogDebug("Running " + configFile)
	data, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Fatal(err)
	}

	slist, err := models.ReadSampleYAML(data)
	if err != nil {
		log.Fatal(err)
	}

	utils.LogDebug("Tags:")
	for key, val := range slist.Tags {
		utils.LogDebug(" " + key + "=" + val)
	}

	m, err := scenarios.RunScenario(ctx, slist.Scenario, slist.Name, slist.Description)
	if err != nil {
		log.Fatal(err)
	}

	scenarios.PrettyPrintScenarioResults(m, slist.Name, slist.Description)
	if m.Error != nil && cfg.exitOnError {
		log.Fatal(m.Error)
		os.Exit(1)
	}
}

// This mode is used for fast checking yaml
func runTestMode(ctx context.Context, cfg *flagConfig, wg *sync.WaitGroup) {
	if cfg.testFile == "" {
		log.Fatal("testFile expected to be not null")
	}

	var configFiles = []string{cfg.testFile}

	// Check if test file ends with .yaml or .yml
	if !strings.Contains(cfg.testFile, ".yaml") && !strings.Contains(cfg.testFile, ".yml") {
		configFiles, _ = utils.AutoDiscoverYML(cfg.testFile)
	}

	for _, configFile := range configFiles {
		runOneTestConfig(ctx, configFile, cfg)
	}
	wg.Done()

}

func runExporter(ctx context.Context, cfg *flagConfig, wg *sync.WaitGroup) {
	buckets := make([]float64, 0)

	bucketStrArray := strings.Split(cfg.buckets, ",")

	for _, bucketStr := range bucketStrArray {
		bucket, err := strconv.ParseFloat(bucketStr, 64)
		if err != nil {
			panic(err)
		}
		buckets = append(buckets, bucket)
	}

	exporter.Run(ctx, wg, cfg.confPath, cfg.maxExecutor, cfg.port, buckets)
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

func runAttackMode(ctx context.Context, cfg *flagConfig, wg *sync.WaitGroup) {
	attack.RunAttackMode(ctx, cfg.testFile, cfg.resultFile, cfg.workers, cfg.duration)
	wg.Done()
}

func main() {
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
	flag.BoolVar(&cfg.exitOnError, "exit-on-error", false, "-exit-on-error exit on error")
	flag.BoolVar(&cfg.screenshotOnError, "screenshot-on-error", false, "-screenshot-on-error take a screenshot on error")
	flag.StringVar(&cfg.screenshotPath, "screenshot-path", "", "-screenshot-path path to save the screenshot")
	flag.StringVar(&cfg.screenshotS3Bucket, "screenshot-s3-bucket", "", "-screenshot-s3-bucket s3 bucket to save the screenshot")
	flag.StringVar(&cfg.screenshotS3Endpoint, "screenshot-s3-endpoint", "", "-screenshot-s3-endpoint s3 endpoint to save the screenshot")
	flag.StringVar(&cfg.screenshotS3Region, "screenshot-s3-region", "", "-screenshot-s3-region s3 region to save the screenshot")
	flag.StringVar(&cfg.screenshotS3AccessKey, "screenshot-s3-access-key", "", "-screenshot-s3-access-key s3 access key to save the screenshot")
	flag.StringVar(&cfg.screenshotS3SecretKey, "screenshot-s3-secret-key", "", "-screenshot-s3-secret-key s3 secret key to save the screenshot")
	flag.StringVar(&cfg.screenshotS3Prefix, "screenshot-s3-prefix", "", "-screenshot-s3-prefix s3 prefix to save the screenshot")
	flag.BoolVar(&cfg.screenshotS3TLS, "screenshot-s3-tls", true, "-screenshot-s3-tls s3 tls to save the screenshot")
	flag.IntVar(&cfg.maxScreenshotExecutor, "max-screenshot-executor", 1, "-max-screenshot-executor max screenshot executor")

	// Exporter mode
	flag.IntVar(&cfg.maxExecutor, "maxExecutor", 1, "-maxExecutor your_max_executor")
	flag.IntVar(&cfg.port, "port", 19090, "-port your_port")
	flag.StringVar(&cfg.buckets, "buckets", "100,200,500,1000,2000,3000,4000,5000,6000,7000,8000,9000,10000,15000,25000,35000,45000,60000", "-buckets your_buckets")
	flag.StringVar(&cfg.confPath, "conf", "", "-conf your_conf_path")

	// Attack mode
	flag.IntVar(&cfg.duration, "duration", 10, "-duration your_duration")
	flag.IntVar(&cfg.workers, "workers", 10, "-workers your_workers")
	flag.StringVar(&cfg.resultFile, "result", "", "-result your_result_file")

	// jaeger config
	flag.StringVar(&cfg.jaegerEndpoint, "jaeger-endpoint", "", "-jaeger-endpoint your_jaeger_endpoint")

	flag.BoolVar(&cfg.quiet, "quiet", false, "-quiet quiet mode")
	flag.Parse()

	var wg sync.WaitGroup

	if cfg.screenshotOnError {
		scenarios.ScreenshotOnError = true
	}

	if cfg.screenshotPath != "" {
		scenarios.ScreenshotPath = cfg.screenshotPath
	}

	scenarios.ScreenshotS3Bucket = cfg.screenshotS3Bucket
	scenarios.ScreenshotS3Endpoint = cfg.screenshotS3Endpoint
	scenarios.ScreenshotS3Region = cfg.screenshotS3Region
	scenarios.ScreenshotS3AccessKey = cfg.screenshotS3AccessKey
	scenarios.ScreenshotS3SecretKey = cfg.screenshotS3SecretKey
	scenarios.ScreenshotS3Prefix = cfg.screenshotS3Prefix

	ctx := baggage.ContextWithoutBaggage(context.Background())

	// disable logging if quiet mode is enabled
	if cfg.quiet {
		log.Println("Hidra has been started in quiet mode")
		log.SetOutput(ioutil.Discard)
	}

	// start screenshot worker if needed
	scenarios.CreateScreenshotWorker(ctx, cfg.maxScreenshotExecutor)

	// initialize otel
	otel.StartOtel(cfg.jaegerEndpoint)

	if testMode {
		wg.Add(1)
		go runTestMode(ctx, &cfg, &wg)
	}

	if exporter {
		wg.Add(1)
		go runExporter(ctx, &cfg, &wg)
	}

	if syntaxMode {
		wg.Add(1)
		go runSyntaxMode(&cfg, &wg)
	}

	if attackMode {
		wg.Add(1)
		go runAttackMode(ctx, &cfg, &wg)
	}

	wg.Wait()
}
