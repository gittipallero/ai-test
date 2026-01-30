package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type ScoreResponse struct {
	HighScore int `json:"highScore"`
}

type User struct {
	ID           int    `json:"id"`
	Nickname     string `json:"nickname"`
	PasswordHash string `json:"-"`
}

type AuthRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

var db *sql.DB

func requireDB(w http.ResponseWriter) bool {
	if db == nil {
		http.Error(w, "Database not configured", http.StatusServiceUnavailable)
		return false
	}
	return true
}

func initDB() {
	var err error
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "require" // Default to require for security
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), sslMode)

	// Fallback for local testing if env vars not set
	if os.Getenv("DB_HOST") == "" {
		fmt.Println("Warning: DB_HOST not set, skipping DB init")
		db = nil
		return
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Error connecting to DB: %v\n", err)
		db = nil
		return
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("Error pinging DB: %v\n", err)
		_ = db.Close()
		db = nil
		return
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		nickname TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		fmt.Printf("Error creating table: %v\n", err)
	}
	fmt.Println("Database initialized successfully")
}

func main() {
	initDB()
	mux := http.NewServeMux()

	// Serve static files from frontend/dist
	rootDir := "."
	if _, err := os.Stat("frontend/dist"); os.IsNotExist(err) {
		rootDir = ".."
	}

	fs := http.FileServer(http.Dir(filepath.Join(rootDir, "frontend/dist")))
	mux.Handle("/", fs)

	mux.HandleFunc("/api/score", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ScoreResponse{HighScore: 1000})
	})

	mux.HandleFunc("/api/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !requireDB(w) {
			return
		}
		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO users (nickname, password_hash) VALUES ($1, $2)", req.Nickname, string(hashedPassword))
		if err != nil {
			fmt.Println("Signup error:", err)
			http.Error(w, "Username already taken or database error", http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "User created", "nickname": req.Nickname})
	})

	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !requireDB(w) {
			return
		}
		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var storedHash string
		err := db.QueryRow("SELECT password_hash FROM users WHERE nickname=$1", req.Nickname).Scan(&storedHash)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

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
