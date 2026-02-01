package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Session management for secure WebSocket authentication
var (
	sessions   = make(map[string]*Session)
	sessionsMu sync.RWMutex
)

type Session struct {
	Token     string
	Nickname  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

const sessionDuration = 24 * time.Hour

// GenerateSessionToken creates a cryptographically secure session token
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new session for an authenticated user
func CreateSession(nickname string) (*Session, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return nil, err
	}

	session := &Session{
		Token:     token,
		Nickname:  nickname,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sessionDuration),
	}

	sessionsMu.Lock()
	sessions[token] = session
	sessionsMu.Unlock()

	return session, nil
}

// ValidateSession checks if a session token is valid and returns the associated nickname
func ValidateSession(token string) (string, bool) {
	sessionsMu.RLock()
	session, exists := sessions[token]
	sessionsMu.RUnlock()

	if !exists {
		return "", false
	}

	if time.Now().After(session.ExpiresAt) {
		// Session expired, remove it
		sessionsMu.Lock()
		delete(sessions, token)
		sessionsMu.Unlock()
		return "", false
	}

	return session.Nickname, true
}

// DeleteSession removes a session (for logout)
func DeleteSession(token string) {
	sessionsMu.Lock()
	delete(sessions, token)
	sessionsMu.Unlock()
}

// CleanupExpiredSessions periodically removes expired sessions
func CleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			sessionsMu.Lock()
			now := time.Now()
			for token, session := range sessions {
				if now.After(session.ExpiresAt) {
					delete(sessions, token)
				}
			}
			sessionsMu.Unlock()
		}
	}()
}

// GetAllowedOrigins returns the list of allowed origins from environment variable
func GetAllowedOrigins() []string {
	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	if originsEnv == "" {
		// Default for local development
		return []string{
			"http://localhost:6060",
			"http://localhost:5173",
			"http://127.0.0.1:6060",
			"http://127.0.0.1:5173",
		}
	}

	// Parse comma-separated origins
	origins := strings.Split(originsEnv, ",")
	var trimmed []string
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o != "" {
			trimmed = append(trimmed, o)
		}
	}
	return trimmed
}

// CheckOrigin validates if the request origin is allowed
func CheckOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	
	// If no origin header (same-origin request), allow it
	if origin == "" {
		return true
	}

	allowedOrigins := GetAllowedOrigins()
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	fmt.Printf("Rejected WebSocket connection from origin: %s\n", origin)
	return false
}
