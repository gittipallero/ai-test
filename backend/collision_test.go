package main

import (
	"testing"
)

func TestGhostPlayerSwapCollision(t *testing.T) {
	// Initialize game with one player, 4 ghosts
	game := NewGame([]string{"tester"}, 4)
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
	// Initialize game with one player, 4 ghosts
	game := NewGame([]string{"tester"}, 4)
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
