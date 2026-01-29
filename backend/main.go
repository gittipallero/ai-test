package main

import (
    "encoding/json"
	"fmt"
	"net/http"
    "path/filepath"
    "os"
)

type ScoreResponse struct {
    HighScore int `json:"highScore"`
}

func main() {
    mux := http.NewServeMux()

    // Serve static files from frontend/dist
    // We assume the binary is run from repo root or backend, so we check path
    rootDir := "."
    if _, err := os.Stat("frontend/dist"); os.IsNotExist(err) {
        // If running from backend dir, frontend is ../frontend
        rootDir = ".."
    }
    
    fs := http.FileServer(http.Dir(filepath.Join(rootDir, "frontend/dist")))
    mux.Handle("/", fs)

    // specific API endpoint
    mux.HandleFunc("/api/score", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(ScoreResponse{HighScore: 1000})
    })

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
