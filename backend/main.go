package main

import (
	"github.com/villepalo/pacman-go-react/db"

	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	db.InitDB()
	CleanupExpiredSessions() // Start session cleanup goroutine
	mux := http.NewServeMux()

	// Initialize Lobby
	lobby := NewLobby()
	go lobby.Run()

	// Serve static files from frontend/dist
	rootDir := "."
	if _, err := os.Stat("frontend/dist"); os.IsNotExist(err) {
		rootDir = ".."
	}

	fs := http.FileServer(http.Dir(filepath.Join(rootDir, "frontend/dist")))
	mux.Handle("/", fs)

	// Register API routes
	RegisterRoutes(mux, lobby)

	// Middleware for security headers
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com;")
		mux.ServeHTTP(w, r)
	})

	fmt.Println("Starting Pacman Backend on :6060...")
	if err := http.ListenAndServe(":6060", handler); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
