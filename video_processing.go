package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

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

	type probeOutput struct {
		Streams []struct {
			CodecType          string `json:"codec_type"`
			Width              int    `json:"width"`
			Height             int    `json:"height"`
			DisplayAspectRatio string `json:"display_aspect_ratio"`
		} `json:"streams"`
	}

	var probe probeOutput

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

func processVideoForFastStart(filePath string) (string, error) {
	outputFile := filePath + ".processing"

	cmd := exec.Command(
		"ffmpeg",
		"-i",
		filePath,
		"-c",
		"copy",
		"-movflags",
		"faststart",
		"-f",
		"mp4",
		outputFile,
	)

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return outputFile, nil
}
