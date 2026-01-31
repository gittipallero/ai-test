package main

const (
	Rows = 21
	Cols = 19
)

const (
	DirUp    Direction = "UP"
	DirDown  Direction = "DOWN"
	DirLeft  Direction = "LEFT"
	DirRight Direction = "RIGHT"
)

var InitialPacman = Position{X: 9, Y: 15}
var InitialGhosts = []Ghost{
	{ID: 1, Pos: Position{X: 9, Y: 7}, Dir: DirLeft, Color: "red"},
	{ID: 2, Pos: Position{X: 9, Y: 8}, Dir: DirRight, Color: "pink"},
	{ID: 3, Pos: Position{X: 10, Y: 7}, Dir: DirUp, Color: "cyan"},
	{ID: 4, Pos: Position{X: 10, Y: 8}, Dir: DirDown, Color: "orange"},
}
