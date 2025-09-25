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
	"time"

	"github.com/google/uuid"
)
type Files struct {
	Base64 string `json:"base64"`
}
type UploadFilesRequest struct {
	Files []Files `json:"files"`
	Folder string `json:"folder"`
}
type OuterInput struct {
	Input UploadFilesRequest `json:"input"`
}
type HasuraActionRequest struct {
	Action map[string]interface{} `json:"action"`
	Input OuterInput    `json:"input"`
	RequestQuery string `json:"request_query"`
	SessionVariables map[string]string `json:"session_variables"`
}
type UploadFilesOutput struct {
	Success bool `json:"success"`
	Urls []string `json:"urls"`
}
func UploadFilesHandler(w http.ResponseWriter, r*http.Request){
	log.Println("UploadFilesHandler called")
	w.Header().Set("Content-Type", "application/json")
	baseURL := os.Getenv("BASE_URL")
	if baseURL == ""{
		port:= os.Getenv("PORT")
		if port == ""{
			port = "8082"
		}
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("Error reading body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Could not read request body"})
		return
	}
	log.Printf("Raw request body: %s\n", string(body))
   var hasuraReq HasuraActionRequest
	err = json.Unmarshal(body, &hasuraReq)
	if err != nil {
		log.Printf("Hasura Action parse error: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid Hasura Action request"})
		return
	}
	req := hasuraReq.Input.Input

	log.Printf("Received files: %d, Folder: %s\n", len(req.Files), req.Folder)
 folder := req.Folder
 if folder == ""{
	folder = "/app/uploads"
 }
 log.Println("Attempting to create folder:", folder)
	err = os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		log.Printf("Folder creation error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Could not create folder"})
		return
	}
	log.Println("Folder creation successful.")
	var savedURLs []string
	var success bool = true
	for i, file := range req.Files {
		log.Printf("Processing file %d, base64 length: %d\n", i, len(file.Base64))
  commaIndex := strings.Index(file.Base64, ",")
		if commaIndex == -1 {
			log.Printf("Base64 string for file %d missing comma separator: %s\n", i, file.Base64)
			success = false
			continue
		}
	rawBase64 := file.Base64[commaIndex+1:]

		log.Println("Attempting base64 decode for file:", i)
		data, err := base64.StdEncoding.DecodeString(rawBase64)
		if err != nil {
			log.Printf("Base64 decode error for file %d: %v\n", i, err)
			success = false
			continue
		}
	log.Println("Base64 decode successful for file:", i, "decoded size:", len(data), "bytes")


		ext := ".png"
		if strings.Contains(file.Base64, "image/jpeg") {
			ext = ".jpeg"
		} else if strings.Contains(file.Base64, "image/jpg") {
			ext = ".jpg"
		} else if strings.Contains(file.Base64, "image/gif") {
			ext = ".gif"
		}
		
		filename := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102150405"), uuid.New().String(), ext)
		filePath := filepath.Join(folder, filename)

		log.Println("Attempting to write file:", filePath)
		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			log.Printf("File write error for file %d: %v\n", i, err)
			success = false
			continue
		}
		log.Println("File write successful for file:", filePath)

		urlPath := filepath.ToSlash(filePath)
		fullURL := fmt.Sprintf("%s/%s", baseURL, urlPath)
		savedURLs = append(savedURLs, fullURL)
	}	
	response := UploadFilesOutput{
		Success: success,
		Urls:    savedURLs,
	}

	log.Printf("Returning response: %+v\n", response)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v\n", err)
	}	
}