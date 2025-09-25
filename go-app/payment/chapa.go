package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const chapaAPIURL = "https://api.chapa.co/v1/transaction/initialize"
const chapaVerifyAPIURL = "https://api.chapa.co/v1/transaction/verify"

// ChapaService encapsulates all Chapa API logic.
type ChapaService struct {
	secretKey string
	client    *http.Client
}

// NewChapaService creates a new instance of ChapaService.
func NewChapaService() *ChapaService {
	secret := os.Getenv("CHAPA_SECRET_KEY")
	if secret == "" {
		log.Fatal("CHAPA_SECRET_KEY environment variable not set.")
	}
	return &ChapaService{
		secretKey: secret,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// InitiatePayment calls the Chapa API to start a new transaction.
func (s *ChapaService) InitiatePayment(ctx context.Context, reqData ChapaInitiateRequest) (ChapaInitiateResponse, error) {
	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		return ChapaInitiateResponse{}, fmt.Errorf("failed to marshal Chapa request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", chapaAPIURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return ChapaInitiateResponse{}, fmt.Errorf("failed to create Chapa API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.secretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return ChapaInitiateResponse{}, fmt.Errorf("failed to call Chapa API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return ChapaInitiateResponse{}, fmt.Errorf("chapa API returned non-OK status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var chapaResp ChapaInitiateResponse
	if err := json.NewDecoder(resp.Body).Decode(&chapaResp); err != nil {
		return ChapaInitiateResponse{}, fmt.Errorf("failed to decode Chapa API response: %w", err)
	}
	return chapaResp, nil
}

// VerifyPayment calls the Chapa API to verify a transaction's status.
func (s *ChapaService) VerifyPayment(ctx context.Context, txRef string) (ChapaVerifyResponse, error) {
	verifyURL := fmt.Sprintf("%s/%s", chapaVerifyAPIURL, txRef)
	req, err := http.NewRequestWithContext(ctx, "GET", verifyURL, nil)
	if err != nil {
		return ChapaVerifyResponse{}, fmt.Errorf("error creating Chapa verify request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.secretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return ChapaVerifyResponse{}, fmt.Errorf("error calling Chapa verify API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return ChapaVerifyResponse{}, fmt.Errorf("chapa verify API returned non-OK status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var chapaVerifyResp ChapaVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&chapaVerifyResp); err != nil {
		return ChapaVerifyResponse{}, fmt.Errorf("failed to decode Chapa verify response: %w", err)
	}
	return chapaVerifyResp, nil
}

// getFrontendRedirectURL builds the final URL for user redirection.
func getFrontendRedirectURL(returnURL string, status string, orderID string, txRef string, message string) string {
	u, err := url.Parse(returnURL)
	if err != nil {
		log.Printf("Error parsing returnURL: %v. Falling back to base URL.", err)
		return fmt.Sprintf("%s?status=%s&order_id=%s&tx_ref=%s&message=%s", returnURL, status, orderID, txRef, message)
	}
	q := u.Query()
	q.Set("status", status)
	q.Set("order_id", orderID)
	q.Set("tx_ref", txRef)
	if message != "" {
		q.Set("message", message)
	}
	u.RawQuery = q.Encode()
	return u.String()
}