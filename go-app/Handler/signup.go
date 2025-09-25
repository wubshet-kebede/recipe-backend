package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	hasura "github.com/wubshet-kebede/go-app/Hasura"
	"golang.org/x/crypto/bcrypt"
)
type SignupRequest struct {
    Username    string `json:"username"`
    Email       string `json:"email"`
    PasswordHash string `json:"password_hash"`
    FirstName   string `json:"first_name"`
    MiddleName  string `json:"middle_name"`
    LastName    string `json:"last_name"`
    PhoneNumber string `json:"phone_number"`
}


type SignupOutput struct {
	ID uuid.UUID `json:"id"`
	Username string `json:"username"`
	Email string `json:"email"`
	Message string `json:"message"`
}
type HasuraActionPayload struct {
	Action struct {
		Name string `json:"name"`
	} `json:"action"`
	Input struct {
		Input SignupRequest `json:"input"`
	} `json:"input"`
	SessionVariables map[string]string `json:"session_variables"`
	RequestQuery string `json:"request_query"`
}
func SignupHandler (w http.ResponseWriter, r*http.Request){
	log.Println("Received request at /signUp")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Printf("Method not allowed : %s", r.Method)
		return
	}
	var payload HasuraActionPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		log.Printf("Error decoding Hasura action payload: %v", err)
		http.Error(w, "Invalid request body format", http.StatusBadRequest)
		return
	}
	req := payload.Input.Input
	log.Printf("Parsed SignUp Ipnut: Username =%s, Email=%s, Password Length=%d", req.Username, req.Email, len(req. PasswordHash))
	if req.Username == "" || req.Email == "" || req. PasswordHash == "" || req.FirstName == "" || req.LastName == "" || req.PhoneNumber == "" {
		log.Println("Validation error: All fields are required from parsed input")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) 
		json.NewEncoder(w).Encode(map[string]string{"message": "All fields are required"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req. PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password for user %s: %v", req.Username, err)
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	var insertUserMutation struct {
    InsertUsersOne struct {
        ID          uuid.UUID `graphql:"id"`
        Username    string    `graphql:"username"`
        Email       string    `graphql:"email"`
        FirstName   string    `graphql:"first_name"`
		MiddleName  string     `graphql:"middle_name"`
        LastName    string    `graphql:"last_name"`
        PhoneNumber string    `graphql:"phone_number"`
    } `graphql:"insert_users_one(object: {username: $username, email: $email, password_hash: $password_hash, first_name: $firstName,middle_name:$middleName last_name: $lastName, phone_number: $phoneNumber})"`
}
variables := map[string]interface{}{
		"username":      req.Username,
		"email":         req.Email,
		"password_hash": string(hashedPassword),
		"firstName":     req.FirstName,
		"middleName":     req.MiddleName,
		"lastName":      req.LastName,
		"phoneNumber":   req.PhoneNumber,
	}
log.Printf("Variables sent to Hasura: %+v", variables)

	// Execute the mutation against Hasura
	if err := hasura.Client.Mutate(context.Background(), &insertUserMutation, variables); err != nil {
		log.Printf("Hasura mutation error for user %s: %v", req.Username, err)

		if err.Error() == "duplicate key value violates unique constraint \"users_email_key\"" {
			http.Error(w, "Email already exists", http.StatusConflict) // 409 Conflict
			return
		}
		if err.Error() == "duplicate key value violates unique constraint \"users_username_key\"" {
			http.Error(w, "Username already exists", http.StatusConflict) // 409 Conflict
			return
		}
		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
		return
	}
   resp := SignupOutput{
		ID:   insertUserMutation.InsertUsersOne.ID, 
		Username: insertUserMutation.InsertUsersOne.Username,
		Email:    insertUserMutation.InsertUsersOne.Email,
		Message:  "Signup successful!", 
	}

	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) 

	// Encode and send the JSON response
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s (%s) registered successfully with ID: %s", resp.Username, resp.Email, resp.ID)
}