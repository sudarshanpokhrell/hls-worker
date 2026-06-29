package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type Resolution struct {
	Name         string
	Width        int
	Height       int
	VideoBitrate string
	AudioBitrate string
	Bandwidth    int
}

var resolutions = []Resolution{
	{Name: "1080p", Width: 1920, Height: 1080, VideoBitrate: "5000k", AudioBitrate: "192k", Bandwidth: 5192000},
	{Name: "720p", Width: 1280, Height: 720, VideoBitrate: "2800k", AudioBitrate: "128k", Bandwidth: 2928000},
	{Name: "480p", Width: 854, Height: 480, VideoBitrate: "1400k", AudioBitrate: "128k", Bandwidth: 1528000},
}

func transcodeAllResolutions(inputPath, workDir string) error {
	var wg sync.WaitGroup

	errChan := make(chan error, len(resolutions))

	for _, res := range resolutions {
		wg.Add(1)

		go func(r Resolution) {
			defer wg.Done()

			log.Printf("[Transcoder] Launching encoding thread for: %s", r.Name)

			outDir := filepath.Join(workDir, r.Name)

			if err := os.MkdirAll(outDir, 0755); err != nil {
				errChan <- fmt.Errorf("%s directory creation failed: %w", r.Name, err)
				return
			}

			//tanscode to HLS

			if err := transcodeToHLS(inputPath, outDir, r); err != nil {
				errChan <- fmt.Errorf("%s transcode failed: %w", r.Name, err)
			}
		}(res)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		var combinedErr error
		for err := range errChan {
			combinedErr = fmt.Errorf("%v | %w", combinedErr, err)
		}
		return combinedErr
	}

	return nil
}

func transcodeToHLS(inputPath, outputDir string, res Resolution) error {
	playlistPath := filepath.Join(outputDir, "index.m3u8")
	segmentPath := filepath.Join(outputDir, "segment_%03d.ts")

	scaleFilter := fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black",
		res.Width, res.Height, res.Width, res.Height,
	)

	args := []string{
		"-y", "-i", inputPath,
		"-c:v", "libx264", "-preset", "fast", "-profile:v", "main",
		"-vf", scaleFilter, "-b:v", res.VideoBitrate, "-maxrate", res.VideoBitrate, "-bufsize", "10000k",
		"-c:a", "aac", "-b:a", res.AudioBitrate, "-ar", "44100",
		"-f", "hls", "-hls_time", "6", "-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments", "-hls_segment_type", "mpegts",
		"-hls_segment_filename", segmentPath, playlistPath,
	}

	cmd := exec.Command("ffmpeg", args...)

	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func createMasterPlaylist(outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n\n")
	displayDimensions := map[string]string{"1080p": "1920x1080", "720p": "1280x720", "480p": "854x480"}

	for _, res := range resolutions {
		fmt.Fprintf(f, "#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s,NAME=\"%s\"\n%s/index.m3u8\n\n",
			res.Bandwidth, displayDimensions[res.Name], res.Name, res.Name)
	}
	return nil
}
