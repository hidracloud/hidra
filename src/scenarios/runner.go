package scenarios

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hidracloud/hidra/src/models"
	"github.com/hidracloud/hidra/src/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// SCREENSHOT_ON_ERROR if true, generate screenshot on error
var SCREENSHOT_ON_ERROR = false

// SCREENSHOT_PATH path to save screenshots
var SCREENSHOT_PATH = "./screenshots"

// SCREENSHOT_BUCKET bucket name to save screenshots
var SCREENSHOT_S3_BUCKET = ""

// SCREENSHOT_S3_ENDPOINT_URL endpoint url to save screenshots
var SCREENSHOT_S3_ENDPOINT = ""

// SCREENSHOT_S3_REGION region to save screenshots
var SCREENSHOT_S3_REGION = ""

// SCREENSHOT_S3_ACCESS_KEY access key id to save screenshots
var SCREENSHOT_S3_ACCESS_KEY = ""

// SCREENSHOT_S3_SECRET_KEY secret access key id to save screenshots
var SCREENSHOT_S3_SECRET_KEY = ""

// SCREENSHOT_S3_PREFIX prefix to save screenshots
var SCREENSHOT_S3_PREFIX = ""

// SCREENSHOT_S3_TLS
var SCREENSHOT_S3_TLS = true

// InitializeScenario initialize a new scenario
func InitializeScenario(s models.Scenario) models.IScenario {
	srunner := Sample[s.Kind]()
	srunner.Init()

	return srunner
}

// uploadScreenshots upload screenshots to S3
func uploadScreenshots(src, dest string) error {
	if SCREENSHOT_S3_BUCKET == "" {
		return nil
	}

	// Initialize minio client object.
	minioClient, err := minio.New(SCREENSHOT_S3_ENDPOINT, &minio.Options{
		Creds:  credentials.NewStaticV4(SCREENSHOT_S3_ACCESS_KEY, SCREENSHOT_S3_SECRET_KEY, ""),
		Secure: SCREENSHOT_S3_TLS,
	})
	if err != nil {
		log.Fatalln(err)
	}

	_, err = minioClient.FPutObject(context.Background(), SCREENSHOT_S3_BUCKET, dest, src, minio.PutObjectOptions{
		ContentType: "image/png",
	})

	if err != nil {
		return err
	}

	return nil
}

// GenerateScreenshots generate screenshots for scenario
func generateScreenshots(m *models.ScenarioResult, s models.Scenario, name, desc string) error {
	if !SCREENSHOT_ON_ERROR || m.Error == nil || (s.Kind != "http") {
		return nil
	}

	// Calculate request step
	url := ""
	for _, step := range s.Steps {
		if step.Type == "request" {
			url = step.Params["url"]
			break
		}
	}

	if url == "" {
		return fmt.Errorf("no request step found")
	}

	path := fmt.Sprintf("%s/hidra-screenshot-%s.png", SCREENSHOT_PATH, name)
	err := utils.TakeScreenshotWithChromedp(url, path)
	if err != nil {
		return err
	}

	err = uploadScreenshots(path, fmt.Sprintf("%s%s.png", SCREENSHOT_S3_PREFIX, name))
	if err != nil {
		return err
	}

	return nil
}

// RunIScenario run already initialize scenario
func RunIScenario(name, desc string, s models.Scenario, srunner models.IScenario) *models.ScenarioResult {
	metric := models.ScenarioResult{}
	metric.Scenario = s
	metric.StepResults = make([]*models.StepResult, 0)
	metric.StartDate = time.Now()

	defer srunner.Close()

	for _, step := range s.Steps {
		smetric := models.StepResult{}
		smetric.Step = step
		smetric.StartDate = time.Now()
		customMetrics, err := srunner.RunStep(step.Type, step.Params, step.Timeout)

		smetric.Metrics = customMetrics
		smetric.EndDate = time.Now()
		metric.StepResults = append(metric.StepResults, &smetric)

		if step.Negate && err == nil {
			metric.Error = fmt.Errorf("expected fail")
			metric.EndDate = time.Now()
			generateScreenshots(&metric, s, name, desc)

			return &metric
		}

		if err != nil && !step.Negate {
			metric.Error = err
			metric.EndDate = time.Now()
			generateScreenshots(&metric, s, name, desc)

			return &metric
		}
	}

	metric.EndDate = time.Now()
	return &metric
}

// RunScenario Run one scenario
func RunScenario(s models.Scenario, name, desc string) *models.ScenarioResult {
	utils.LogDebug("[%s] Running new scenario, \"%s\"\n", name, desc)
	srunner := InitializeScenario(s)
	return RunIScenario(name, desc, s, srunner)
}

// PrettyPrintScenarioResults Print scenario metrics
func PrettyPrintScenarioResults(m *models.ScenarioResult, name, desc string) {
	log.Printf("[%s] Metrics results for: %s\n", name, desc)
	if m.Error == nil {
		log.Printf("[%s] Scenario ran without issues\n", name)
	} else {
		log.Printf("[%s] Scenario ran with issues: %s\n", name, m.Error)

	}
	log.Printf("[%s] Total scenario duration: %d (ms)\n", name, m.EndDate.Sub(m.StartDate).Milliseconds())

	for _, s := range m.StepResults {
		log.Printf("[%s]   |_ %s duration: %d (ms)\n", name, s.Step.Type, s.EndDate.Sub(s.StartDate).Milliseconds())
	}
}

// PrettyPrintScenarioResults2String Print scenario metrics
func PrettyPrintScenarioResults2String(m *models.ScenarioResult, name, desc string) string {
	out := ""

	out += fmt.Sprintf("[%s] Metrics results for: %s\n", name, desc)
	if m.Error == nil {
		out += fmt.Sprintf("[%s] Scenario ran without issues\n", name)
	} else {
		out += fmt.Sprintf("[%s] Scenario ran with issues: %s\n", name, m.Error)

	}

	out += fmt.Sprintf("[%s] Total scenario duration: %d (ms)\n", name, m.EndDate.Sub(m.StartDate).Milliseconds())

	for _, s := range m.StepResults {
		out += fmt.Sprintf("[%s]   |_ %s duration: %d (ms)\n", name, s.Step.Type, s.EndDate.Sub(s.StartDate).Milliseconds())
	}

	out += fmt.Sprintf("[%s] Total scenario duration: %d (ms)\n", name, m.EndDate.Sub(m.StartDate).Milliseconds())
	return out
}
