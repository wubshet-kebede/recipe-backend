package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	hasura "github.com/wubshet-kebede/go-app/Hasura"
	"golang.org/x/crypto/bcrypt"
)
	type loginRequest struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type LoginResponse struct {
		Id uuid.UUID `json:"id"`
		Username string `json:"username"`
		Token string `json:"token"`
		Email   string    `json:"email"`
		FirstName string   `json:"first_name"`
	    LastName   string   `json:"last_name"`
	    Message string    `json:"message"`
	}

	type LoginActionPayload struct {
		Action struct {
			Name string `json:"name"`} `json:"action"`
		Input struct {
			Input loginRequest `json:"input"`
		} `json:"input"`
		SessionVariables map[string]string `json:"session_variables"`
		RequestQuery string `json:"request_query"`
		
	}

	type getUserByEmailQuery struct {
		Users []struct {
		Id uuid.UUID `json:"id" graphql: "id"`
		Username    string    `json:"username" graphql:"username"`
        Email       string    `json:"email" graphql:"email"`
        PasswordHash string   `json:"password_hash" graphql:"password_hash"`
		FirstName    string    `json:"first_name" graphql:"first_name"`
		LastName    string    `json:"last_name" graphql:"last_name"`
		}`graphql:"users(where: {email: {_eq: $email}})"`
	}
	var jwtSecret []byte

	func init() {
    err := godotenv.Load()
    if err != nil {
        fmt.Println("Error loading .env file:", err)
    }

    
    jwtSecretString := os.Getenv("HASURA_GRAPHQL_JWT_SECRET")
    fmt.Println("Go Backend JWT_SECRET loaded:", jwtSecretString)

    if len(jwtSecretString) == 0 {
        fmt.Println("WARNING: HASURA_GRAPHQL_JWT_SECRET environment variable is empty or not set!")
        return
    }

    var hasuraSecret struct {
        Key string `json:"key"`
    }

  
    if err := json.Unmarshal([]byte(jwtSecretString), &hasuraSecret); err != nil {
        fmt.Println("Error unmarshalling JWT secret JSON:", err)
        return
    }

    jwtSecret = []byte(hasuraSecret.Key)
    fmt.Println("Go Backend JWT Key:", string(jwtSecret))

    if len(jwtSecret) == 0 {
        fmt.Println("WARNING: Extracted JWT key is empty!")
    }
}

func generateJWT(ID uuid.UUID, username string, email string, FirstName string, LastName string)(string, error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": ID.String(),
		"email": email,
		"username": username,
		"first_name":FirstName,
		"last_name": LastName,
		"metadata": map[string]interface{}{   
            "roles": []string{"user"},
    },
		"https://hasura.io/jwt/claims": map[string]interface{}{
            "x-hasura-allowed-roles": []string{"user", "public"},
            "x-hasura-default-role":  "user",
			"x-hasura-user-id":       ID.String(),
	}, "exp": time.Now().Add(time.Hour * 24).Unix(),})
	tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        return "", fmt.Errorf("failed to sign token: %w", err)
    }
    return tokenString, nil
}
func LoginHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return 
	}
	var actionInput LoginActionPayload
	if err:= json.NewDecoder(r.Body).Decode(&actionInput);
	err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req:= actionInput.Input.Input
	if req.Email == ""|| req.Password==""{
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}
	var q getUserByEmailQuery
	variables := map[string]interface{}{
		"email": req.Email,
	}
	if err:= hasura.Client.Query(context.Background(), &q, variables); err!=nil {
		http.Error(w, "Error querying user ", http.StatusInternalServerError)
		log.Printf("Error querying user %s: %v", req.Email, err)
		return
	}
	if len(q.Users) == 0 {
   http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
   user := q.Users[0]
   err:= bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
     if err!= nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	 }
	 tokenString, err := generateJWT(user.Id, user.Username, user.Email, user.FirstName, user.LastName)
	 if err!= nil {
		log.Printf("Error generating JWT for user %s: %v", user.Username, err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	 }
	 resp:= LoginResponse{
		Id:   user.Id,
		Username: user.Username,
		Email:    user.Email,
		FirstName: user.FirstName,
		LastName: user.LastName,
		Token:    tokenString,
		Message:  "Login successful!",
	 }
	   w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
    log.Printf("User %s (%s) %s, %s logged in successfully. Token: %s", resp.Username, resp.Email, resp.FirstName, resp.LastName, resp.Token)
}