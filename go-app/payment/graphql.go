package payment

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
	hasura "github.com/wubshet-kebede/go-app/Hasura"
)

// HasuraService encapsulates all Hasura-related logic.
type HasuraService struct {
	client *graphql.Client
}

// NewHasuraService creates a new instance of HasuraService.
func NewHasuraService() *HasuraService {
	return &HasuraService{
		client: hasura.Client,
	}
}

// QueryRecipeDetails fetches details for multiple recipes from Hasura.
func (s *HasuraService) QueryRecipeDetails(ctx context.Context, recipeIDs []string) (recipesDetailsQuery, error) {
	var resp recipesDetailsQuery
	var ids []uuid
	for _, id := range recipeIDs {
		ids = append(ids, uuid(id))
	}
	err := s.client.Query(ctx, &resp, map[string]interface{}{"recipeIds": ids})
	return resp, err
}

// QueryUserDetails fetches a single user's details from Hasura.
func (s *HasuraService) QueryUserDetails(ctx context.Context, userID string) (userDetailsQuery, error) {
	var resp userDetailsQuery
	err := s.client.Query(ctx, &resp, map[string]interface{}{"id": uuid(userID)})
	return resp, err
}

// InsertOrder inserts a new order into the database.
func (s *HasuraService) InsertOrder(ctx context.Context, order orders_insert_input) (insertOrderMutation, error) {
	var resp insertOrderMutation
	vars := map[string]interface{}{"object": order}
	err := s.client.Mutate(ctx, &resp, vars)
	return resp, err
}

// InsertOrderItems inserts multiple order items into the database.
func (s *HasuraService) InsertOrderItems(ctx context.Context, items []order_items_insert_input) (insertOrderItemsMutation, error) {
	var resp insertOrderItemsMutation
	vars := map[string]interface{}{"objects": items}
	err := s.client.Mutate(ctx, &resp, vars)
	return resp, err
}

// UpdateOrderStatus updates an existing order's status and Chapa transaction ID.

func (s *HasuraService) UpdateOrderStatus(ctx context.Context, txRef string, status string, chapaTxID string) (updateOrderStatusMutation, error) {
	var resp updateOrderStatusMutation
	vars := map[string]interface{}{
		"txRef": graphql.String(txRef),
		"set": map[string]interface{}{
			"status":               graphql.String(status),
			"chapa_transaction_id": graphql.String(chapaTxID),
			"updated_at":           DateTime(time.Now()),
		},
	}
	err := s.client.Mutate(ctx, &resp, vars)
	return resp, err
}

// QueryOrderForCallback fetches a specific order to get the return URL.
func (s *HasuraService) QueryOrderForCallback(ctx context.Context, txRef string) ([]struct {
	OrderID   string `graphql:"id"`
	ReturnURL string `graphql:"return_url"`
}, error) {
	var orderQuery struct {
		Orders []struct {
			OrderID   string `graphql:"id"`
			ReturnURL string `graphql:"return_url"`
		} `graphql:"orders(where: {chapa_tx_ref: {_eq: $txRef}})"`
	}
	err := s.client.Query(ctx, &orderQuery, map[string]interface{}{"txRef": graphql.String(txRef)})
	return orderQuery.Orders, err
}