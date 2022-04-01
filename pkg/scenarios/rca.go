package scenarios

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hidracloud/hidra/v2/pkg/models"
	"github.com/hidracloud/hidra/v2/pkg/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ScreenshotOnError if true, generate screenshot on error
var ScreenshotOnError = false

// ScreenshotPath path to save screenshots
var ScreenshotPath = "./screenshots"

// ScreenshotS3Bucket bucket name to save screenshots
var ScreenshotS3Bucket = ""

// ScreenshotS3Endpoint endpoint url to save screenshots
var ScreenshotS3Endpoint = ""

// ScreenshotS3Region region to save screenshots
var ScreenshotS3Region = ""

// ScreenshotS3AccessKey access key id to save screenshots
var ScreenshotS3AccessKey = ""

// ScreenshotS3SecretKey secret access key id to save screenshots
var ScreenshotS3SecretKey = ""

// ScreenshotS3Prefix prefix to save screenshots
var ScreenshotS3Prefix = ""

// ScreenshotS3TLS if true, use TLS to connect to S3
var ScreenshotS3TLS = true

// screenshotQueueItem represent a item of screenshots to generate
type screenshotQueueItem struct {
	Result   *models.ScenarioResult
	Scenario models.Scenario
	Name     string
	Desc     string
}

// screenshotQueue represent a queue of screenshots to generate
var screenshotQueue chan *screenshotQueueItem

// uploadScreenshots upload screenshots to S3
func uploadScreenshots(src, dest string) error {
	if ScreenshotS3Bucket == "" {
		return nil
	}

	// Initialize minio client object.
	minioClient, err := minio.New(ScreenshotS3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(ScreenshotS3AccessKey, ScreenshotS3SecretKey, ""),
		Secure: ScreenshotS3TLS,
	})
	if err != nil {
		return err
	}

	_, err = minioClient.FPutObject(context.Background(), ScreenshotS3Bucket, dest, src, minio.PutObjectOptions{
		ContentType: "image/png",
	})

	if err != nil {
		return err
	}

	return nil
}

// GenerateScreenshots generate screenshots for scenario
func GenerateScreenshots(m *models.ScenarioResult, s models.Scenario, name, desc string) error {
	if !ScreenshotOnError || m.Error == nil || (s.Kind != "http") {
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

	path := fmt.Sprintf("%s/hidra-screenshot-%s.png", ScreenshotPath, name)
	err := utils.TakeScreenshotWithChromedp(url, path)
	if err != nil {
		return err
	}

	err = uploadScreenshots(path, fmt.Sprintf("%s%s.png", ScreenshotS3Prefix, name))
	if err != nil {
		return err
	}

	return nil
}

// AddScreenshotsToQueue add screenshots to queue
func AddScreenshotsToQueue(m *models.ScenarioResult, s models.Scenario, name, desc string) {
	if !ScreenshotOnError || m.Error == nil || (s.Kind != "http") {
		return
	}

	screenshotQueue <- &screenshotQueueItem{
		Result:   m,
		Scenario: s,
		Name:     name,
		Desc:     desc,
	}
}

// CreateScreenshotWorker create a new worker to generate screenshots
func CreateScreenshotWorker(ctx context.Context, maxExecutor int) {
	if !ScreenshotOnError {
		return
	}

	screenshotQueue = make(chan *screenshotQueueItem, maxExecutor)
	for i := 0; i < maxExecutor; i++ {
		go func(workerID int) {
			log.Println("Initializing screenshot worker", workerID)

			for {
				screenshotQueueItem := <-screenshotQueue
				log.Println("["+strconv.Itoa(workerID)+"] Generating screenshot", screenshotQueueItem.Name)

				err := GenerateScreenshots(screenshotQueueItem.Result, screenshotQueueItem.Scenario, screenshotQueueItem.Name, screenshotQueueItem.Desc)
				if err != nil {
					log.Println("["+strconv.Itoa(workerID)+"] Error generating screenshot", err)
				}

				log.Println("["+strconv.Itoa(workerID)+"] Screenshot generated", screenshotQueueItem.Name)
			}
		}(i)
	}
}
