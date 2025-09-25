package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	google_uuid "github.com/google/uuid"
)

// HandleInitiateChapaPayment handles the Hasura Action webhook for payment initiation.
func HandleInitiateChapaPayment(w http.ResponseWriter, r *http.Request, hasuraService *HasuraService, chapaService *ChapaService) {
	log.Println("🚀 Received /initiate_chapa_payment request")
	if hasuraService == nil || chapaService == nil {
		log.Println("❌ Services not initialized")
		respondWithError(w, http.StatusInternalServerError, "Internal server error: Services not ready")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var payload HasuraActionPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse Hasura Action payload")
		return
	}

	input := payload.Input.Input
	buyerID := payload.SessionVariables["x-hasura-user-id"]
	if buyerID == "" || len(input.RecipeItems) == 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload or missing user ID")
		return
	}

	// 1. Prepare recipe IDs and quantities
	recipeIDs := make([]string, len(input.RecipeItems))
	recipeQuantities := make(map[string]int)
	for i, item := range input.RecipeItems {
		recipeIDs[i] = item.RecipeID
		recipeQuantities[item.RecipeID] = item.Quantity
	}

	ctx := context.Background()
	recipesResp, err := hasuraService.QueryRecipeDetails(ctx, recipeIDs)
	if err != nil || len(recipesResp.Recipes) == 0 {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve recipe details")
		return
	}

	var backendCalculatedAmount float64
	var orderItemsForInsertion []order_items_insert_input
	for _, dbRecipe := range recipesResp.Recipes {
		quantity := recipeQuantities[dbRecipe.ID]
		backendCalculatedAmount += dbRecipe.PriceETB * float64(quantity)

		// Pick featured image, fallback to first image if available
		var imgURL *string
		for _, img := range dbRecipe.Images {
			if img.IsFeatured != nil && *img.IsFeatured {
				imgURL = &img.ImageURL
				break
			}
		}
		if imgURL == nil && len(dbRecipe.Images) > 0 {
			imgURL = &dbRecipe.Images[0].ImageURL
		}

		orderItemsForInsertion = append(orderItemsForInsertion, order_items_insert_input{
			ID:             uuid(google_uuid.New().String()),
			RecipeID:       uuid(dbRecipe.ID),
			Quantity:       quantity,
			PriceAtPurchase: dbRecipe.PriceETB,
			RecipeName:     dbRecipe.Title,
			RecipeImageURL: imgURL,
			CreatedAt:      DateTime(time.Now()),
			UpdatedAt:      DateTime(time.Now()),
		})
	}

	if fmt.Sprintf("%.2f", input.Amount) != fmt.Sprintf("%.2f", backendCalculatedAmount) {
		respondWithError(w, http.StatusBadRequest, "Amount mismatch. Please try again.")
		return
	}

	// 2. Query user details
	userResp, err := hasuraService.QueryUserDetails(ctx, buyerID)
	if err != nil || userResp.UsersByPk == nil {
		respondWithError(w, http.StatusNotFound, "Buyer details not found")
		return
	}
	user := userResp.UsersByPk

	// 3. Insert pending order
	//txRef := fmt.Sprintf("chapa-order-%s-%d", google_uuid.New().String(), time.Now().Unix())
	txRef := fmt.Sprintf("c-%s-%d", google_uuid.New().String()[:8], time.Now().Unix())

	orderObject := orders_insert_input{
		ID:        uuid(google_uuid.New().String()),
		UserID:    uuid(buyerID),
		TotalAmount: backendCalculatedAmount,
		Currency:  input.Currency,
		ReturnURL: input.ReturnURL,
		Status:    "pending",
		ChapaTxRef: txRef,
		CreatedAt: DateTime(time.Now()),
		UpdatedAt: DateTime(time.Now()),
	}

	orderResp, err := hasuraService.InsertOrder(ctx, orderObject)
	if err != nil || orderResp.InsertOrdersOne == nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to record pending order")
		return
	}
	orderID := orderResp.InsertOrdersOne.OrderID

	// Set OrderID in order items
	for i := range orderItemsForInsertion {
		orderItemsForInsertion[i].OrderID = uuid(orderID)
	}
	hasuraService.InsertOrderItems(ctx, orderItemsForInsertion) // Log error but continue

	// 4. Initiate Chapa payment
	
	chapaRequest := ChapaInitiateRequest{
		Amount:      fmt.Sprintf("%.2f", backendCalculatedAmount),
		Currency:    input.Currency,
		TxRef:       txRef,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		CallbackURL: os.Getenv("CHAPA_CALLBACK_URL"),
		ReturnURL:   getFrontendRedirectURL(input.ReturnURL, "pending", orderID, txRef, ""),
		Title:       "Food Recipes Order",
		Description: fmt.Sprintf("Order ID: %s", orderID),
	}
log.Printf("Chapa request payload: %+v", chapaRequest)

	chapaResp, err := chapaService.InitiatePayment(ctx, chapaRequest)
	if err != nil {
		log.Printf("Chapa API error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Payment service failed to initiate")
		return
	}
log.Printf("✅ Chapa response: %+v", chapaResp)
	respondWithJSON(w, http.StatusOK, map[string]string{
		"checkoutUrl": chapaResp.Data.CheckoutURL,
		"message":     "Payment initiated successfully",
		"orderId":     orderID,
		"txRef":       txRef,
	})
}

// HandleChapaCallback handles the webhook callback from Chapa after a payment attempt.
func HandleChapaCallback(w http.ResponseWriter, r *http.Request, hasuraService *HasuraService, chapaService *ChapaService) {
	txRef := r.URL.Query().Get("trx_ref")
log.Printf("🚀 Chapa callback received: tx_ref=%s, full query=%v", txRef, r.URL.Query())

	
	if txRef == "" {
		respondWithError(w, http.StatusBadRequest, "Missing tx_ref in query")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Query Hasura for order
	orderData, err := hasuraService.QueryOrderForCallback(ctx, txRef)
log.Printf("Queried order for txRef=%s: %+v, err=%v", txRef, orderData, err)

	
	if err != nil || len(orderData) == 0 {
		log.Printf("Failed to query order for TxRef %s: %v", txRef, err)
		http.Error(w, "Order not found for verification", http.StatusOK)
		return
	}
	originalReturnURL := orderData[0].ReturnURL
	orderID := orderData[0].OrderID

	// Verify with Chapa
	chapaVerifyResp, err := chapaService.VerifyPayment(ctx, txRef)
	log.Printf("Chapa verification response: %+v, err=%v", chapaVerifyResp, err)
	if err != nil {
		log.Printf("Failed to verify transaction %s with Chapa: %v", txRef, err)
		finalRedirectURL := getFrontendRedirectURL(originalReturnURL, "failure", orderID, txRef, "Verification failed")
		http.Redirect(w, r, finalRedirectURL, http.StatusFound)
		return
	}

	status := chapaVerifyResp.Data.Status
	dbStatus := "unknown"
	message := "Payment status unknown."
	if status == "success" {
		dbStatus = "completed"
		message = "Your payment was successful!"
	} else if status == "failed" {
		dbStatus = "failed"
		message = "Your payment failed."
	}

	// Update order with Chapa transaction ID
	chapaTxID := chapaVerifyResp.Data.ID
	hasuraService.UpdateOrderStatus(ctx, txRef, dbStatus, chapaTxID)

	finalRedirectURL := getFrontendRedirectURL(originalReturnURL, dbStatus, orderID, txRef, message)
	http.Redirect(w, r, finalRedirectURL, http.StatusFound)
}
