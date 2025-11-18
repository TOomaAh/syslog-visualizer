package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user with authentication credentials
type User struct {
	Username     string
	PasswordHash string
	APIToken     string
	CreatedAt    time.Time
}

// Session represents an active user session
type Session struct {
	Token     string
	Username  string
	ExpiresAt time.Time
}

// AuthManager manages authentication and authorization
type AuthManager struct {
	users    map[string]*User
	sessions map[string]*Session
	mu       sync.RWMutex
	enabled  bool
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(enabled bool) *AuthManager {
	return &AuthManager{
		users:    make(map[string]*User),
		sessions: make(map[string]*Session),
		enabled:  enabled,
	}
}

// IsEnabled returns whether authentication is enabled
func (am *AuthManager) IsEnabled() bool {
	return am.enabled
}

// AddUser adds a new user with a hashed password and generates an API token
func (am *AuthManager) AddUser(username, password string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.users[username]; exists {
		return fmt.Errorf("user %s already exists", username)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate API token
	apiToken, err := generateAPIToken()
	if err != nil {
		return fmt.Errorf("failed to generate API token: %w", err)
	}

	am.users[username] = &User{
		Username:     username,
		PasswordHash: string(hash),
		APIToken:     apiToken,
		CreatedAt:    time.Now(),
	}

	return nil
}

// VerifyPassword verifies a username and password combination
func (am *AuthManager) VerifyPassword(username, password string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	user, exists := am.users[username]
	if !exists {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// VerifyAPIToken verifies an API token and returns the associated username
func (am *AuthManager) VerifyAPIToken(token string) (string, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, user := range am.users {
		if user.APIToken == token {
			return user.Username, true
		}
	}

	return "", false
}

// CreateSession creates a new session for a user
func (am *AuthManager) CreateSession(username string) (string, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.users[username]; !exists {
		return "", fmt.Errorf("user not found")
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	am.sessions[sessionToken] = &Session{
		Token:     sessionToken,
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return sessionToken, nil
}

// ValidateSession validates a session token
func (am *AuthManager) ValidateSession(token string) (string, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	session, exists := am.sessions[token]
	if !exists {
		return "", false
	}

	if time.Now().After(session.ExpiresAt) {
		return "", false
	}

	return session.Username, true
}

// DeleteSession deletes a session
func (am *AuthManager) DeleteSession(token string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	delete(am.sessions, token)
}

// GetAPIToken returns the API token for a user
func (am *AuthManager) GetAPIToken(username string) (string, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	user, exists := am.users[username]
	if !exists {
		return "", fmt.Errorf("user not found")
	}

	return user.APIToken, nil
}

// CleanupExpiredSessions removes expired sessions
func (am *AuthManager) CleanupExpiredSessions() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	for token, session := range am.sessions {
		if now.After(session.ExpiresAt) {
			delete(am.sessions, token)
		}
	}
}

// generateAPIToken generates a secure random API token
func generateAPIToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
}

// generateSessionToken generates a secure random session token
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// Middleware returns an HTTP middleware that requires authentication
func (am *AuthManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If authentication is disabled, allow all requests
		if !am.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Check for API token in Authorization header (Bearer token)
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 {
				if parts[0] == "Bearer" {
					if username, valid := am.VerifyAPIToken(parts[1]); valid {
						r.Header.Set("X-Username", username)
						next.ServeHTTP(w, r)
						return
					}
				} else if parts[0] == "Basic" {
					// Basic auth
					payload, err := base64.StdEncoding.DecodeString(parts[1])
					if err == nil {
						credentials := strings.SplitN(string(payload), ":", 2)
						if len(credentials) == 2 {
							if am.VerifyPassword(credentials[0], credentials[1]) {
								r.Header.Set("X-Username", credentials[0])
								next.ServeHTTP(w, r)
								return
							}
						}
					}
				}
			}
		}

		// Check for session cookie
		if cookie, err := r.Cookie("session"); err == nil {
			if username, valid := am.ValidateSession(cookie.Value); valid {
				r.Header.Set("X-Username", username)
				next.ServeHTTP(w, r)
				return
			}
		}

		// No valid authentication found
		w.Header().Set("WWW-Authenticate", `Bearer realm="API", Basic realm="Web"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
