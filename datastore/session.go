package datastore

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// TokenID defines the type for the session ID
type TokenID = string

// SessionStore defines the structure of the session store
type SessionStore struct {
	sessions map[TokenID]UserID
	users    map[UserID]User
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	ss := &SessionStore{
		sessions: make(map[TokenID]UserID),
		users:    make(map[UserID]User),
	}
	// TODO: Remove this
	ss.seedData()
	return ss
}

// createNewSessionID creates a new session ID
func (s *SessionStore) createNewSessionID() (TokenID, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("unable to create session")
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// NewSessionStore creates a new session store
func (s *SessionStore) AddSession(userID string, user *User) (TokenID, error) {
	if prevSessionID, ok := s.sessions[userID]; ok {
		return prevSessionID, fmt.Errorf("session already exists for the user")
	}
	sessionID, err := s.createNewSessionID()
	if err != nil {
		return "", err

	}
	s.sessions[sessionID] = user.ID
	return sessionID, nil
}

// RemoveSession removes a session from the session store based on the user ID
func (s *SessionStore) RemoveSession(userID string) {
	// check if user exists
	_, ok := s.sessions[userID]
	if !ok {
		return
	}
	delete(s.sessions, userID)
}

// GetUserID retrieves a session from the session store based on the user ID
func (s *SessionStore) GetUserID(tokenID string) UserID {
	return s.sessions[tokenID]
}
