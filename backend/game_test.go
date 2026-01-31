package main

import (
	"testing"
	"time"
)

func TestScoreCalculation(t *testing.T) {
	// Initialize game with one player
	game := NewGame([]string{"tester"})
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

func TestGhostPlayerSwapCollision(t *testing.T) {
	// Initialize game with one player
	game := NewGame([]string{"tester"})
	p := game.Players["tester"]

	// Set positions: Player at (1, 1), Ghost at (1, 2)
	// They are facing each other
	p.Pos = Position{X: 1, Y: 1}
	p.LastPos = Position{X: 1, Y: 1} // Initialize LastPos
	p.Dir = DirDown
	p.NextDir = DirDown

	// Find the first ghost and move it to (1, 2)
	ghost := &game.Ghosts[0]
	ghost.Pos = Position{X: 1, Y: 2}
	ghost.LastPos = Position{X: 1, Y: 2} // Initialize LastPos
	ghost.Dir = DirUp

	// We need to ensure the Ghost moves UP.
	// Since ghosts move somewhat randomly, we might need to force it or mock canMove.
	// However, the game logic for ghosts has some randomness.
	// Let's modify the ghost logic slightly for the test? No, that's bad.
	// In the game loop:
	// 1. movePlayer() -> Player moves to (1, 2). Player.LastPos = (1, 1)
	// 2. moveGhosts() -> Ghost moves to (1, 1). Ghost.LastPos = (1, 2)
	// 3. checkCollisions() -> Player at (1, 2), Ghost at (1, 1).
	//    Swap check: Ghost.Pos (1, 1) == Player.LastPos (1, 1) && Ghost.LastPos (1, 2) == Player.Pos (1, 2)
	//    This should trigger collision.
	
	// BUT, moveGhosts() has randomness.
	// We can manually call moveOneGhost? Or just manually set the new positions and call checkCollisions.
	// Since we are testing checkCollisions logic primarily here.
	
	// Step 1: Simulate movement manually to guarantee the swap state
	p.LastPos = Position{X: 1, Y: 1}
	p.Pos = Position{X: 1, Y: 2}
	
	ghost.LastPos = Position{X: 1, Y: 2}
	ghost.Pos = Position{X: 1, Y: 1}
	
	// Step 2: Run collision check
	game.checkCollisions()
	
	// Step 3: Verify player is dead
	if p.Alive {
		t.Errorf("Expected player to be dead after swap collision, but is alive")
	}
}

func TestDirectCollision(t *testing.T) {
	// Initialize game with one player
	game := NewGame([]string{"tester"})
	p := game.Players["tester"]

	// Set positions: Player at (1, 1), Ghost at (1, 1)
	p.Pos = Position{X: 1, Y: 1}
	ghost := &game.Ghosts[0]
	ghost.Pos = Position{X: 1, Y: 1}
	
	game.checkCollisions()
	
	if p.Alive {
		t.Errorf("Expected player to be dead after direct collision, but is alive")
	}
}
