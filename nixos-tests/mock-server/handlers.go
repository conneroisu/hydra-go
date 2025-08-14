package main

// This file is included in server.go through the handler methods
// Additional utility functions for the mock server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// generateID generates a random ID for new resources
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// cleanupSessions removes expired sessions
func (s *Server) cleanupSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.Sub(session.CreatedAt) > 24*time.Hour {
			delete(s.sessions, id)
		}
	}
}

// validateSession checks if a session is valid
func (s *Server) validateSession(sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return false
	}

	// Check if session is expired
	if time.Since(session.CreatedAt) > 24*time.Hour {
		return false
	}

	return true
}

// getSessionFromCookie extracts session ID from cookies
func getSessionFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("hydra_session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// requireAuth checks if request is authenticated
func (s *Server) requireAuth(r *http.Request) bool {
	sessionID := getSessionFromCookie(r)
	if sessionID == "" {
		return false
	}
	return s.validateSession(sessionID)
}
