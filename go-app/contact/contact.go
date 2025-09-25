package contact

import (
	//"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	hasura "github.com/wubshet-kebede/go-app/Hasura"
)
func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"message": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Error marshaling JSON response: %v", err) 
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"message": "Internal server error: Failed to marshal response"}`))
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}

// Hasura Action Payload for incoming contact form data
type ContactActionPayload struct {
	Action struct {
		Name string `json:"name"`
	} `json:"action"`
	Input struct {
		Input struct { 
			Name    string `json:"name"`
			Email   string `json:"email"`
			Subject string `json:"subject"`
			Message string `json:"message"`
		} `json:"input"`
	} `json:"input"`
	SessionVariables map[string]string `json:"session_variables"`
}

// Hasura Action Response
type ContactActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	
}
// Go struct to insert into Hasura contact_messages table via the Hasura client
type contact_messages_insert_input struct { // <--- RENAME THIS TYPE
	ID        uuid.UUID `json:"id" graphql:"id"`
	Name      string    `json:"name" graphql:"name"`
	Email     string    `json:"email" graphql:"email"`
	Subject   string    `json:"subject" graphql:"subject"`
	Message   string    `json:"message" graphql:"message"`
	
	CreatedAt time.Time `json:"created_at" graphql:"created_at"` 
	IsRead    bool      `json:"is_read" graphql:"is_read"`       
}
// GraphQL mutation struct for inserting into contact_messages table
type insertContactMessageMutation struct {
    InsertContactMessagesOne *struct {
        ID string `graphql:"id"`
    } `graphql:"insert_contact_messages_one(object: $object)"`
}


// HandleSubmitContactForm is the Go backend endpoint for the Hasura Action
func HandleSubmitContactForm(w http.ResponseWriter, r *http.Request) {
	var payload ContactActionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("ERROR: Failed to decode contact form payload: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	 log.Printf("DEBUG: Contact Form Payload: %+v", payload) 

	ctx := r.Context()
	newContactID := uuid.New() // Generate a new UUID for the message

	// Prepare data for insertion into Hasura's contact_messages table
	 contactMessage := contact_messages_insert_input{ // <--- USE THE RENAMED TYPE HERE
        ID:        newContactID,
        Name:      payload.Input.Input.Name,
        Email:     payload.Input.Input.Email,
        Subject:   payload.Input.Input.Subject,
        Message:   payload.Input.Input.Message,
        CreatedAt: time.Now(),
        IsRead:    false,
    }
	var mutationResp insertContactMessageMutation
	insertVars := map[string]interface{}{
		"object": contactMessage,
	}
log.Println("DEBUG: Attempting to insert contact message into Hasura.") 
	// Insert the message into Hasura
	err := hasura.Client.Mutate(ctx, &mutationResp, insertVars)
	if err != nil {
		log.Printf("ERROR: Failed to insert contact message into Hasura: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to submit contact message")
		return
	}
log.Println("DEBUG: Contact message inserted successfully into Hasura.")
	// You could add logic here to send an email notification, etc.
	log.Printf("Contact message from %s (%s) submitted successfully. ID: %s",
		payload.Input.Input.Name, payload.Input.Input.Email, newContactID.String())

	 responsePayload := ContactActionResponse{
        Success: true,
        Message: "Your message has been sent successfully!",
    }
    log.Printf("DEBUG: Sending JSON response: %+v", responsePayload) // Add this to see the final response struct

    respondWithJSON(w, http.StatusOK, responsePayload)
}

