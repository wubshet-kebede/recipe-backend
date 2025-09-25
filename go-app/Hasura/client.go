package hasura

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/hasura/go-graphql-client"
)

var Client *graphql.Client
type loggingRoundTripper struct {
	next http.RoundTripper 
}
func (lrt *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("ERROR: Failed to read request body for logging: %v", err)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) 
			return lrt.next.RoundTrip(req)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) 

		var raw map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &raw); err == nil {
			prettyJSON, _ := json.MarshalIndent(raw, "", "  ")
			log.Printf("DEBUG: Outgoing GraphQL Request Body:\n%s", prettyJSON)
		} else {
			log.Printf("DEBUG: Outgoing GraphQL Request Body (raw):\n%s", string(bodyBytes))
		}
	} else {
		log.Println("DEBUG: Outgoing GraphQL Request Body: (empty)")
	}

	return lrt.next.RoundTrip(req)
}


type AuthTransport struct {
	Headers             map[string]string
	UnderlyingTransport http.RoundTripper 
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range t.Headers {
		req.Header.Add(key, value)
	}
	
	return t.UnderlyingTransport.RoundTrip(req)
}


func InitClient() {
	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	adminSecret := os.Getenv("HASURA_ADMIN_SECRET")
	fmt.Println("Go backend connecting to Hasura at:", hasuraURL)
	fmt.Println("Go backend using admin secret:", adminSecret)

	if hasuraURL == "" {
		hasuraURL = "http://graphql-engine:8080/v1/graphql" 
	}

	
	finalTransport := &AuthTransport{
		Headers: map[string]string{
			"X-Hasura-Admin-Secret": adminSecret,
		},
		
		UnderlyingTransport: &loggingRoundTripper{
			next: http.DefaultTransport, 
		},
	}

	httpClient := &http.Client{
		Transport: finalTransport, 
	}

	Client = graphql.NewClient(hasuraURL, httpClient)
	
}