package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	// inputVideo := "input.mp4"
	// outputDir := "./output"

	// log.Println("Starting transcoding ....")

	// if _, err := os.Stat(inputVideo); os.IsNotExist(err) {
	// 	log.Fatalf("Missing video named 'input.mp4' ")
	// }

	// if err := transcodeAllResolutions(inputVideo, outputDir); err != nil {
	// 	log.Fatalf("Transcode failed. %v", err)
	// }

	// masterManifestPath := filepath.Join(outputDir, "master.m3u8")

	// if err := createMasterPlaylist(masterManifestPath); err != nil {
	// 	log.Fatalf("Master playlist mapping failed: %v", err)
	// }

	// log.Print("Tanscoding sucessfull.")

	ctx := context.Background()

	sourceFileKey := "input.mp4"
	// raw/test/input.mp4

	targetPrefix := "test"
	//processed/test

	log.Println("HLS worker with cloudflare r2 integration")

	cfgR2, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion("auto"),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			getRequiredEnv("R2_ACCESS_KEY"), getRequiredEnv("R2_SECRET_KEY"), "",
		)),
	)

	if err != nil {
		log.Fatalf("Configuration error : %v", err)
	}

	r2Client := s3.NewFromConfig(cfgR2, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(getRequiredEnv("R2_ENDPOINT"))
		o.UsePathStyle = true
	})

	rawBucket := getRequiredEnv("RAW_BUCKET")
	processedBucket := getRequiredEnv("PROCESSED_BUCKET")

	scratchPath := filepath.Join("/tmp", "hls-scratch")
	_ = os.MkdirAll(scratchPath, 0755)
	defer os.RemoveAll(scratchPath)

	localRawFile := filepath.Join(scratchPath, "raw_video.mp4")
	localOutpurDir := filepath.Join(scratchPath, "output")

	if err := downloadFromR2(ctx, r2Client, rawBucket, sourceFileKey, localRawFile); err != nil {
		log.Fatalf("failed at downloading phase %v", err)
	}

	if err := transcodeAllResolutions(localRawFile, localOutpurDir); err != nil {
		log.Fatalf("failed at encoding phase %v", err)
	}

	_ = createMasterPlaylist(filepath.Join(localOutpurDir, "master.m3u8"))

	if err := uploadToR2(ctx, r2Client, scratchPath, processedBucket, targetPrefix); err != nil {
		log.Fatalf("failed at uploading phase %v", err)
	}

	log.Print("The encoding is successfull")

}

func getRequiredEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("env variable [%s] is missing.", key)
	}
	return val
}
