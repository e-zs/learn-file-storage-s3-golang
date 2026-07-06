package main

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(videoId uuid.UUID, contentType string) (string, error) {
	fileExtension, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return "", err
	}
	if len(fileExtension) == 0 {
		return "", fmt.Errorf("no extension found for type %s", contentType)
	}
	return fmt.Sprintf("%v%v", videoId, fileExtension[0]), nil
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%v/assets/%v", cfg.port, assetPath)
}
