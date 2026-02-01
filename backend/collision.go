package main

func (g *GameState) checkCollisions() {
	collisions := g.getCollisions()
	for _, c := range collisions {
		g.resolveCollision(c.Ghost, c.Player)
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

func isCollision(ghost *Ghost, p *PlayerState) bool {
	// Direct collision or Swap collision
	// Swap collision: Ghost is at Player's old pos, Player is at Ghost's old pos
	return (ghost.Pos == p.Pos) || (ghost.Pos == p.LastPos && ghost.LastPos == p.Pos)
}

func (g *GameState) resolveCollision(ghost *Ghost, p *PlayerState) {
	if g.PowerModeTime > 0 {
		g.Score += 200
		ghost.Pos = Position{X: 9, Y: 8} // Send home
		ghost.LastPos = Position{X: 9, Y: 8}
	} else {
		p.Alive = false // Kill player
	}
}

type CollisionEvent struct {
	Ghost  *Ghost
	Player *PlayerState
}

func (g *GameState) getCollisions() []CollisionEvent {
	var collisions []CollisionEvent
	for i := range g.Ghosts {
		// Use pointer to ghost in the slice so we can modify it later if needed (though getCollisions shouldn't modify)
		// but resolveCollision needs a pointer to the actual ghost in the slice to modify position.
		ghost := &g.Ghosts[i]
		for _, p := range g.Players {
			if !p.Alive {
				continue
			}

			if isCollision(ghost, p) {
				collisions = append(collisions, CollisionEvent{Ghost: ghost, Player: p})
			}
		}
	}
	return collisions
}
