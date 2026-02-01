package main

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

	if err = createUsersTable(); err != nil {
		fmt.Printf("Error creating users table: %v\n", err)
		closeDB()
		return
	}

	if err = createScoresTable(); err != nil {
		fmt.Printf("Error creating scores table: %v\n", err)
		closeDB()
		return
	}

	if err = createPairScoresTable(); err != nil {
		fmt.Printf("Error creating pair_scores table: %v\n", err)
		closeDB()
		return
	}

	if err = migratePairScores(); err != nil {
		fmt.Printf("Error migrating pair_scores table: %v\n", err)
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

func createUsersTable() error {
	createUsersTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		nickname TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);`
	_, err := db.Exec(createUsersTableSQL)
	return err
}

func createScoresTable() error {
	createScoresTableSQL := `CREATE TABLE IF NOT EXISTS scores (
		id SERIAL PRIMARY KEY,
		nickname TEXT UNIQUE NOT NULL,
		score INT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(createScoresTableSQL)
	return err
}

func createPairScoresTable() error {
	createPairScoresTableV2 := `CREATE TABLE IF NOT EXISTS pair_scores (
		id SERIAL PRIMARY KEY,
		player1 TEXT NOT NULL,
		player2 TEXT NOT NULL,
		score INT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(player1, player2)
	);`

	_, err := db.Exec(createPairScoresTableV2)
	return err
}

func migratePairScores() error {
	// Migration for existing tables: ensure unique constraint exists for (player1, player2).
	// The ON CONFLICT clause in SavePairScore requires this constraint.
	// First, check if the unique index already exists to avoid unnecessary work.
	var indexExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE tablename = 'pair_scores' 
			AND indexname = 'pair_scores_player1_player2_key'
		)`).Scan(&indexExists)
	if err != nil {
		return fmt.Errorf("checking for existing index: %w", err)
	}

	if !indexExists {
		// Before creating the unique index, consolidate any duplicate (player1, player2) entries
		// by keeping only the row with the highest score for each pair.
		// This handles existing databases that may have duplicate entries from before this constraint.
		consolidateDuplicatesSQL := `
			DELETE FROM pair_scores p1
			USING pair_scores p2
			WHERE p1.player1 = p2.player1 
			  AND p1.player2 = p2.player2 
			  AND (p1.score < p2.score OR (p1.score = p2.score AND p1.id > p2.id))
		`
		_, err = db.Exec(consolidateDuplicatesSQL)
		if err != nil {
			return fmt.Errorf("consolidating duplicate pair_scores: %w", err)
		}

		// Now create the unique index - this will succeed since duplicates have been removed
		addPairScoresUniqueConstraint := `CREATE UNIQUE INDEX IF NOT EXISTS pair_scores_player1_player2_key ON pair_scores (player1, player2);`
		_, err = db.Exec(addPairScoresUniqueConstraint)
		if err != nil {
			return fmt.Errorf("adding unique constraint to pair_scores: %w", err)
		}
		fmt.Println("Migrated pair_scores table: consolidated duplicates and added unique constraint")
	}
	return nil
}

func SaveScore(nickname string, score int) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	upsertSQL := `
		INSERT INTO scores (nickname, score, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (nickname)
		DO UPDATE SET score = EXCLUDED.score, updated_at = CURRENT_TIMESTAMP
		WHERE scores.score < EXCLUDED.score
	`
	_, err := db.Exec(upsertSQL, nickname, score)
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

func GetTopScores() ([]ScoreEntry, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	rows, err := db.Query("SELECT nickname, score FROM scores ORDER BY score DESC LIMIT 10")
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
