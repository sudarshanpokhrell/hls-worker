package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func downloadFromR2(ctx context.Context, client *s3.Client, bucket, key, localTarget string) error {
	log.Printf("[Storage] Fetching raw video from R2: %s: ", key)

	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("R2 download request failed: %w", err)
	}

	defer output.Body.Close()

	localFile, err := os.Create(localTarget)

	if err != nil {
		return fmt.Errorf("failed to create file %q: %w: ", localTarget, err)
	}

	defer localFile.Close()

	_, err = io.Copy(localFile, output.Body)
	return err
}

func uploadToR2(ctx context.Context, client *s3.Client, localSourceDir, bucket, targetPrefix string) error {
	log.Printf("[Storage] Uploading video to R2 on prefix: %s", targetPrefix)

	return filepath.Walk(localSourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		relativePath, _ := filepath.Rel(localSourceDir, path)

		cleanObjectKey := targetPrefix + "/" + strings.ReplaceAll(relativePath, "\\", "/")

		fileStream, err := os.Open(path)

		if err != nil {
			return err
		}

		defer fileStream.Close()

		var contentType string

		switch filepath.Ext(path) {
		case ".m3u8":
			contentType = "application/x-mpegURL"
		case ".ts":
			contentType = "video/MP2T"
		default:
			contentType = "application/octet-stream"
		}

		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(cleanObjectKey),
			Body:        fileStream,
			ContentType: aws.String(contentType),
		})

		return err

	})

}
