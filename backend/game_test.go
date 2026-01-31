package main

import (
	"testing"
	"time"
)

func TestScoreCalculation(t *testing.T) {
	game := NewGame()
	
	// Mock a scenario where Pacman eats a dot
	// Set Pacman position next to a dot
	// There is a dot at (1, 1) in initial map
	game.Pacman = Position{X: 1, Y: 0}
	game.NextDirection = DirDown
	game.Direction = DirDown
	
	// Cheat time to be 10 seconds ago to ensure 0 bonus first
	game.LastEatTime = time.Now().Add(-10 * time.Second).UnixMilli()
	
	game.movePacman()

	if game.Score != 10 {
		t.Errorf("Expected score 10 for slow eat, got %d", game.Score)
	}

	// Now move to another dot immediately
	// (1, 2) has a dot. Pacman is now at (1, 1)
	game.NextDirection = DirDown
	game.Direction = DirDown
	
	// We just ate, so LastEatTime was updated.
	// Let's force it to be very recent (0ms ago) to get max points
	game.LastEatTime = time.Now().UnixMilli()
	
	game.movePacman()
	
	// Should be 10 (base) + 10 (previous) + 100 (max bonus) = 120
	if game.Score != 120 {
		t.Errorf("Expected score 120 for fast eat, got %d", game.Score)
	}
	
	// Move to (1, 3) with 500ms delay
	// (1, 3) has a dot. Pacman is at (1, 2)
	game.NextDirection = DirDown
	game.Direction = DirDown
	
	game.LastEatTime = time.Now().Add(-500 * time.Millisecond).UnixMilli()
	
	game.movePacman()
	
	// Bonus should be 100 - (500/10) = 50.
	// Previous score 120. + 10 base + 50 bonus = 180
	if game.Score != 180 {
		t.Errorf("Expected score 180 for medium eat, got %d", game.Score)
	}
}
