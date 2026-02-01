package main

import (
	"testing"
	"time"
)

func TestScoreCalculation(t *testing.T) {
	// Initialize game with one player, 4 ghosts
	game := NewGame([]string{"tester"}, 4)
	p := game.Players["tester"]
	
	// Mock a scenario where Pacman eats a dot
	// Set Pacman position next to a dot
	// There is a dot at (1, 1) in initial map
	p.Pos = Position{X: 1, Y: 0}
	p.NextDir = DirDown
	p.Dir = DirDown
	
	// Cheat time to be 10 seconds ago to ensure 0 bonus first
	game.LastEatTime = time.Now().Add(-10 * time.Second).UnixMilli()
	
	game.movePlayer(p)

	if game.Score != 10 {
		t.Errorf("Expected score 10 for slow eat, got %d", game.Score)
	}

	// Now move to another dot immediately
	// (1, 2) has a dot. Pacman is now at (1, 1)
	p.NextDir = DirDown
	p.Dir = DirDown
	
	// We just ate, so LastEatTime was updated.
	// Let's force it to be very recent (0ms ago) to get max points
	game.LastEatTime = time.Now().UnixMilli()
	
	game.movePlayer(p)
	
	// Should be 10 (base) + 10 (previous) + 100 (max bonus) = 120
	// Wait, did I update movePlayer to affect game.Score? Yes.
	if game.Score != 120 {
		t.Errorf("Expected score 120 for fast eat, got %d", game.Score)
	}
	
	// Move to (1, 3) with 500ms delay
	// (1, 3) has a dot. Pacman is at (1, 2)
	p.NextDir = DirDown
	p.Dir = DirDown
	
	game.LastEatTime = time.Now().Add(-500 * time.Millisecond).UnixMilli()
	
	game.movePlayer(p)
	
	// Bonus should be 100 - (500/10) = 50.
	// Previous score 120. + 10 base + 50 bonus = 180
	if game.Score != 180 {
		t.Errorf("Expected score 180 for medium eat, got %d", game.Score)
	}
}


