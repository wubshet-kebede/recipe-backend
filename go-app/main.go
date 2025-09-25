package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	fileupload "github.com/wubshet-kebede/go-app/FileUpload"
	Handler "github.com/wubshet-kebede/go-app/Handler"
	hasura "github.com/wubshet-kebede/go-app/Hasura"
	"github.com/wubshet-kebede/go-app/contact"
	"github.com/wubshet-kebede/go-app/payment"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Printf("Starting server on port %s...", port)

	hasura.InitClient()
hService := payment.NewHasuraService()
cService := payment.NewChapaService()

	r := mux.NewRouter()
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("/app/uploads"))))
	r.HandleFunc("/signUp", Handler.SignupHandler).Methods("POST")
	r.HandleFunc("/login", Handler.LoginHandler).Methods("POST")
	r.HandleFunc("/uploadFiles", fileupload.UploadFilesHandler).Methods("POST")
r.HandleFunc("/initiate_chapa_payment", func(w http.ResponseWriter, r *http.Request) {
    payment.HandleInitiateChapaPayment(w, r, hService, cService)
}).Methods("POST")
   


r.HandleFunc("/chapa/callback", func(w http.ResponseWriter, r *http.Request) {
    log.Println("🔥 Chapa callback received!")
    log.Printf("Method: %s, URL: %s\n", r.Method, r.URL.String())
    payment.HandleChapaCallback(w, r, hService, cService)
}).Methods("GET", "POST")
	r.HandleFunc("/submitContactForm", contact.HandleSubmitContactForm).Methods("POST")
	// THIS MUST BLOCK
	err := http.ListenAndServe("0.0.0.0:"+port, r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
