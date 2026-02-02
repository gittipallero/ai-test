package main

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/villepalo/pacman-go-react/db"
)

type Client struct {
	Nickname string
	Conn     *websocket.Conn
	Send     chan []byte
	Lobby    *Lobby
	mu       sync.RWMutex
	Game     *GameState // Nil if in lobby/waiting; guarded by mu
	stopCh   chan struct{}
}

func (c *Client) GetGame() *GameState {
	c.mu.RLock()
	game := c.Game
	c.mu.RUnlock()
	return game
}

func (c *Client) SetGame(game *GameState) {
	c.mu.Lock()
	c.Game = game
	c.mu.Unlock()
}

type Lobby struct {
	clients    map[*Client]bool
	waiting    []*Client
	games      map[*GameState]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

var (
	errClientDisconnected = errors.New("client disconnected")
	errSendBufferFull     = errors.New("client send buffer full")
)

func (c *Client) WriteJSON(message interface{}) (err error) {
	payload, marshalErr := json.Marshal(message)
	if marshalErr != nil {
		return marshalErr
	}

	defer func() {
		if r := recover(); r != nil {
			err = errClientDisconnected
		}
	}()

	select {
	case <-c.stopCh:
		return errClientDisconnected
	default:
	}

	select {
	case c.Send <- payload:
		select {
		case <-c.stopCh:
			return errClientDisconnected
		default:
			return nil
		}
	case <-c.stopCh:
		return errClientDisconnected
	default:
		// Buffer full; let callers decide how to handle backpressure.
		return errSendBufferFull
	}
}

// writePump pumps messages from the hub to the websocket connection.
// An application runs writePump in a per-connection goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(50 * time.Second) // Ping ticker
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		
		case <-c.stopCh:
			return
		}
	}
}

func NewLobby() *Lobby {
	return &Lobby{
		clients:    make(map[*Client]bool),
		waiting:    make([]*Client, 0),
		games:      make(map[*GameState]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (l *Lobby) Run() {
	for {
		select {
		case client := <-l.register:
			l.onClientRegistered(client)

		case client := <-l.unregister:
			l.onClientUnregistered(client)
		}
	}
}

// JoinPairQueue handles a client's request to join the matchmaking queue.
// It checks if the client is already waiting, adds them to the queue if not,
// and starts a game if a pair is found. Otherwise, it notifies the client that they are waiting.
func (l *Lobby) JoinPairQueue(client *Client) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if already waiting
	if l.isWaiting(client) {
		return
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
		// Use a goroutine to avoid blocking the lock - though WriteJSON is now non-blocking(ish)
		// With channel, it IS safe to call directly if channel has buffer.
		if err := client.WriteJSON(msg); err != nil {
			log.Printf("Error sending wait message: %v", err)
		}
	}
}

func (l *Lobby) StartPairGame(p1, p2 *Client) {
	log.Printf("Starting pair game for %s and %s", p1.Nickname, p2.Nickname)

	// Create new game with two players, default ghosts 4
	game := NewGame([]string{p1.Nickname, p2.Nickname}, 4)
	l.games[game] = true
	p1.SetGame(game)
	p2.SetGame(game)

	// Start game loop
	go func() {
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()

		// Notify start
		startMsg := map[string]interface{}{
			"type": "game_start",
			"mode": "pair",
			"p1":   p1.Nickname,
			"p2":   p2.Nickname,
		}
		l.broadcastToPair(p1, p2, startMsg)

		for range ticker.C {
			if !l.handleGameTick(game, p1, p2) {
				return
			}
		}
	}()
}

func (l *Lobby) broadcastToPair(p1, p2 *Client, message interface{}) error {
	err1 := p1.WriteJSON(message)
	err2 := p2.WriteJSON(message)
	if err1 != nil {
		return err1
	}
	return err2
}

func (l *Lobby) handleGameTick(game *GameState, p1, p2 *Client) bool {
	game.Update()

	game.mu.RLock()
	// Check if game is over (both dead)
	if game.GameOver {
		l.handleGameOver(game, p1, p2)
		return false
	}

	err := l.broadcastToPair(p1, p2, game)
	game.mu.RUnlock()

	if err != nil {
		log.Println("Error writing to client in pair game, ending game")
		game.mu.Lock()
		game.GameOver = true // Stop updates
		game.mu.Unlock()
		l.cleanupPairGame(game, p1, p2)
		return false
	}
	return true
}

func (l *Lobby) handleGameOver(game *GameState, p1, p2 *Client) {
	game.mu.RUnlock()
	// Send final state
	game.mu.RLock()
	state := game
	game.mu.RUnlock()

	l.broadcastToPair(p1, p2, state)

	// Save Score
	db.SavePairScore(p1.Nickname, p2.Nickname, state.Score)

	l.cleanupPairGame(game, p1, p2)
}

func (l *Lobby) cleanupPairGame(game *GameState, p1, p2 *Client) {
	p1.SetGame(nil)
	p2.SetGame(nil)

	l.mu.Lock()
	delete(l.games, game)
	l.mu.Unlock()
}

func (l *Lobby) BroadcastPlayerCount() {
	// This is just a helper to let clients know how many people are online
	// to show/hide the "Pair Mode" button ideally
	l.mu.Lock()
	defer l.mu.Unlock()

	count := len(l.clients)
	
	msg := map[string]interface{}{
		"type":         "lobby_stats",
		"online_count": count,
	}

	for client := range l.clients {
		client.WriteJSON(msg)
	}
}

func (l *Lobby) isWaiting(client *Client) bool {
	for _, c := range l.waiting {
		if c == client {
			return true
		}
	}
	return false
}

func (l *Lobby) onClientRegistered(client *Client) {
	l.mu.Lock()
	l.clients[client] = true
	l.mu.Unlock()
	log.Printf("Client registered: %s", client.Nickname)
	l.BroadcastPlayerCount()
}

func (l *Lobby) onClientUnregistered(client *Client) {
	l.mu.Lock()
	if _, ok := l.clients[client]; ok {
		delete(l.clients, client)
		close(client.stopCh) // Signal pump to stop
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
}
