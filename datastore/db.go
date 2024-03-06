package datastore

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/13thuser/bookstore/bookstore/entities"
)

// Define type aliases for all the types from entities.go file
type UserID = entities.UserID
type SKU = entities.SKU
type OrderID = entities.OrderID
type User = entities.User
type Item = entities.Item
type Cart = entities.Cart
type ItemWithQty = entities.ItemWithQty
type ItemQuantity = int
type Order = entities.Order

// Datastore defines the structure of the datastore
type Datastore struct {
	inventory map[SKU]ItemQuantity
	items     map[SKU]Item
	orders    map[UserID][]*Order
	carts     map[UserID]*Cart
}

// NewDatastore creates a new datastore
func NewDatastore() *Datastore {
	db := &Datastore{
		inventory: make(map[SKU]ItemQuantity),
		items:     make(map[SKU]Item),
		orders:    make(map[UserID][]*Order),
		carts:     make(map[UserID]*Cart),
	}
	// TODO: Remove this
	db.seedItemData()
	return db
}

// AddItem adds an item to the datastore
func (ds *Datastore) AddItem(ctx context.Context, item Item, quantity int) error {
	// If no item exists, add the item to the inventory
	if _, ok := ds.items[item.SKU]; !ok {
		ds.items[item.SKU] = item
	}
	ds.inventory[item.SKU] += quantity
	return nil
}

// RemoveItem removes an item from the datastore
func (ds *Datastore) RemoveItem(ctx context.Context, item Item, quantity int) error {
	if _, ok := ds.items[item.SKU]; !ok {
		return fmt.Errorf("item not found in the datastore")
	}
	if ds.inventory[item.SKU] < quantity {
		return fmt.Errorf("insufficient stock")
	}
	ds.inventory[item.SKU] -= quantity
	return nil
}

// ListItems lists all the items from the datastore
func (ds *Datastore) ListItems(ctx context.Context) ([]Item, error) {
	var items []Item
	for _, item := range ds.items {
		items = append(items, item)
	}
	return items, nil
}

// GetItem retrieves an item from the datastore based on the item ID
func (ds *Datastore) GetItem(ctx context.Context, id string) (Item, error) {
	if item, ok := ds.items[id]; ok {
		return item, nil
	}
	return Item{}, fmt.Errorf("item not found in the datastore")
}

// AddToCart adds an item to the cart in the datastore
func (ds *Datastore) AddToCart(ctx context.Context, userID string, itemID string, quantity int) (Cart, error) {
	if _, ok := ds.carts[userID]; !ok {
		ds.carts[userID] = entities.NewCart(userID)
	}
	item, err := ds.GetItem(ctx, itemID)
	if err != nil {
		return Cart{}, err
	}
	if ds.inventory[item.SKU] < quantity {
		return Cart{}, fmt.Errorf("insufficient stock for item %s", item.SKU)
	}
	cart := ds.carts[userID]
	cart.AddToCart(&item, quantity)
	return *cart, nil
}

// RemoveFromCart removes an item from the cart in the datastore
func (ds *Datastore) RemoveFromCart(ctx context.Context, userID string, itemID string, quantity int) (Cart, error) {
	if _, ok := ds.carts[userID]; !ok {
		return Cart{}, fmt.Errorf("cart not found in the datastore")
	}
	item, err := ds.GetItem(ctx, itemID)
	if err != nil {
		return Cart{}, err
	}
	cart := ds.carts[userID]
	if err := cart.RemoveFromCart(item, quantity); err != nil {
		return Cart{}, err
	}
	return *cart, nil
}

// GetCart retrieves the cart from the datastore based on the user ID
func (ds *Datastore) GetCart(ctx context.Context, userID string) Cart {
	cart, ok := ds.carts[userID]
	if !ok {
		cart = entities.NewCart(userID)
		ds.carts[userID] = cart
	}
	return *cart
}

// GetCartTotalPrice retrieves the total price of the items in the cart from the datastore
func (ds *Datastore) GetCartTotalPrice(ctx context.Context, userID string) float64 {
	cart, ok := ds.carts[userID]
	if !ok {
		cart = entities.NewCart(userID)
		ds.carts[userID] = cart
	}
	return cart.TotalPrice
}

