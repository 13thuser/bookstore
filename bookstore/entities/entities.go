package entities

import "fmt"

// Type aliases for better code readability
type UserID = string
type SKU = string
type OrderID = string

// User defines the structure of a user
type User struct {
	ID   UserID
	Name string
}

// UserCredentials defines the structure of user credentials
type UserCredentials struct {
	UserID   string `json:"username"`
	Password string `json:"password"`
}

// Item defines the structure of an item
type Item struct {
	SKU   SKU     `json:"sku"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Items defines a list of items
type ItemsResponse struct {
	Items []Item `json:"items"`
}

// ItemCartRequest defines the structure of an item SKU
type ItemCartRequest struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// CreditCardDetails represents a credit card
type CreditCardDetails struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Number     string `json:"credit_card_number"`
	Expiration string `json:"credit_card_expiration"`
	CVV        string `json:"credit_card_cvv"`
}

// ConfirmPurchaseRequest defines the structure of a purchase confirmation request
type ConfirmPurchaseRequest struct {
	OrderID           string            `json:"order_id"`
	CreditCardDetails CreditCardDetails `json:"credit_card_details"`
}

// ItemWithQty defines the structure of a cart item
type ItemWithQty struct {
	Item     Item
	Quantity int
}

// Order defines the structure of an order
type Order struct {
	ID                  OrderID
	UserID              UserID
	Items               []ItemWithQty
	TotalItems          int
	TotalPrice          float64
	PaymentConfirmation string `json:"payment_confirmation,omitempty"`
}

// CartItem defines the structure of a cart item
type CartItem struct {
	Item     *Item
	Quantity int
}

// Cart defines the structure of a cart
type Cart struct {
	UserID     UserID           `json:"user_id"`
	Items      map[SKU]CartItem `json:"items"`
	TotalItems int              `json:"total_items"`
	TotalPrice float64          `json:"total_price"`
}

// NewCart creates a new cart
func NewCart(userID UserID) *Cart {
	return &Cart{
		UserID: userID,
		Items:  make(map[SKU]CartItem),
	}
}

// AddToCart adds an item to the cart and updates the total price
func (c *Cart) AddToCart(item *Item, quantity int) {
	if c.Items == nil {
		c.Items = make(map[SKU]CartItem)
	}
	if cartItem, ok := c.Items[item.SKU]; ok {
		cartItem.Quantity += quantity
		c.Items[item.SKU] = cartItem
	} else {
		c.Items[item.SKU] = CartItem{
			Item:     item,
			Quantity: quantity,
		}
	}
	c.TotalItems += quantity
	c.TotalPrice += item.Price * float64(quantity)
}

// RemoveFromCart removes an item from the cart and updates the total price
func (c *Cart) RemoveFromCart(item Item, quantity int) error {
	if c.Items == nil {
		return fmt.Errorf(fmt.Sprintf("item id %s not found in the cart", item.SKU))
	}
	if cartItem, ok := c.Items[item.SKU]; ok {
		if cartItem.Quantity > quantity {
			cartItem.Quantity -= quantity
			c.Items[item.SKU] = cartItem
			c.TotalItems -= quantity
		} else {
			delete(c.Items, item.SKU)
			c.TotalItems -= cartItem.Quantity
		}
		c.TotalPrice -= item.Price * float64(quantity)
	}
	return nil
}
