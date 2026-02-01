package db

import (
	"fmt"
)

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
	// Added ghost_count column defaulting to 4 for existing logic compatibility
	createScoresTableSQL := `CREATE TABLE IF NOT EXISTS scores (
		id SERIAL PRIMARY KEY,
		nickname TEXT NOT NULL,
		score INT NOT NULL,
		ghost_count INT DEFAULT 4,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(nickname, ghost_count)
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

func migrateScoresTable() error {
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
		fmt.Println("Migrating scores table: adding ghost_count column")
		// Add column with default 4
		_, err = db.Exec(`ALTER TABLE scores ADD COLUMN ghost_count INT DEFAULT 4`)
		if err != nil {
			return fmt.Errorf("adding ghost_count column: %w", err)
		}
	}
	
	// Check for the old unique constraint (nickname) and drop it if needed
	// Postgres usually names unique constraints as tablename_columnname_key
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
		fmt.Println("Migrating scores table: dropping old unique constraint scores_nickname_key")
		_, err = db.Exec(`ALTER TABLE scores DROP CONSTRAINT scores_nickname_key`)
		if err != nil {
			return fmt.Errorf("dropping old constraint: %w", err)
		}
	}

	// Ensure new unique constraint (nickname, ghost_count) exists
	// We can try to add it, if it exists it might fail or we can check first. 
	// Easier to just use ADD CONSTRAINT IF NOT EXISTS syntax if using newer Postgres, 
	// but standard way is create unique index.
	
	// Check if the new unique index/constraint exists
	var newConstraintExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE tablename = 'scores' 
			AND indexname = 'scores_nickname_ghost_count_key'
		)`).Scan(&newConstraintExists)

	if !newConstraintExists {
		fmt.Println("Migrating scores table: adding unique constraint (nickname, ghost_count)")
		// It might fail if there are duplicates now (same nickname multiple times with ghost_count=4)
		// We should consolidate duplicates first if any.
		
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
