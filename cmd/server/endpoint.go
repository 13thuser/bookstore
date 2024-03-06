package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/13thuser/bookstore/bookstore/entities"
	"github.com/13thuser/bookstore/datastore"
	"github.com/gorilla/mux"
)

// WithEndpointsSetup sets up the endpoints
func (s *Server) WithEndpointsSetup(router *mux.Router, middlewares ...Middleware) http.Handler {
	// public endpoints
	router.HandleFunc("/", s.Health).Methods("GET")
	router.HandleFunc("/health", s.Health).Methods("GET")
	router.HandleFunc("/login", s.loginHandler).Methods("POST")
	router.HandleFunc("/logout", s.logoutHandler).Methods("GET")
	router.HandleFunc("/listItems", s.listItems).Methods("GET")
	router.HandleFunc("/getItem/{itemID}", s.GetItem).Methods("GET")

	// auth enabled sub-routes
	router.HandleFunc("/addToCart", requireLogin(s, s.AddToCart)).Methods("POST")
	router.HandleFunc("/removeFromCart", requireLogin(s, s.RemoveFromCart)).Methods("POST")
	router.HandleFunc("/getCart", requireLogin(s, s.GetCart)).Methods("GET")
	router.HandleFunc("/getCartTotalPrice", requireLogin(s, s.GetCartTotalPrice)).Methods("GET")
	router.HandleFunc("/checkout", requireLogin(s, s.Checkout)).Methods("POST")
	router.HandleFunc("/confirmPurchase", requireLogin(s, s.ConfirmPurchase)).Methods("POST")
	router.HandleFunc("/orderHistory", requireLogin(s, s.GetOrderHistory)).Methods("GET")
	return router
}

// Health is the health check endpoint
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// WithMiddlewares adds middlewares to the handler
func (s *Server) WithMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	var wrapped http.Handler = handler
	// reverse the middlewares
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

// loginHandler logs in the user
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var userCreds entities.UserCredentials
	json.NewDecoder(r.Body).Decode(&userCreds)

	if userCreds.UserID == "" || userCreds.Password == "" {

		writeError(w, "Invalid request with one or more missing parameters", http.StatusBadRequest)
		return
	}

	user, authenticated := s.auth.Authenticate(userCreds.UserID, userCreds.Password)
	if !authenticated {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	sessionID, err := s.sessions.AddSession(user.ID, &user)
	if err != nil {
		writeError(w, "Unable to create session", http.StatusInternalServerError)
	}
	log.Printf("Logged in as %s\n", user.ID)
	response := struct {
		Token string `json:"token"`
	}{
		Token: sessionID,
	}
	json.NewEncoder(w).Encode(response)
}

// logoutHandler logs out the user
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	token := getTokenFromRequest(r)
	if token != "" {
		s.sessions.RemoveSession(token)
	}
	// message logout successfully
	response := struct {
		Message string `json:"message"`
	}{Message: "Logged out successfully"}
	json.NewEncoder(w).Encode(response)
}

// getTokenFromRequest gets the user from the request
func getTokenFromRequest(r *http.Request) string {
	ctx := r.Context()
	token := ctx.Value(contextKey(TOKEN)).(string)
	return token
}

// getUserIDFromRequest gets the user ID from the request
func getUserIDFromRequest(r *http.Request, sessionStore *datastore.SessionStore) string {
	token := getTokenFromRequest(r)
	if token == "" {
		return ""
	}
	return sessionStore.GetUserID(token)
}

// requireLogin is an interceptor middleware that checks if the user is logged in
func requireLogin(s *Server, next http.HandlerFunc) http.HandlerFunc {
	if s == nil {
		log.Fatal("Server is nil")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tokenID := getTokenFromRequest(r)
		if tokenID != "" {
			userID := s.sessions.GetUserID(tokenID)
			if userID != "" {
				if _, err := s.auth.GetUser(userID); err == nil {
					next(w, r)
					return
				}
			}
		}
		writeError(w, "Unauthorized", http.StatusUnauthorized)
	}
}

