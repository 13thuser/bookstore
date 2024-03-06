package main

import (
	"context"

	"github.com/13thuser/bookstore/bookstore/entities"
	"github.com/13thuser/bookstore/payments"
)

// StoreService defines the interface for the bookstore service
type StoreService interface {
	// ListItems lists all the items
	ListItems(ctx context.Context) ([]entities.Item, error)
	// GetItem gets an item by its SKU
	GetItem(ctx context.Context, sku string) (entities.Item, error)
	// AddToCart adds an item to the cart
	AddToCart(ctx context.Context, userID string, sku string, quantity int) (entities.Cart, error)
	// RemoveFromCart removes an item from the cart
	RemoveFromCart(ctx context.Context, userID string, sku string, quantity int) (entities.Cart, error)
	// GetCart gets the cart
	GetCart(ctx context.Context, userID string) entities.Cart
	// GetCartTotalPrice gets the total price of the items in the cart
	GetCartTotalPrice(ctx context.Context, userID string) float64
	// Checkout checks out the cart
	Checkout(ctx context.Context, userID string) (entities.Order, error)
	// ConfirmOrder confirms the purchase
	ConfirmPurchase(ctx context.Context, userID string, orderID string, creditCardDetails payments.CreditCardDetails) (entities.Order, error)
	// GetOrderHistory gets the order history
	GetOrderHistory(ctx context.Context, userID string) []entities.Order
}
