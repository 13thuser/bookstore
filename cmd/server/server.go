package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/13thuser/bookstore/bookstore"
	"github.com/13thuser/bookstore/datastore"
	"github.com/13thuser/bookstore/payments"
	"github.com/gorilla/mux"
)

// Server defines the structure of the server
type Server struct {
	server   *http.Server
	handler  http.Handler
	service  StoreService
	auth     *datastore.UserStore
	sessions *datastore.SessionStore
}

// NewServer creates a new server
func NewServer() *Server {
	paymentGateway := payments.NewPaymentGateway()
	storeService := bookstore.NewBookstoreService(datastore.NewDatastore(), paymentGateway)
	return &Server{
		server:   nil,
		handler:  nil,
		service:  storeService,
		auth:     datastore.NewUserStore(),
		sessions: datastore.NewSessionStore(),
	}
}

// init initializes the server
func (s *Server) init(port string) {
	handler := s.WithMiddlewares(s.WithEndpointsSetup(mux.NewRouter()), authMiddleware)
	s.server = &http.Server{
		Addr:    port,
		Handler: handler,
	}
}

// Serve starts the server
func (s *Server) Serve(port string) error {
	if s.server == nil {
		s.init(port)
	}
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting any active connections
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func main() {
	s := NewServer()
	port := fmt.Sprintf(":%s", SERVER_PORT)
	fmt.Printf("Server listening on port %s...", SERVER_PORT)
	go func() {
		if err := s.Serve(port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for an interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