// listItems lists the items
func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.ListItems(r.Context())
	if err != nil {
		writeError(w, "Failed to get items from datastore", http.StatusInternalServerError)
		return
	}

	response := entities.ItemsResponse{
		Items: items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetItem gets an item
func (s *Server) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// sku := r.URL.Query().Get("sku")
	sku := vars["itemID"]
	if sku == "" {
		writeError(w, "Invalid request with one or more missing parameters", http.StatusBadRequest)
		return
	}
	item, err := s.service.GetItem(r.Context(), sku)
	if err != nil {
		writeError(w, "Failed to get item from datastore", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}

// AddToCart adds an item to the cart
func (s *Server) AddToCart(w http.ResponseWriter, r *http.Request) {
	userID, shouldReturn := s.validateRequest(r, w)
	if shouldReturn {
		return
	}

	var req entities.ItemCartRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	if req.SKU == "" || req.Quantity == 0 {
		writeError(w, "Invalid request with one or more missing parameters", http.StatusBadRequest)
		return
	}

	// Add the item to the cart
	cart, err := s.service.AddToCart(r.Context(), userID, req.SKU, req.Quantity)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to add item to cart: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cart)
}

// RemoveFromCart removes an item from the cart
func (s *Server) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	userID, shouldReturn := s.validateRequest(r, w)
	if shouldReturn {
		return
	}

	var req entities.ItemCartRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	if req.SKU == "" || req.Quantity == 0 {
		writeError(w, "Invalid request with one or more missing parameters", http.StatusBadRequest)
		return
	}

	// Remove the item from the cart
	cart, err := s.service.RemoveFromCart(r.Context(), userID, req.SKU, req.Quantity)
	if err != nil {
		writeError(w, "Failed to remove item from cart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cart)
}

// GetCart gets the cart
func (s *Server) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, shouldReturn := s.validateRequest(r, w)
	if shouldReturn {
		return
	}

	// Get the cart
	cart := s.service.GetCart(r.Context(), userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cart)
}

// GetCartTotalPrice gets the total price of the items in the cart
func (s *Server) GetCartTotalPrice(w http.ResponseWriter, r *http.Request) {
	userID, shouldReturn := s.validateRequest(r, w)
	if shouldReturn {
		return
	}

	// Get the total price of the cart
	totalPrice := s.service.GetCartTotalPrice(r.Context(), userID)

	// struct to hold the total price
	resp := struct {
		TotalPrice float64 `json:"total_price"`
	}{TotalPrice: totalPrice}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Checkout checks out the cart
func (s *Server) Checkout(w http.ResponseWriter, r *http.Request) {
	userID, shouldReturn := s.validateRequest(r, w)
	if shouldReturn {
		return
	}

	// Checkout the order
	order, err := s.service.Checkout(r.Context(), userID)
	if err != nil {
		writeError(w, "Failed to checkout cart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

// ConfirmPurchase confirms the purchase
func (s *Server) ConfirmPurchase(w http.ResponseWriter, r *http.Request) {
	userID, shouldReturn := s.validateRequest(r, w)
	if shouldReturn {
		return
	}

	var req entities.ConfirmPurchaseRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	if req.OrderID == "" || req.CreditCardDetails.Number == "" {
		writeError(w, "Invalid request with one or more missing parameters", http.StatusBadRequest)
		return
	}

	// Confirm the purchase
	order, err := s.service.ConfirmPurchase(r.Context(), userID, req.OrderID, req.CreditCardDetails)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to confirm purchase: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

// orderHistoryHandler gets the order history
func (s *Server) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	userID, shouldReturn := s.validateRequest(r, w)
	if shouldReturn {
		return
	}

	orders := s.service.GetOrderHistory(r.Context(), userID)
	orderHistory := struct {
		Orders []entities.Order `json:"orders"`
	}{Orders: orders}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orderHistory)
}

// validateRequest validates the request
func (s *Server) validateRequest(r *http.Request, w http.ResponseWriter) (string, bool) {
	userID := getUserIDFromRequest(r, s.sessions)
	if userID == "" {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return "", true
	}
	return userID, false
}

// writeError writes an error response
func writeError(w http.ResponseWriter, message string, statusCode int) {
	response := struct {
		Error string `json:"error"`
	}{Error: message}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
