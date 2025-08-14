// Package auth provides authentication operations for Hydra.
package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/conneroisu/hydra-go/hydra/client"
	"github.com/conneroisu/hydra-go/hydra/models"
)

// Service handles authentication operations.
type Service struct {
	client *client.Client
}

// NewService creates a new authentication service.
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// Login authenticates with username and password.
func (s *Service) Login(ctx context.Context, username, password string) (*models.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	req := &models.LoginRequest{
		Username: username,
		Password: password,
	}

	var user models.User
	if err := s.client.DoRequest(ctx, "POST", "/login", req, &user); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Store username in client for reference
	s.client.Username = username

	return &user, nil
}

// Logout clears the current session.
func (s *Service) Logout() {
	s.client.Logout()
}

// IsAuthenticated checks if the client is authenticated.
func (s *Service) IsAuthenticated() bool {
	return s.client.IsAuthenticated()
}

// GetCurrentUser returns the username of the authenticated user.
func (s *Service) GetCurrentUser() string {
	return s.client.GetUsername()
}
