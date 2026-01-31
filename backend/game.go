package main

import (
	"math/rand"
	"time"
)

func NewGame() *GameState {
	// Deep copy grid
	var grid [Rows][Cols]int
	for y := 0; y < Rows; y++ {
		for x := 0; x < Cols; x++ {
			grid[y][x] = InitialMap[y][x]
		}
	}

	game := &GameState{
		Grid:          grid,
		Pacman:        InitialPacman,
		Ghosts:        make([]Ghost, len(InitialGhosts)),
		Score:         0,
		PowerModeTime: 0,
		LastEatTime:   time.Now().UnixMilli(),
		GameOver:      false,
	}
	copy(game.Ghosts, InitialGhosts)
	return game
}

func (g *GameState) SetNextDirection(dir Direction) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.NextDirection = dir
}

func (g *GameState) Update() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.GameOver {
		return
	}

	g.movePacman()
	g.moveGhosts()
	g.checkCollisions()

	if g.PowerModeTime > 0 {
		g.PowerModeTime -= 150
	}
}

func (g *GameState) movePacman() {
	currentDir := g.Direction

	if g.NextDirection != "" && g.canMove(g.Pacman, g.NextDirection) {
		currentDir = g.NextDirection
		g.Direction = g.NextDirection
	}

	if currentDir != "" && g.canMove(g.Pacman, currentDir) {
		newPos := g.getNextPos(g.Pacman, currentDir)
		newPos = g.handleTeleport(newPos)
		g.handleEating(newPos)
		g.Pacman = newPos
	}
}

func (g *GameState) handleTeleport(pos Position) Position {
	if pos.X < 0 {
		pos.X = Cols - 1
	} else if pos.X >= Cols {
		pos.X = 0
	}
	return pos
}

func (g *GameState) handleEating(pos Position) {
	// Eat Dot
	cell := g.Grid[pos.Y][pos.X]
	if cell == 2 {
		g.Grid[pos.Y][pos.X] = 0

		// Calculate time-dependent bonus
		now := time.Now().UnixMilli()
		timeDiff := now - g.LastEatTime
		bonus := 0
		if timeDiff < 1000 {
			bonus = int(100 - (timeDiff / 10))
			if bonus < 0 {
				bonus = 0
			}
		}
		g.Score += 10 + bonus
		g.LastEatTime = now
	}
	// Eat Power
	if cell == 3 {
		g.Grid[pos.Y][pos.X] = 0
		g.Score += 50
		g.PowerModeTime = 5000
	}
}

func (g *GameState) moveGhosts() {
	// Seed random if not seeded (better to do in init, but for safety)
	if rand.Int() == 0 {
		rand.Seed(time.Now().UnixNano())
	}

	for i := range g.Ghosts {
		ghost := &g.Ghosts[i]
		g.moveOneGhost(ghost)
	}
}

func (g *GameState) moveOneGhost(ghost *Ghost) {
	validDirs := g.getValidGhostDirs(ghost)

	// Don't reverse immediately if possible
	reverseDir := getReverseDir(ghost.Dir)
	if len(validDirs) > 1 && ghost.Dir != "" {
		var nonReverse []Direction
		for _, d := range validDirs {
			if d != reverseDir {
				nonReverse = append(nonReverse, d)
			}
		}
		if len(nonReverse) > 0 {
			validDirs = nonReverse
		}
	}

	nextDir := ghost.Dir
	// Randomly change direction or if stuck
	if ghost.Dir == "" || !g.canMove(ghost.Pos, ghost.Dir) || rand.Float64() < 0.2 {
		if len(validDirs) > 0 {
			nextDir = validDirs[rand.Intn(len(validDirs))]
		}
	}

	if nextDir != "" && g.canMove(ghost.Pos, nextDir) {
		newPos := g.getNextPos(ghost.Pos, nextDir)
		newPos = g.handleTeleport(newPos)
		ghost.Pos = newPos
		ghost.Dir = nextDir
	}
}

func (g *GameState) getValidGhostDirs(ghost *Ghost) []Direction {
	possibleDirs := []Direction{DirUp, DirDown, DirLeft, DirRight}
	var validDirs []Direction
	for _, d := range possibleDirs {
		if g.canMove(ghost.Pos, d) {
			validDirs = append(validDirs, d)
		}
	}
	return validDirs
}

func (g *GameState) checkCollisions() {
	for i, ghost := range g.Ghosts {
		if ghost.Pos == g.Pacman {
			// Collision
			if g.PowerModeTime > 0 {
				g.Score += 200
				g.Ghosts[i].Pos = Position{X: 9, Y: 8} // Send home
			} else {
				g.GameOver = true
			}
		}
	}
}

func (g *GameState) canMove(pos Position, dir Direction) bool {
	next := g.getNextPos(pos, dir)
	if next.Y < 0 || next.Y >= Rows {
		return false
	}
	if next.X < 0 || next.X >= Cols {
		return true // Tunnel
	}
	return g.Grid[next.Y][next.X] != 1
}

func (g *GameState) getNextPos(pos Position, dir Direction) Position {
	newPos := pos
	switch dir {
	case DirUp:
		newPos.Y--
	case DirDown:
		newPos.Y++
	case DirLeft:
		newPos.X--
	case DirRight:
		newPos.X++
	}
	return newPos
}

func getReverseDir(dir Direction) Direction {
	switch dir {
	case DirUp:
		return DirDown
	case DirDown:
		return DirUp
	case DirLeft:
		return DirRight
	case DirRight:
		return DirLeft
	}
	return ""
}
