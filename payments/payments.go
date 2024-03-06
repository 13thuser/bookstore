package payments

import (
	"context"

	"github.com/13thuser/bookstore/bookstore/entities"
)

// Ideally you would like to read it from a config file or environment variable
const PAYMENT_GATEWAY_API_KEY = "my-api-key"

// PaymentRequest represents a payment
type PaymentRequest struct {
	ID     string
	UserID string
	Amount float64
}

// CreditCardDetails represents a credit card
type CreditCardDetails = entities.CreditCardDetails

// PaymentProcessor defines an interface for processing payments
type PaymentProcessor interface {
	ProcessPayment(ctx context.Context, payment PaymentRequest, cardDetails CreditCardDetails) (string, error)
}

// PaymentGateway represents a payment gateway
type PaymentGateway struct {
	APIKey string
}

// NewPaymentGateway creates a new payment gateway
func NewPaymentGateway() PaymentProcessor {
	return &PaymentGateway{
		APIKey: PAYMENT_GATEWAY_API_KEY,
	}
}

// ProcessPayment processes a payment
func (pg *PaymentGateway) ProcessPayment(ctx context.Context, payment PaymentRequest, cardDetails CreditCardDetails) (string, error) {
	// NOTE: One important thing is to have idempotency keys for the payment processing
	// to avoid processing the same payment multiple times. So, maybe, internally create
	// a unique idempotency key for each payment and use it in the payment processing
	// and associate it with the orderID. Ideally we would like to do it with in the service

	// generate random string
	idempotencyKey := payment.ID
	confirmationID := idempotencyKey

	// Simulate payment processing, return true if successful, false otherwise
	return confirmationID, nil
}
