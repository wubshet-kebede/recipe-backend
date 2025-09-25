package fileupload

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Request structs
type ImageInput struct {
	Base64 string `json:"base64"`
}

type UploadProfilePictureInput struct {
	File   ImageInput `json:"file"`
	Folder string    `json:"folder"`
}

type ProfileInput struct {
	Input UploadProfilePictureInput `json:"input"`
}

type HasuraImageActionRequest struct {
	Action           map[string]interface{} `json:"action"`
	Input            ProfileInput           `json:"input"`
	RequestQuery     string                 `json:"request_query"`
	SessionVariables map[string]string      `json:"session_variables"`
}

// Response
type UploadProfilePictureOutput struct {
	Success           bool   `json:"success"`
	ProfilePictureUrl string `json:"profilePictureUrl"`
}

func UploadProfilePicHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("UploadProfilePicHandler called")
	w.Header().Set("Content-Type", "application/json")

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8082"
		}
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("Error reading body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Could not read request body"})
		return
	}

	var hasuraReq HasuraImageActionRequest
	err = json.Unmarshal(body, &hasuraReq)
	if err != nil {
		log.Printf("Hasura Action parse error: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid Hasura Action request"})
		return
	}

	req := hasuraReq.Input.Input
	file := req.File

	// Folder for profile pictures
	folder := req.Folder
	if folder == "" {
		folder = "/app/uploads/profile_pics"
	}
	err = os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		log.Printf("Folder creation error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Could not create folder"})
		return
	}

	// Decode base64
	commaIndex := strings.Index(file.Base64, ",")
	if commaIndex == -1 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid base64 string"})
		return
	}
	rawBase64 := file.Base64[commaIndex+1:]
	data, err := base64.StdEncoding.DecodeString(rawBase64)
	if err != nil {
		log.Printf("Base64 decode error: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Base64 decode failed"})
		return
	}

	// Determine extension
	ext := ".png"
	if strings.Contains(file.Base64, "image/jpeg") || strings.Contains(file.Base64, "image/jpg") {
		ext = ".jpg"
	} else if strings.Contains(file.Base64, "image/gif") {
		ext = ".gif"
	}

	// Filename based on user ID (overwrite existing)
	userID := hasuraReq.SessionVariables["x-hasura-user-id"]
	if userID == "" {
		userID = uuid.New().String() // fallback
	}
	filename := fmt.Sprintf("profile_%s%s", userID, ext)
	filePath := filepath.Join(folder, filename)

	// Save file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Printf("File write error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Could not save file"})
		return
	}

	fullURL := fmt.Sprintf("%s/%s", baseURL, filepath.ToSlash(filePath))

	// Return response
	response := UploadProfilePictureOutput{
		Success:           true,
		ProfilePictureUrl: fullURL,
	}

	json.NewEncoder(w).Encode(response)
}
