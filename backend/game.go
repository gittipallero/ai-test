package main

import (
	"math/rand"
	"sync"
	"time"
)

type PlayerState struct {
	Nickname string   `json:"nickname"`
	Pos      Position `json:"pos"`
	Dir      Direction `json:"dir"` // Current movement direction
	NextDir  Direction `json:"nextDir"` // Buffered next direction
	Alive    bool     `json:"alive"`
}

type GameState struct {
	Grid          [Rows][Cols]int         `json:"grid"`
	Players       map[string]*PlayerState `json:"players"`
	Ghosts        []Ghost                 `json:"ghosts"`
	Score         int                     `json:"score"`
	PowerModeTime int                     `json:"powerModeTime"`
	LastEatTime   int64                   `json:"lastEatTime"`
	GameOver      bool                    `json:"gameOver"`
	mu            sync.RWMutex            `json:"-"`
}

func NewGame(nicknames []string) *GameState {
	// Deep copy grid
	var grid [Rows][Cols]int
	for y := 0; y < Rows; y++ {
		for x := 0; x < Cols; x++ {
			grid[y][x] = InitialMap[y][x]
		}
	}

	players := make(map[string]*PlayerState)
	
	// Single player default position
	startPositions := []Position{
		{X: 9, Y: 15}, // Player 1
		{X: 10, Y: 15}, // Player 2 (slightly offset)
	}

	for i, nick := range nicknames {
		pos := startPositions[0]
		if i < len(startPositions) {
			pos = startPositions[i]
		}
		
		players[nick] = &PlayerState{
			Nickname: nick,
			Pos:      pos,
			Dir:      "",
			NextDir:  "",
			Alive:    true,
		}
	}

	game := &GameState{
		Grid:          grid,
		Players:       players,
		Ghosts:        make([]Ghost, len(InitialGhosts)),
		Score:         0,
		PowerModeTime: 0,
		LastEatTime:   time.Now().UnixMilli(),
		GameOver:      false,
	}
	copy(game.Ghosts, InitialGhosts)
	return game
}

func (g *GameState) SetNextDirection(nickname string, dir Direction) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	if p, ok := g.Players[nickname]; ok && p.Alive {
		p.NextDir = dir
	}
}

func (g *GameState) Update() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.GameOver {
		return
	}

	// Move all alive players
    activePlayers := 0
	for _, p := range g.Players {
		if p.Alive {
            activePlayers++
			g.movePlayer(p)
		}
	}
    
    // If all players dead, game over
    if activePlayers == 0 {
        g.GameOver = true
        return
    }

	g.moveGhosts()
	g.checkCollisions()

	if g.PowerModeTime > 0 {
		g.PowerModeTime -= 150
	}
}

func (g *GameState) movePlayer(p *PlayerState) {
	currentDir := p.Dir

	if p.NextDir != "" && g.canMove(p.Pos, p.NextDir) {
		currentDir = p.NextDir
		p.Dir = p.NextDir
	}

	if currentDir != "" && g.canMove(p.Pos, currentDir) {
		newPos := g.getNextPos(p.Pos, currentDir)
		newPos = g.handleTeleport(newPos)
		g.handleEating(newPos)
		p.Pos = newPos
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
	if cell == CellDot {
		g.Grid[pos.Y][pos.X] = CellEmpty

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
	if cell == CellPower {
		g.Grid[pos.Y][pos.X] = CellEmpty
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
        for _, p := range g.Players {
            if !p.Alive {
                continue
            }
            if ghost.Pos == p.Pos {
                // Collision
                if g.PowerModeTime > 0 {
                    g.Score += 200
                    g.Ghosts[i].Pos = Position{X: 9, Y: 8} // Send home
                } else {
                    p.Alive = false // Kill player
                }
            }
        }
	}
    
    // Check if any players alive
    anyAlive := false
    for _, p := range g.Players {
        if p.Alive {
            anyAlive = true
            break
        }
    }
    if !anyAlive {
        g.GameOver = true
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
	return g.Grid[next.Y][next.X] != CellWall
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
