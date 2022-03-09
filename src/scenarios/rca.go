package scenarios

import (
	"context"
	"fmt"

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
		return err
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
func GenerateScreenshots(m *models.ScenarioResult, s models.Scenario, name, desc string) error {
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
