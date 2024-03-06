package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/13thuser/bookstore/bookstore/entities"
)

// testHelperEncodeJson is a helper function to encode a JSON string
func testHelperEncodeJson(t *testing.T, s interface{}) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(s); err != nil {
		t.Fatal(err, "unable to encode JSON")
	}
	return buf.String()
}

func TestHealthEndpoint(t *testing.T) {
	s := NewServer()
	s.init("")

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := struct {
		Status string `json:"status"`
	}{Status: "OK"}
	want := testHelperEncodeJson(t, expected)
	got := rr.Body.String()
	if got != want {
		t.Errorf("handler returned unexpected body: got '%v' \nwant '%+v', %v, %v", got, want, len(got), len(want))
	}
}

func TestLoginEndpoint(t *testing.T) {
	s := NewServer()
	s.init("")

	reqBody := []byte(`{"username": "testuser", "password": "testuser"}`)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	got := rr.Body.String()
	var response struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(got), &response); err != nil {
		t.Errorf("failed to parse JSON response: %v", err)
	}
	token := response.Token

	if token == "" {
		t.Errorf("expected login to return token, got empty string")
	}
}

func TestLogoutEndpoint(t *testing.T) {
	s := NewServer()
	s.init("")

	req, err := http.NewRequest("GET", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	got := rr.Body.String()
	var response struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(got), &response); err != nil {
		t.Errorf("failed to parse JSON response: %v", err)
	}
	message := response.Message

	if message != "Logged out successfully" {
		t.Errorf("expected logout message to be 'Logged out successfully', got '%s'", message)
	}
}

func TestPurchaseFlow(t *testing.T) {
	s := NewServer()
	s.init("")

	reqBody := []byte(`{"username": "testuser", "password": "testuser"}`)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	got := rr.Body.String()
	var response struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(got), &response); err != nil {
		t.Errorf("failed to parse JSON response: %v", err)
	}
	token := response.Token

	if token == "" {
		t.Errorf("expected login to return token, got empty string")
	}

	// Add item to cart
	reqBody = []byte(`{"sku": "item-2", "quantity": 2}`)
	req, err = http.NewRequest("POST", "/addToCart", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)

	rr = httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	got = rr.Body.String()
	var cartResponse entities.Cart
	if err := json.Unmarshal([]byte(got), &cartResponse); err != nil {
		t.Errorf("failed to parse JSON response: %v", err)
	}
	if len(cartResponse.Items) != 1 {
		t.Errorf("expected cart to have 1 item, got %d", len(cartResponse.Items))
	}
	if _, ok := cartResponse.Items["item-2"]; !ok {
		t.Errorf("expected cart to contain item-2, but it was not found")
	}

	// Checkout
	req, err = http.NewRequest("POST", "/checkout", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)

	rr = httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	got = rr.Body.String()
	var checkoutOrder entities.Order
	if err := json.Unmarshal([]byte(got), &checkoutOrder); err != nil {
		t.Errorf("failed to parse JSON response: %v", err)
	}
	if checkoutOrder.TotalItems != 2 {
		t.Errorf("expected order to have 2 items, got %d", checkoutOrder.TotalItems)
	}
	if checkoutOrder.PaymentConfirmation != "" {
		t.Errorf("expected checkout order to have empty payment confirmation, got %s", checkoutOrder.PaymentConfirmation)
	}

	orderID := checkoutOrder.ID

	// Confirm purchase
	reqBody = []byte(`{"order_id": "` + orderID + `", "credit_card_details": {"credit_card_number": "123456789", "credit_card_expiration": "12/22", "credit_card_cvv": "123"}}`)
	req, err = http.NewRequest("POST", "/confirmPurchase", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)

	rr = httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	got = rr.Body.String()
	var purchasedOrder entities.Order
	if err := json.Unmarshal([]byte(got), &purchasedOrder); err != nil {
		t.Errorf("failed to parse JSON response: %v", err)
	}
	if purchasedOrder.ID != orderID {
		t.Errorf("expected purchased order to have ID %s, got %s", orderID, purchasedOrder.ID)
	}
	if purchasedOrder.PaymentConfirmation == "" {
		t.Errorf("expected purchased order to have payment confirmation, got empty string")
	}
}
