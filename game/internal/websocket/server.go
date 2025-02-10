package websocket

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/ahmetkoprulu/rtrp/game/common/utils"
	"github.com/ahmetkoprulu/rtrp/game/models"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, this should be more restrictive
	},
}

type Client struct {
	ID       string
	RoomID   string
	PlayerID string
	Conn     *websocket.Conn
	Server   *Server
	mu       sync.Mutex
	send     chan []byte
}

type Server struct {
	clients    map[*Client]bool
	rooms      map[string]*models.Room
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	handler    *MessageHandler
}

func NewServer(handler *MessageHandler) *Server {
	defaultRoom := models.NewRoom("room_1", 5, 2) // Create default room 5 players max, 2 min bet

	// Register the initial game with the game manager
	handler.gameManager.RegisterGame(defaultRoom.Game)

	return &Server{
		clients:    make(map[*Client]bool),
		rooms:      map[string]*models.Room{defaultRoom.ID: defaultRoom},
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		handler:    handler,
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()

		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				s.mu.Lock()
				delete(s.clients, client)
				close(client.send)
				s.mu.Unlock()
			}

		case message := <-s.broadcast:
			s.mu.RLock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mu.RUnlock()
		}
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, playerID, err := validateToken(r)
	if err != nil {
		log.Printf("Failed to validate token: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Get the default room
	var defaultRoomID string
	for roomID := range s.rooms {
		defaultRoomID = roomID
		break
	}

	client := &Client{
		ID:       userID,
		RoomID:   defaultRoomID,
		PlayerID: playerID,
		Conn:     conn,
		Server:   s,
		send:     make(chan []byte, 256),
	}

	s.register <- client

	go client.readPump()
	go client.writePump()
}

// GetRoom returns a room by its ID
func (s *Server) GetRoom(roomID string) *models.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rooms[roomID]
}

// BroadcastToRoom broadcasts a message to all clients in a specific room
func (s *Server) BroadcastToRoom(roomID string, message []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for client := range s.clients {
		if client.RoomID == roomID {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(s.clients, client)
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		// Remove player from game when disconnected
		room := c.Server.GetRoom(c.RoomID)
		if room != nil && room.Game != nil {
			if err := c.Server.handler.gameManager.LeaveGame(room.Game.ID, c.PlayerID); err != nil {
				log.Printf("Error removing player from game: %v", err)
			}
			// Broadcast the updated room state to other players
			c.Server.handler.broadcastRoomState(room)
		}

		// Unregister client from server
		c.Server.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Handle the message using the message handler
		if err := c.Server.handler.HandleMessage(c, message); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// Server closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()

			if err != nil {
				log.Printf("error writing message: %v", err)
				return
			}
		}
	}
}

// BroadcastToGame broadcasts a message to all clients in a specific game
func (s *Server) BroadcastToGame(gameID string, message []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for client := range s.clients {
		// TODO: Add game ID to client struct and check if client is in the game
		client.send <- message
	}
}

func validateToken(r *http.Request) (string, string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", "", errors.New("authorization header is required")
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", "", errors.New("invalid authorization header format")
	}

	token := tokenParts[1]
	claims, err := utils.ValidateJwTTokenWithClaims(token)
	if err != nil {
		return "", "", err
	}

	return claims.UserID, claims.PlayerID, nil
}
