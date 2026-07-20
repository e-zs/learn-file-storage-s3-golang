package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime"
	"os"
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
