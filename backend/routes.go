package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/villepalo/pacman-go-react/db"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: CheckOrigin, // Validate origin against allowed list from ALLOWED_ORIGINS env var
}

func RegisterRoutes(mux *http.ServeMux, lobby *Lobby) {
	mux.HandleFunc("/api/ws", onApiWs(lobby))
	mux.HandleFunc("/api/score", onApiScore)
	mux.HandleFunc("/api/scoreboard", onApiScoreboard)
	mux.HandleFunc("/api/scoreboard/pair", onApiScoreboardPair)
	mux.HandleFunc("/api/signup", onApiSignup)
	mux.HandleFunc("/api/login", onApiLogin)
	mux.HandleFunc("/api/logout", onApiLogout)
}

func onApiWs(lobby *Lobby) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			stopCh:   make(chan struct{}),
		}

		// Start the write pump in a separate goroutine
		go client.writePump()

		lobby.register <- client

		// Handle incoming messages (Read Pump)
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
					ghostCount := 4
					if countFloat, ok := msg["ghostCount"].(float64); ok {
						ghostCount = int(countFloat)
					}
					startSinglePlayerGame(client, ghostCount)
				case "update_ghost_count":
					if countFloat, ok := msg["count"].(float64); ok {
						if game := client.GetGame(); game != nil {
							game.UpdateGhostCount(int(countFloat))
							// Broadcast updated gamestate to client immediately
							// Hold read lock to prevent data race with concurrent game.Update()
							game.mu.RLock()
							client.WriteJSON(game)
							game.mu.RUnlock()
						}
					}
				}
			}
		}()
	}
}

func onApiScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !db.RequireDB(w) {
		return
	}

	var req ScoreSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Using ghost count 4 as default/legacy if not provided in JSON or struct yet, 
	// though standard request doesn't have it yet. Will update struct later.
	// For now, assuming default 4 for legacy endpoint use, or we add field to struct.
	if err := db.SaveScore(req.Nickname, req.Score, 4); err != nil {
		fmt.Println("Score update error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Score submitted"})
}

func onApiScoreboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !db.RequireDB(w) {
		return
	}
	
	ghosts := 4
	// Allow query param ?ghosts=N
	if gStr := r.URL.Query().Get("ghosts"); gStr != "" {
		fmt.Sscanf(gStr, "%d", &ghosts)
	}

	scores, err := db.GetTopScores(ghosts)
	if err != nil {
		fmt.Println("Scoreboard query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

func onApiScoreboardPair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !db.RequireDB(w) {
		return
	}

	scores, err := db.GetTopPairScores()
	if err != nil {
		fmt.Println("PairScoreboard query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

func onApiSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !db.RequireDB(w) {
		return
	}
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := db.CreateUser(req.Nickname, req.Password); err != nil {
		fmt.Println("Signup error:", err)
		if errors.Is(err, db.ErrUsernameTaken) {
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
}

func onApiLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !db.RequireDB(w) {
		return
	}
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := db.VerifyUser(req.Nickname, req.Password); err != nil {
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
}

func onApiLogout(w http.ResponseWriter, r *http.Request) {
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

func startSinglePlayerGame(client *Client, ghostCount int) {
	// Similar to old main loop but running in goroutine per client (or associated with client)
	// Actually, client.Game stores the state.

	// If already in game, stop it?
	// Let's assume start_single starts a new one.

	game := NewGame([]string{client.Nickname}, ghostCount)
	client.SetGame(game)

	// Start Ticker Loop for this single player game
	go func() {
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()

		// Notify start
		startMsg := map[string]interface{}{
			"type": "game_start",
			"mode": "single",
			"p1":   client.Nickname,
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
				// Save score with ghost count from game state. 
				// We need to access game.GhostCount.
				if err := db.SaveScore(client.Nickname, score, game.GhostCount); err != nil {
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
