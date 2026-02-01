package main

// Game Types

type Direction string

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Ghost struct {
	ID      int       `json:"id"`
	Pos     Position  `json:"pos"`
	LastPos Position  `json:"-"` // Internal use for collision
	Dir     Direction `json:"dir"`
	Color   string    `json:"color"`
}

// Request/Response types for API
type ScoreSubmitRequest struct {
	Nickname string `json:"nickname"`
	Score    int    `json:"score"`
}

type AuthRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}
