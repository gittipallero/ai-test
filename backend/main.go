package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity in this demo
	},
}

func main() {
	InitDB()
	mux := http.NewServeMux()

	// Serve static files from frontend/dist
	rootDir := "."
	if _, err := os.Stat("frontend/dist"); os.IsNotExist(err) {
		rootDir = ".."
	}

	fs := http.FileServer(http.Dir(filepath.Join(rootDir, "frontend/dist")))
	mux.Handle("/", fs)

	mux.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Upgrade error:", err)
			return
		}
		defer conn.Close()

		nickname := r.URL.Query().Get("nickname")
		if nickname == "" {
			nickname = "Anonymous"
		}

		game := NewGame()

		// Handle Input
		go func() {
			for {
				var msg map[string]string
				if err := conn.ReadJSON(&msg); err != nil {
					return
				}
				if dir, ok := msg["direction"]; ok {
					game.SetNextDirection(Direction(dir))
				}
			}
		}()

		// Game Loop
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			game.Update()

			game.mu.RLock()
			gameOver := game.GameOver
			score := game.Score
			// Create a clean state object to send
			// We can just send the game object since fields are exported and tagged
			// But we must marshal inside lock to prevent race on slices
			bytes, err := json.Marshal(game)
			game.mu.RUnlock()

			if err != nil {
				fmt.Println("Marshal error:", err)
				break
			}

			if err := conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
				break
			}

			if gameOver {
				// Save score
				if err := SaveScore(nickname, score); err != nil {
					fmt.Println("Failed to save score:", err)
				} else {
					fmt.Println("Score saved for", nickname, ":", score)
				}
				// Give meaningful time for client to receive GameOver state before closing
				time.Sleep(500 * time.Millisecond)
				// We don't break immediately, let the client disconnect or just stop sending updates?
				// Actually, we should probably stop the loop.
				break
			}
		}
	})

	mux.HandleFunc("/api/score", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !RequireDB(w) {
			return
		}

		var req ScoreSubmitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := SaveScore(req.Nickname, req.Score); err != nil {
			fmt.Println("Score update error:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Score submitted"})
	})

	mux.HandleFunc("/api/scoreboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !RequireDB(w) {
			return
		}

		scores, err := GetTopScores()
		if err != nil {
			fmt.Println("Scoreboard query error:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(scores)
	})

	mux.HandleFunc("/api/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !RequireDB(w) {
			return
		}
		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := CreateUser(req.Nickname, req.Password); err != nil {
			fmt.Println("Signup error:", err)
			http.Error(w, "Username already taken or database error", http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "User created", "nickname": req.Nickname})
	})

	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !RequireDB(w) {
			return
		}
		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := VerifyUser(req.Nickname, req.Password); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful", "nickname": req.Nickname})
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
