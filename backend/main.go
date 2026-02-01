package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: CheckOrigin, // Validate origin against allowed list from ALLOWED_ORIGINS env var
}

func main() {
	InitDB()
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

	mux.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		// Validate session token for WebSocket authentication
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Missing authentication token", http.StatusUnauthorized)
			return
		}

		nickname, valid := ValidateSession(token)
		if !valid {
			http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Upgrade error:", err)
			return
		}

        client := &Client{
            Nickname: nickname,
            Conn:     conn,
            Send:     make(chan []byte, 256),
            Lobby:    lobby,
        }
        
        lobby.register <- client
        
        // Handle incoming messages
        go func() {
            defer func() {
                lobby.unregister <- client
                conn.Close()
            }()
            
            for {
                var msg map[string]interface{}
                if err := conn.ReadJSON(&msg); err != nil {
                    break
                }
                
                msgType, ok := msg["type"].(string)
                if !ok {
                   // Fallback for legacy single player directional input which might just be { direction: "..." }
                   if _, ok := msg["direction"]; ok {
                        handleGameInput(client, msg)
                   }
                   continue
                }
                
                switch msgType {
                case "join_pair":
                    lobby.JoinPairQueue(client)
                case "input":
                    handleGameInput(client, msg)
                case "start_single":
                     startSinglePlayerGame(client)
                }
            }
        }()
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
        // ... existing single player scoreboard ...
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
    
    mux.HandleFunc("/api/scoreboard/pair", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !RequireDB(w) {
			return
		}

		scores, err := GetTopPairScores()
		if err != nil {
			fmt.Println("PairScoreboard query error:", err)
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
			if errors.Is(err, ErrUsernameTaken) {
				http.Error(w, "Username already taken", http.StatusConflict)
			} else {
				http.Error(w, "Server error", http.StatusInternalServerError)
			}
			return
		}

		// Create session token for newly registered user (auto-login)
		session, err := CreateSession(req.Nickname)
		if err != nil {
			fmt.Println("Session creation error:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message":  "User created",
			"nickname": req.Nickname,
			"token":    session.Token,
		})
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

		// Create session token for authenticated user
		session, err := CreateSession(req.Nickname)
		if err != nil {
			fmt.Println("Session creation error:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message":  "Login successful",
			"nickname": req.Nickname,
			"token":    session.Token,
		})
	})

	mux.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract token from Authorization header (Bearer token)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Parse "Bearer <token>" format
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		token := authHeader[len(bearerPrefix):]

		// Validate that the session exists before deleting
		if _, valid := ValidateSession(token); !valid {
			http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
			return
		}

		// Delete the session to invalidate it
		DeleteSession(token)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Logout successful",
		})
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

func handleGameInput(client *Client, msg map[string]interface{}) {
    game := client.GetGame()
    if game == nil {
        return
    }
    // Expected: { "direction": "UP/DOWN..." }
    if dirVal, ok := msg["direction"].(string); ok {
        game.SetNextDirection(client.Nickname, Direction(dirVal))
    }
}

func startSinglePlayerGame(client *Client) {
    // Similar to old main loop but running in goroutine per client (or associated with client)
    // Actually, client.Game stores the state.
    
    // If already in game, stop it?
    // Let's assume start_single starts a new one.
    
    game := NewGame([]string{client.Nickname})
    client.SetGame(game)
    
    // Start Ticker Loop for this single player game
    go func() {
        ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()
        
        // Notify start
        startMsg := map[string]interface{}{
            "type": "game_start", 
            "mode": "single",
             "p1": client.Nickname,
        }
		client.WriteJSON(startMsg)

		for range ticker.C {
            // Check if client disconnected (handled by lobby unregister closing channel?)
            // We can check if client.Game is still this game
            if client.GetGame() != game {
                return 
            }
            
			game.Update()

			game.mu.RLock()
			// Send update
			err := client.WriteJSON(game)
            gameOver := game.GameOver
            score := game.Score
			game.mu.RUnlock()

			if err != nil {
				break
			}

			if gameOver {
				// Save score
				if err := SaveScore(client.Nickname, score); err != nil {
					fmt.Println("Failed to save score:", err)
				}
				// Sleep a bit and stop
				time.Sleep(500 * time.Millisecond)
                client.SetGame(nil)
				break
			}
		}
    }()
}
