package main

import (
	"sync"
)

// --- Game Types ---

type Direction string

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Ghost struct {
	ID    int       `json:"id"`
	Pos   Position  `json:"pos"`
	Dir   Direction `json:"dir"`
	Color string    `json:"color"`
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
	LastEatTime   int64           `json:"lastEatTime"`
	mu            sync.RWMutex
}

// --- Auth & API Types ---

type User struct {
	ID           int    `json:"id"`
	Nickname     string `json:"nickname"`
	PasswordHash string `json:"-"`
}

type AuthRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type ScoreSubmitRequest struct {
	Nickname string `json:"nickname"`
	Score    int    `json:"score"`
}

type ScoreEntry struct {
	Nickname string `json:"nickname"`
	Score    int    `json:"score"`
}
