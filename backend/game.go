package main

import (
	"math/rand"
	"sync"
	"time"
)

const (
	Rows = 21
	Cols = 19
)

type Direction string

const (
	DirUp    Direction = "UP"
	DirDown  Direction = "DOWN"
	DirLeft  Direction = "LEFT"
	DirRight Direction = "RIGHT"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Ghost struct {
	ID    int      `json:"id"`
	Pos   Position `json:"pos"`
	Dir   Direction `json:"dir"`
	Color string   `json:"color"`
}

type GameState struct {
	Grid          [Rows][Cols]int `json:"grid"`
	Pacman        Position        `json:"pacman"`
	Ghosts        []Ghost         `json:"ghosts"`
	Score         int             `json:"score"`
	GameOver      bool            `json:"gameOver"`
	PowerModeTime int             `json:"powerModeTime"`
	Direction     Direction       `json:"direction"`
	NextDirection Direction       `json:"nextDirection"`
	mu            sync.RWMutex
}

var InitialMap = [Rows][Cols]int{
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 2, 1, 1, 1, 2, 1, 2, 1, 1, 1, 2, 1, 1, 2, 1},
	{1, 2, 1, 1, 2, 1, 1, 1, 2, 1, 2, 1, 1, 1, 2, 1, 1, 2, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 2, 1, 2, 1, 1, 1, 1, 1, 2, 1, 2, 1, 1, 2, 1},
	{1, 2, 2, 2, 2, 1, 2, 2, 2, 1, 2, 2, 2, 1, 2, 2, 2, 2, 1},
	{1, 1, 1, 1, 2, 1, 1, 1, 0, 1, 0, 1, 1, 1, 2, 1, 1, 1, 1},
	{0, 0, 0, 1, 2, 1, 0, 0, 0, 0, 0, 0, 0, 1, 2, 1, 0, 0, 0},
	{1, 1, 1, 1, 2, 1, 0, 1, 1, 9, 1, 1, 0, 1, 2, 1, 1, 1, 1},
	{0, 2, 2, 2, 2, 0, 0, 1, 0, 0, 0, 1, 0, 0, 2, 2, 2, 2, 0},
	{1, 1, 1, 1, 2, 1, 0, 1, 1, 1, 1, 1, 0, 1, 2, 1, 1, 1, 1},
	{0, 0, 0, 1, 2, 1, 0, 0, 0, 0, 0, 0, 0, 1, 2, 1, 0, 0, 0},
	{1, 1, 1, 1, 2, 1, 2, 1, 1, 1, 1, 1, 2, 1, 2, 1, 1, 1, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 2, 1, 1, 1, 2, 1, 2, 1, 1, 1, 2, 1, 1, 2, 1},
	{1, 2, 2, 1, 2, 2, 2, 2, 2, 0, 2, 2, 2, 2, 2, 1, 2, 2, 1},
	{1, 1, 2, 1, 2, 1, 2, 1, 1, 1, 1, 1, 2, 1, 2, 1, 2, 1, 1},
	{1, 2, 2, 2, 2, 1, 2, 2, 2, 1, 2, 2, 2, 1, 2, 2, 2, 2, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
}

var InitialPacman = Position{X: 9, Y: 15}
var InitialGhosts = []Ghost{
	{ID: 1, Pos: Position{X: 9, Y: 7}, Dir: DirLeft, Color: "red"},
	{ID: 2, Pos: Position{X: 9, Y: 8}, Dir: DirRight, Color: "pink"},
	{ID: 3, Pos: Position{X: 10, Y: 7}, Dir: DirUp, Color: "cyan"},
	{ID: 4, Pos: Position{X: 10, Y: 8}, Dir: DirDown, Color: "orange"},
}

func NewGame() *GameState {
	// Deep copy grid
	var grid [Rows][Cols]int
	for y := 0; y < Rows; y++ {
		for x := 0; x < Cols; x++ {
			grid[y][x] = InitialMap[y][x]
		}
	}

	game := &GameState{
		Grid:   grid,
		Pacman: InitialPacman,
		Ghosts: make([]Ghost, len(InitialGhosts)),
		Score:  0,
		PowerModeTime: 0,
		GameOver: false,
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

		// Teleport
		if newPos.X < 0 {
			newPos.X = Cols - 1
		} else if newPos.X >= Cols {
			newPos.X = 0
		}

		// Eat Dot
		cell := g.Grid[newPos.Y][newPos.X]
		if cell == 2 {
			g.Grid[newPos.Y][newPos.X] = 0
			g.Score += 10
		}
		// Eat Power
		if cell == 3 {
			g.Grid[newPos.Y][newPos.X] = 0
			g.Score += 50
			g.PowerModeTime = 5000
		}

		g.Pacman = newPos
	}
}

func (g *GameState) moveGhosts() {
	// Seed random if not seeded (better to do in init, but for safety)
	if rand.Int() == 0 {
		rand.Seed(time.Now().UnixNano())
	}
	
	for i := range g.Ghosts {
		ghost := &g.Ghosts[i]
		
		possibleDirs := []Direction{DirUp, DirDown, DirLeft, DirRight}
		var validDirs []Direction
		for _, d := range possibleDirs {
			if g.canMove(ghost.Pos, d) {
				validDirs = append(validDirs, d)
			}
		}

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
			// Teleport
			if newPos.X < 0 {
				newPos.X = Cols - 1
			} else if newPos.X >= Cols {
				newPos.X = 0
			}
			ghost.Pos = newPos
			ghost.Dir = nextDir
		}
	}
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