// GetOrderHistory retrieves the order history from the datastore based on the user ID
func (ds *Datastore) GetOrderHistory(ctx context.Context, userID string) []Order {
	if _, ok := ds.orders[userID]; !ok {
		ds.orders[userID] = make([]*Order, 0)
	}
	orders := make([]Order, len(ds.orders[userID]))
	for _, order := range ds.orders[userID] {
		orders = append(orders, *order)
	}
	return orders
}

// Checkout checks out the cart in the datastore
func (ds *Datastore) Checkout(ctx context.Context, userID string) (Order, error) {
	return ds.ConfirmOrder(ctx, userID)
	// return cart for the current user
	// cart, ok := ds.carts[userID]
	// if !ok {
	// 	return Cart{}, fmt.Errorf("cart not found in the datastore")
	// }
	// // Check for inventory
	// for k, v := range cart.Items {
	// 	if ds.inventory[k] < v.Quantity {
	// 		return Cart{}, fmt.Errorf("insufficient stock for item %s", k)
	// 	}
	// }
	// return *cart, nil
}

// ConfirmOrder confirms the purchase in the datastore
func (ds *Datastore) ConfirmOrder(ctx context.Context, userID string) (Order, error) {
	cart, ok := ds.carts[userID]
	if !ok {
		return Order{}, fmt.Errorf("cart not found in the datastore")
	}
	if len(cart.Items) == 0 {
		return Order{}, fmt.Errorf("cart is empty")
	}
	// DoubleCheck for inventory
	for k, v := range cart.Items {
		if ds.inventory[k] < v.Quantity {
			return Order{}, fmt.Errorf("insufficient stock for item %s", k)
		}
	}
	// Update inventory
	for k, v := range cart.Items {
		ds.inventory[k] -= v.Quantity
	}

	newOrder, err := newOrderFromCart(userID, cart)
	if err != nil {
		return Order{}, fmt.Errorf("unable to create new order for user %s", userID)
	}
	// Append element at the front of the slice to show the latest order first
	ds.orders[userID] = append([]*Order{&newOrder}, ds.orders[userID]...)
	// Clear the cart
	ds.carts[userID] = entities.NewCart(userID)
	return newOrder, nil
}

// ConfirmPayment confirms the purchase in the datastore
func (ds *Datastore) ConfirmPayment(ctx context.Context, userID string, orderID OrderID, paymentConfirmationID string) (Order, error) {
	order, err := ds.findOrderByOrderID(userID, orderID)
	if err != nil {
		return Order{}, fmt.Errorf("order %v not found in the datastore", orderID)
	}
	if order.PaymentConfirmation != "" {
		return *order, fmt.Errorf("order %v already confirmed", orderID)
	}
	order.PaymentConfirmation = paymentConfirmationID
	return *order, nil
}

// findOrderByOrderID finds an order from the datastore based on the user ID and order ID
func (ds *Datastore) findOrderByOrderID(userID string, orderID OrderID) (*Order, error) {
	orders, ok := ds.orders[userID]
	if !ok {
		return nil, fmt.Errorf("order not found in the datastore")
	}
	for _, o := range orders {
		if o.ID == orderID {
			return o, nil
		}
	}
	return nil, fmt.Errorf("order not found in the datastore")
}

// FindOrder finds an order from the datastore based on the user ID and order ID
func (ds *Datastore) FindOrder(ctx context.Context, userID string, orderID OrderID) (Order, error) {
	order, err := ds.findOrderByOrderID(userID, orderID)
	if err != nil {
		return Order{}, err
	}
	return *order, nil
}

// createNeworderID creates a new order ID
func createNewOrderID() (OrderID, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("unable to create new order id")
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// newOrderFromCart creates a new order
func newOrderFromCart(userID UserID, cart *Cart) (Order, error) {
	rawOrderId, err := createNewOrderID()
	if err != nil {
		return Order{}, fmt.Errorf("unable to create new order id for user %s", userID)
	}
	orderId := fmt.Sprintf("order-%s", rawOrderId)
	order := Order{
		ID:         orderId,
		UserID:     userID,
		TotalItems: cart.TotalItems,
		TotalPrice: cart.TotalPrice,
	}
	for _, v := range cart.Items {
		order.Items = append(order.Items, ItemWithQty{
			Item:     *v.Item,
			Quantity: v.Quantity,
		})
	}
	return order, nil
}
