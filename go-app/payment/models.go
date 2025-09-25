package payment

import (
	"encoding/json"
	"time"
)

// A custom UUID type to match Hasura's GraphQL UUID scalar
type uuid string

// A custom DateTime type to handle Hasura's timestamptz
type DateTime time.Time

func (dt DateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(dt)
	return json.Marshal(t.Format(time.RFC3339Nano))
}

func (dt *DateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	*dt = DateTime(t)
	return nil
}

// Hasura Actions
type RecipeItemInput struct {
	RecipeID string `json:"recipeId"`
	Quantity int    `json:"quantity"`
}

type HasuraActionPayload struct {
	Action struct {
		Name string `json:"name"`
	} `json:"action"`
	Input struct {
		Input struct {
			RecipeItems []RecipeItemInput `json:"recipeItems"`
			ReturnURL   string            `json:"returnUrl"`
			Amount      float64           `json:"amount"`
			Currency    string            `json:"currency"`
		} `json:"input"`
	} `json:"input"`
	SessionVariables map[string]string `json:"session_variables"`
}

// GraphQL Mutations and Queries
type orders_insert_input struct {
	ID        uuid    `json:"id" graphql:"id"`
	UserID    uuid    `json:"user_id" graphql:"user_id"`
	TotalAmount float64 `json:"total_amount" graphql:"total_amount"`
	Currency    string  `json:"currency" graphql:"currency"`
	ReturnURL   string  `json:"return_url" graphql:"return_url"`
	Status      string  `json:"status" graphql:"status"`
	ChapaTxRef  string  `json:"chapa_tx_ref" graphql:"chapa_tx_ref"`
	ChapaTransactionID *string `json:"chapa_transaction_id,omitempty" graphql:"chapa_transaction_id"`
	CreatedAt   DateTime `json:"created_at" graphql:"created_at"`
	UpdatedAt   DateTime `json:"updated_at" graphql:"updated_at"`
}

type recipesDetailsQuery struct {
	Recipes []struct {
		ID       string  `json:"id" graphql:"id"`
		PriceETB float64 `json:"price_etb" graphql:"price_etb"`
		Title    string  `json:"title" graphql:"title"`
		Images   []struct {
			ID         string  `json:"id" graphql:"id"`
			ImageURL   string  `json:"image_url" graphql:"image_url"`
			
			IsFeatured *bool   `json:"is_featured" graphql:"is_featured"`
		} `graphql:"recipe_images(order_by: {image_order: asc})"` // Fetch all images ordered
	} `graphql:"recipes(where: {id: {_in: $recipeIds}})"`
}



type userDetailsQuery struct {
	UsersByPk *struct {
		FirstName   string `json:"first_name" graphql:"first_name"`
		LastName    string `json:"last_name" graphql:"last_name"`
		Email       string `json:"email" graphql:"email"`
		PhoneNumber string `json:"phone_number" graphql:"phone_number"`
	} `graphql:"users_by_pk(id: $id)"`
}


type insertOrderMutation struct {
	InsertOrdersOne *struct {
		OrderID   string `graphql:"id"`
		TxRef     string `graphql:"chapa_tx_ref"`
		ReturnURL string `graphql:"return_url"`
	} `graphql:"insert_orders_one(object: $object)"`
}

type insertOrderItemsMutation struct {
	InsertOrderItems *struct {
		AffectedRows int `graphql:"affected_rows"`
	} `graphql:"insert_order_items(objects: $objects)"`
}


type order_items_insert_input struct {
	ID             uuid      `json:"id" graphql:"id"`
	OrderID        uuid      `json:"order_id" graphql:"order_id"`
	RecipeID       uuid      `json:"recipe_id" graphql:"recipe_id"`
	Quantity       int       `json:"quantity" graphql:"quantity"`
	PriceAtPurchase float64  `json:"price_at_purchase" graphql:"price_at_purchase"`
	RecipeName     string    `json:"recipe_name" graphql:"recipe_name"`
	RecipeImageURL *string   `json:"recipe_image_url,omitempty" graphql:"recipe_image_url"`
	CreatedAt      DateTime  `json:"created_at" graphql:"created_at"`
	UpdatedAt      DateTime  `json:"updated_at" graphql:"updated_at"`
}



type updateOrderStatusMutation struct {
    UpdateOrders *struct {
        AffectedRows int `graphql:"affected_rows"`
        Returning    []struct {
            OrderID string `graphql:"id"`
            TxRef   string `graphql:"chapa_tx_ref"`
            Status  string `graphql:"status"`
        } `graphql:"returning"`
    } `graphql:"update_orders(where: {chapa_tx_ref: {_eq: $txRef}}, _set: $set)"`
}

// Chapa API
type ChapaInitiateRequest struct {
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	TxRef       string `json:"tx_ref"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	CallbackURL string `json:"callback_url"`
	ReturnURL   string `json:"return_url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ChapaInitiateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		CheckoutURL string `json:"checkout_url"`
	} `json:"data"`
}

type ChapaCallbackPayload struct {
	TxRef string `json:"tx_ref"`
}

type ChapaVerifyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID       string  `json:"id"`
		Amount   float64 `json:"amount"`
		TxRef    string  `json:"tx_ref"`
		Currency string  `json:"currency"`
		Status   string  `json:"status"`
	} `json:"data"`
}