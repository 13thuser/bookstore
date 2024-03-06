package bookstore

import (
	"context"
	"fmt"

	"github.com/13thuser/bookstore/bookstore/entities"
	"github.com/13thuser/bookstore/datastore"
	"github.com/13thuser/bookstore/payments"
)

// BookstoreService defines the structure of the bookstore service
type BookstoreService struct {
	Datastore      *datastore.Datastore
	PaymentGateway payments.PaymentProcessor
}

// NewBookstoreService creates a new bookstore service
func NewBookstoreService(ds *datastore.Datastore, pg payments.PaymentProcessor) *BookstoreService {
	return &BookstoreService{
		Datastore:      ds,
		PaymentGateway: pg,
	}
}

func (s *BookstoreService) ListItems(ctx context.Context) ([]entities.Item, error) {
	return s.Datastore.ListItems(ctx)
}

func (s *BookstoreService) GetItem(ctx context.Context, sku string) (entities.Item, error) {
	return s.Datastore.GetItem(ctx, sku)
}

func (s *BookstoreService) AddToCart(ctx context.Context, userID string, sku string, quantity int) (entities.Cart, error) {
	return s.Datastore.AddToCart(ctx, userID, sku, quantity)
}

func (s *BookstoreService) RemoveFromCart(ctx context.Context, userID string, sku string, quantity int) (entities.Cart, error) {
	return s.Datastore.RemoveFromCart(ctx, userID, sku, quantity)
}

func (s *BookstoreService) GetCart(ctx context.Context, userID string) entities.Cart {
	return s.Datastore.GetCart(ctx, userID)
}

func (s *BookstoreService) GetCartTotalPrice(ctx context.Context, userID string) float64 {
	return s.Datastore.GetCartTotalPrice(ctx, userID)
}

func (s *BookstoreService) Checkout(ctx context.Context, userID string) (entities.Order, error) {
	return s.Datastore.Checkout(ctx, userID)
}

func (s *BookstoreService) ConfirmPurchase(ctx context.Context, userID string, orderID string, creditCardDetails payments.CreditCardDetails) (entities.Order, error) {
	var order entities.Order
	if orderID == "" {
		order, err := s.Datastore.ConfirmOrder(ctx, userID)
		if err != nil {
			return entities.Order{}, err
		}
		orderID = order.ID
	} else {
		order, err := s.Datastore.FindOrder(ctx, userID, orderID)
		if err != nil {
			return entities.Order{}, err
		}
		if order.PaymentConfirmation != "" {
			return entities.Order{}, fmt.Errorf("order already confirmed")
		}
	}
	paymentRequest := payments.PaymentRequest{
		ID:     orderID,
		UserID: userID,
		Amount: order.TotalPrice,
	}
	paymentConfirmationID, err := s.PaymentGateway.ProcessPayment(ctx, paymentRequest, creditCardDetails)
	if err != nil {
		return order, fmt.Errorf("payment processing failed for order id %s. please retry", order.ID)
	}
	order, err = s.Datastore.ConfirmPayment(ctx, userID, orderID, paymentConfirmationID)
	if err != nil {
		return entities.Order{}, err
	}
	return order, nil
}

func (s *BookstoreService) GetOrderHistory(ctx context.Context, userID string) []entities.Order {
	return s.Datastore.GetOrderHistory(ctx, userID)
}
