package scenarios

import (
	"context"
	"fmt"

	"github.com/hidracloud/hidra/src/models"
	"github.com/hidracloud/hidra/src/utils"
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
