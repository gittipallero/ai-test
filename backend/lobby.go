package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Nickname string
	Conn     *websocket.Conn
	Send     chan []byte
	Lobby    *Lobby
	Game     *GameState // Nil if in lobby/waiting
}

type Lobby struct {
	clients    map[*Client]bool
	waiting    []*Client
	games      map[*GameState]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.Mutex
}

func NewLobby() *Lobby {
	return &Lobby{
		clients:    make(map[*Client]bool),
		waiting:    make([]*Client, 0),
		games:      make(map[*GameState]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

func (l *Lobby) Run() {
	for {
		select {
		case client := <-l.register:
			l.mu.Lock()
			l.clients[client] = true
			l.mu.Unlock()
			log.Printf("Client registered: %s", client.Nickname)
			l.BroadcastPlayerCount()

		case client := <-l.unregister:
			l.mu.Lock()
			if _, ok := l.clients[client]; ok {
				delete(l.clients, client)
				close(client.Send)
			}
			// Remove from waiting list if present
			for i, c := range l.waiting {
				if c == client {
					l.waiting = append(l.waiting[:i], l.waiting[i+1:]...)
					break
				}
			}
			l.mu.Unlock()
			log.Printf("Client unregistered: %s", client.Nickname)
			l.BroadcastPlayerCount()

		case message := <-l.broadcast:
			l.mu.Lock()
			for client := range l.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(l.clients, client)
				}
			}
			l.mu.Unlock()
		}
	}
}

func (l *Lobby) JoinPairQueue(client *Client) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if already waiting
	for _, c := range l.waiting {
		if c == client {
			return
		}
	}

	l.waiting = append(l.waiting, client)
	log.Printf("%s joined pair queue. Queue length: %d", client.Nickname, len(l.waiting))

	if len(l.waiting) >= 2 {
		// Match found!
		player1 := l.waiting[0]
		player2 := l.waiting[1]
		l.waiting = l.waiting[2:]

		l.StartPairGame(player1, player2)
	} else {
		// Notify client they are waiting
		msg := map[string]interface{}{
			"type": "waiting",
		}
        // Use a goroutine to avoid blocking the lock
        go func() {
            if err := client.Conn.WriteJSON(msg); err != nil {
                log.Printf("Error sending wait message: %v", err)
            }
        }()
	}
}

func (l *Lobby) StartPairGame(p1, p2 *Client) {
	log.Printf("Starting pair game for %s and %s", p1.Nickname, p2.Nickname)
    
    // Create new game with two players
	game := NewGame([]string{p1.Nickname, p2.Nickname})
	l.games[game] = true
    
    p1.Game = game
    p2.Game = game

	// Start game loop
	go func() {
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()

        // Notify start
        startMsg := map[string]interface{}{
            "type": "game_start", 
            "mode": "pair",
            "p1": p1.Nickname,
            "p2": p2.Nickname,
        }
        p1.Conn.WriteJSON(startMsg)
        p2.Conn.WriteJSON(startMsg)

		for range ticker.C {
			game.Update()

			game.mu.RLock()
            // Check if game is over (both dead)
            if game.GameOver {
                 game.mu.RUnlock()
                 // Send final state
                 game.mu.RLock()
                 state := game
                 game.mu.RUnlock()
                 
                 p1.Conn.WriteJSON(state)
                 p2.Conn.WriteJSON(state)

                 // Save Score
                 SavePairScore(p1.Nickname, p2.Nickname, state.Score)

                 // Clean up references
                 p1.Game = nil
                 p2.Game = nil
                 
                 l.mu.Lock()
                 delete(l.games, game)
                 l.mu.Unlock()
                 return
            }
            
			// Broadcast state to both players
            // We need to marshal it once
            // Actually, WriteJSON does marshal.
            // Let's rely on that for now, minimal optimization needed.
            
            // NOTE: We are sending the WHOLE state. P1 and P2 need to know which Pacman they are.
            // But the client can just check the "Players" map by their nickname.
            
			err1 := p1.Conn.WriteJSON(game)
			err2 := p2.Conn.WriteJSON(game)
            game.mu.RUnlock()

			if err1 != nil || err2 != nil {
				log.Println("Error writing to client in pair game, ending game")
				game.mu.Lock()
				game.GameOver = true // Stop updates
				game.mu.Unlock()
				// In a real app we might handle reconnection or pause
				break
			}
		}
	}()
}

func (l *Lobby) BroadcastPlayerCount() {
    // This is just a helper to let clients know how many people are online
    // to show/hide the "Pair Mode" button ideally
    l.mu.Lock()
    defer l.mu.Unlock()

    count := len(l.clients)
    if count < 2 {
        // If 0 or 1, no pair mode possible really (unless waiting for someone)
    }
    msg := map[string]interface{}{
        "type": "lobby_stats",
        "online_count": count,
    }

    for client := range l.clients {
         client.Conn.WriteJSON(msg)
    }
}
