package main

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	const uploadLimit = 1 << 30
	r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid IDL", err)
		return
	}

	dbVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not find video", err)
		return
	}
	if dbVideo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "no authorized to update video", err)
		return
	}

	videoFile, videoHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not load file", err)
		return
	}
	defer videoFile.Close()

	videoType, fileExt, err := validateFileType(videoHeader, "video/mp4")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unsupported file type", err)
		return
	}

	osFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create temporary file", err)
		return
	}
	defer os.Remove(osFile.Name())
	defer osFile.Close()

	_, err = io.Copy(osFile, videoFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not copy file", err)
		return
	}

	processedFile, err := processVideoForFastStart(osFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not process file for fast start", err)
		return
	}

	processedOsFile, err := os.Open(processedFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not open processed file", err)
		return
	}
	defer os.Remove(processedOsFile.Name())
	defer processedOsFile.Close()

	_, err = processedOsFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not reset file", err)
		return
	}

	fileName, err := randomFilename()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not generate file name", err)
		return
	}
	fileName += fileExt

	videoAspectRatio, err := getVideoAspectRatio(processedOsFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get video aspect ratio", err)
		return
	}

	var videoFrame string
	switch videoAspectRatio {
	case "16:9":
		videoFrame = "landscape"
	case "9:16":
		videoFrame = "portrait"
	default:
		videoFrame = "other"
	}

	fileKey := fmt.Sprintf("%s/%s", videoFrame, fileName)

	_, err = cfg.s3Client.PutObject(
		r.Context(),
		&s3.PutObjectInput{
			Bucket:      &cfg.s3Bucket,
			Key:         &fileKey,
			Body:        processedOsFile,
			ContentType: &videoType,
		},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not upload file", err)
		return
	}

	url := cfg.getAssetAWSURL(fileKey)
	dbVideo.VideoURL = &url

	err = cfg.db.UpdateVideo(dbVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVideo)

}

func validateFileType(fileHeader *multipart.FileHeader, targetFileType string) (contentType string, fileExtension string, err error) {
	contentType = fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		return "", "", fmt.Errorf("Missing Content-Type: %v", contentType)
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", "", err
	}
	if mediaType != targetFileType {
		return "", "", fmt.Errorf("Unsupported file type: Expected: %v, Got: %v", targetFileType, mediaType)
	}
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return "", "", fmt.Errorf("could not get file extension from content type: %w", err)
	}
	if len(extensions) < 1 {
		return "", "", fmt.Errorf("no file extension found in content type: %v", contentType)
	}
	fileExtension = extensions[0]
	return mediaType, fileExtension, nil
}
