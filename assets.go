package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(contentType string) (string, error) {
	fileExtension, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return "", err
	}
	if len(fileExtension) == 0 {
		return "", fmt.Errorf("no extension found for type %s", contentType)
	}

	fileName, err := randomFilename()
	if err != nil {
		return "", fmt.Errorf("could not create filename: %w", err)
	}

	return fmt.Sprintf("%s%s", fileName, fileExtension[0]), nil
}

func randomFilename() (string, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%v/assets/%v", cfg.port, assetPath)
}

func (cfg apiConfig) getAssetAWSURL(assetKey string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, assetKey)
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-print_format",
		"json",
		"-show_streams",
		filePath,
	)
	buffy := &bytes.Buffer{}
	cmd.Stdout = buffy

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	type FFProbeOutput struct {
		Streams []struct {
			CodecType          string `json:"codec_type"`
			Width              int    `json:"width"`
			Height             int    `json:"height"`
			DisplayAspectRatio string `json:"display_aspect_ratio"`
		} `json:"streams"`
	}

	var probe FFProbeOutput

	if err := json.Unmarshal(buffy.Bytes(), &probe); err != nil {
		return "", err
	}

	if len(probe.Streams) == 0 {
		return "", fmt.Errorf("no video stream")
	}

	var width int
	var height int

	for i, stream := range probe.Streams {
		if stream.CodecType == "video" {
			width = probe.Streams[i].Width
			height = probe.Streams[i].Height
			break
		}
	}

	return calcAspectRatio(width, height), nil
}

func calcAspectRatio(width int, height int) string {
	const tolerance = 0.02
	ratio := float64(width) / float64(height)

	if math.Abs(ratio-(16.0/9.0)) < tolerance {
		return "16:9"
	} else if math.Abs(ratio-(9.0/16.0)) < tolerance {
		return "9:16"
	} else {
		return "other"
	}
}
