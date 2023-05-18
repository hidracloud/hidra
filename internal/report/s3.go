package report

import (
	"bytes"
	"context"
	"errors"
	"mime"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	log "github.com/sirupsen/logrus"
)

// SaveS3 saves the report to S3.
func (r *Report) SaveS3() error {
	if ReportS3Conf == nil {
		return nil
	}

	log.Debug("Saving report to S3")

	minioClient, err := minio.New(ReportS3Conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(ReportS3Conf.AccessKeyID, ReportS3Conf.SecretAccessKey, ""),
		Secure: ReportS3Conf.UseSSL,
		Region: ReportS3Conf.Region,
	})

	if err != nil {
		return err
	}

	rDump := r.Dump()

	// rDump to reader
	reader := strings.NewReader(rDump)

	if reader == nil {
		return errors.New("reader is nil")
	}

	_, err = minioClient.PutObject(context.Background(), ReportS3Conf.Bucket, r.Name+".json", reader, int64(len(rDump)), minio.PutObjectOptions{
		ContentType: "application/json",
	})

	// upload attachments to r.Name.more/ folder
	for dest, content := range r.Attachments {
		// create a reader from origin file
		reader := bytes.NewReader(content)

		contentType := "application/octet-stream"

		// get content type from file extension
		if ext := filepath.Ext(dest); ext != "" {
			if ct := mime.TypeByExtension(ext); ct != "" {
				contentType = ct
				// remove encoding from content type
				if i := strings.Index(contentType, ";"); i != -1 {
					contentType = contentType[:i]
				}
			}
		}

		_, err = minioClient.PutObject(context.Background(), ReportS3Conf.Bucket, r.Name+".more/"+dest, reader, int64(len(content)), minio.PutObjectOptions{
			ContentType: contentType,
		})

		if err != nil {
			log.Warnf("Failed to upload attachment to S3: %s", err)
			continue
		}
	}

	if len(r.Attachments) > 0 {
		indexHTML := r.GenerateMoreIndexHTML()

		_, err = minioClient.PutObject(context.Background(), ReportS3Conf.Bucket, r.Name+".more/index.html", strings.NewReader(indexHTML), int64(len(indexHTML)), minio.PutObjectOptions{
			ContentType: "text/html",
		})

		if err != nil {
			log.Warnf("Failed to upload index.html to S3: %s", err)
		}
	}
	return err
}
