package main

import (
	"log"
	"os"
	"path/filepath"
)

func main() {
	inputVideo := "input.mp4"
	outputDir := "./output"

	log.Println("Starting transcoding ....")

	if _, err := os.Stat(inputVideo); os.IsNotExist(err) {
		log.Fatalf("Missing video named 'input.mp4' ")
	}

	if err := transcodeAllResolutions(inputVideo, outputDir); err != nil {
		log.Fatalf("Transcode failed. %v", err)
	}

	masterManifestPath := filepath.Join(outputDir, "master.m3u8")

	if err := createMasterPlaylist(masterManifestPath); err != nil {
		log.Fatalf("Master playlist mapping failed: %v", err)
	}

	log.Print("Tanscoding sucessfull.")

}
