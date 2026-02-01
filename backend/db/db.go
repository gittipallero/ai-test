package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

var ErrUsernameTaken = errors.New("username already taken")

// ScoreEntry represents a row in the scoreboard
type ScoreEntry struct {
	Nickname string `json:"nickname"`
	Score    int    `json:"score"`
}

// PairScoreEntry represents a row in the pair scoreboard
type PairScoreEntry struct {
	Player1 string `json:"player1"`
	Player2 string `json:"player2"`
	Score   int    `json:"score"`
}

func RequireDB(w http.ResponseWriter) bool {
	if db == nil {
		http.Error(w, "Database not configured", http.StatusServiceUnavailable)
		return false
	}
	return true
}

func InitDB() {
	var err error

	if err = connectDB(); err != nil {
		fmt.Println(err)
		return
	}

	if err = RunMigrations(db); err != nil {
		fmt.Printf("Error running migrations: %v\n", err)
		closeDB()
		return
	}

	fmt.Println("Database initialized successfully")
}

func closeDB() {
	if db != nil {
		_ = db.Close()
		db = nil
	}
}

func connectDB() error {
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "require" // Default to require for security
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), sslMode)

	// Fallback for local testing if env vars not set
	if os.Getenv("DB_HOST") == "" {
		closeDB() // Reset any existing connection when DB is not configured
		return fmt.Errorf("Warning: DB_HOST not set, skipping DB init")
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("Error connecting to DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		closeDB()
		return fmt.Errorf("Error pinging DB: %v", err)
	}
	return nil
}

func SaveScore(nickname string, score int, ghostCount int) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// Default to 4 ghosts if not specified (legacy logic safety, though new callers should pass it)
	if ghostCount <= 0 {
		ghostCount = 4
	}

	upsertSQL := `
		INSERT INTO scores (nickname, score, ghost_count, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (nickname, ghost_count)
		DO UPDATE SET score = EXCLUDED.score, updated_at = CURRENT_TIMESTAMP
		WHERE scores.score < EXCLUDED.score
	`
	_, err := db.Exec(upsertSQL, nickname, score, ghostCount)
	return err
}

func SavePairScore(player1, player2 string, score int) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Ensure alphabetical order for consistency in scoreboard
	if player1 > player2 {
		player1, player2 = player2, player1
	}

	// Upsert: only store the best score for each pair
	upsertSQL := `
		INSERT INTO pair_scores (player1, player2, score, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (player1, player2)
		DO UPDATE SET score = EXCLUDED.score, updated_at = CURRENT_TIMESTAMP
		WHERE pair_scores.score < EXCLUDED.score
	`
	_, err := db.Exec(upsertSQL, player1, player2, score)
	return err
}

func GetTopScores(ghostCount int) ([]ScoreEntry, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	if ghostCount <= 0 {
		ghostCount = 4
	}

	rows, err := db.Query("SELECT nickname, score FROM scores WHERE ghost_count = $1 ORDER BY score DESC LIMIT 10", ghostCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []ScoreEntry
	for rows.Next() {
		var entry ScoreEntry
		if err := rows.Scan(&entry.Nickname, &entry.Score); err != nil {
			continue
		}
		scores = append(scores, entry)
	}
	return scores, nil
}

func GetTopPairScores() ([]PairScoreEntry, error) {
    if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
    // Simple top list
	rows, err := db.Query("SELECT player1, player2, score FROM pair_scores ORDER BY score DESC LIMIT 10")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []PairScoreEntry
	for rows.Next() {
		var entry PairScoreEntry
		if err := rows.Scan(&entry.Player1, &entry.Player2, &entry.Score); err != nil {
			continue
		}
		scores = append(scores, entry)
	}
	return scores, nil
}

func CreateUser(nickname, password string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = db.Exec("INSERT INTO users (nickname, password_hash) VALUES ($1, $2)", nickname, string(hashedPassword))
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUsernameTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

func VerifyUser(nickname, password string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE nickname=$1", nickname).Scan(&storedHash)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
}
