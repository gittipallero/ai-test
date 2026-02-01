package db

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
)

type Migration struct {
	ID   int
	Name string
	Run  func(*sql.DB) error
}

var migrations = []Migration{
	{
		ID:   1,
		Name: "InitSchema",
		Run:  initSchema,
	},
	{
		ID:   2,
		Name: "AddGhostCountToScores",
		Run:  addGhostCountToScores,
	},
	{
		ID:   3,
		Name: "FixPairScoresConstraint",
		Run:  fixPairScoresConstraint,
	},
}

func ensureSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		id INT PRIMARY KEY,
		name TEXT NOT NULL,
		run_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

func RunMigrations(db *sql.DB) error {
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return fmt.Errorf("ensuring migrations table: %w", err)
	}

	// Get applied migrations
	rows, err := db.Query("SELECT id FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("fetching migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		applied[id] = true
	}

	// Sort migrations just in case
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	for _, m := range migrations {
		if applied[m.ID] {
			continue
		}

		log.Printf("Running migration %d: %s", m.ID, m.Name)
		if err := m.Run(db); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", m.ID, m.Name, err)
		}

		_, err := db.Exec("INSERT INTO schema_migrations (id, name) VALUES ($1, $2)", m.ID, m.Name)
		if err != nil {
			return fmt.Errorf("recording migration %d: %w", m.ID, err)
		}
	}

	return nil
}

// Migration 1: Initial Schema
func initSchema(db *sql.DB) error {
	createUsersTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		nickname TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);`
	if _, err := db.Exec(createUsersTableSQL); err != nil {
		return fmt.Errorf("creating users table: %w", err)
	}

	createScoresTableSQL := `CREATE TABLE IF NOT EXISTS scores (
		id SERIAL PRIMARY KEY,
		nickname TEXT NOT NULL,
		score INT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createScoresTableSQL); err != nil {
		return fmt.Errorf("creating scores table: %w", err)
	}
	
	createPairScoresTableV2 := `CREATE TABLE IF NOT EXISTS pair_scores (
		id SERIAL PRIMARY KEY,
		player1 TEXT NOT NULL,
		player2 TEXT NOT NULL,
		score INT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createPairScoresTableV2); err != nil {
		return fmt.Errorf("creating pair_scores table: %w", err)
	}
	
	return nil
}

// Migration 2: Add Ghost Count
func addGhostCountToScores(db *sql.DB) error {
	// Check if ghost_count column exists
	var colExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name='scores' AND column_name='ghost_count'
		)`).Scan(&colExists)
	if err != nil {
		return fmt.Errorf("checking for ghost_count column: %w", err)
	}

	if !colExists {
		// Add column with default 4
		_, err = db.Exec(`ALTER TABLE scores ADD COLUMN ghost_count INT DEFAULT 4`)
		if err != nil {
			return fmt.Errorf("adding ghost_count column: %w", err)
		}
	}
	
	// Check for the old unique constraint (nickname) and drop it if needed
	var constraintExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_constraint 
			WHERE conname = 'scores_nickname_key'
		)`).Scan(&constraintExists)
	if err != nil {
		return fmt.Errorf("checking for old unique constraint: %w", err)
	}

	if constraintExists {
		_, err = db.Exec(`ALTER TABLE scores DROP CONSTRAINT scores_nickname_key`)
		if err != nil {
			return fmt.Errorf("dropping old constraint: %w", err)
		}
	}

	// Check if the new unique index/constraint exists
	var newConstraintExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE tablename = 'scores' 
			AND indexname = 'scores_nickname_ghost_count_key'
		)`).Scan(&newConstraintExists)
	if err != nil {
		return fmt.Errorf("checking for new unique index: %w", err)
	}

	if !newConstraintExists {
		consolidateSQL := `
			DELETE FROM scores s1
			USING scores s2
			WHERE s1.nickname = s2.nickname
			  AND s1.ghost_count = s2.ghost_count
			  AND (s1.score < s2.score OR (s1.score = s2.score AND s1.id > s2.id))
		`
		_, err = db.Exec(consolidateSQL)
		if err != nil {
			return fmt.Errorf("consolidating duplicate scores: %w", err)
		}

		_, err = db.Exec(`CREATE UNIQUE INDEX scores_nickname_ghost_count_key ON scores (nickname, ghost_count)`)
		if err != nil {
			return fmt.Errorf("creating new unique index: %w", err)
		}
	}
	return nil
}

// Migration 3: Pair Scores Constraint
func fixPairScoresConstraint(db *sql.DB) error {
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

		addPairScoresUniqueConstraint := `CREATE UNIQUE INDEX IF NOT EXISTS pair_scores_player1_player2_key ON pair_scores (player1, player2);`
		_, err = db.Exec(addPairScoresUniqueConstraint)
		if err != nil {
			return fmt.Errorf("adding unique constraint to pair_scores: %w", err)
		}
	}
	return nil
}
