package websocket

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/ahmetkoprulu/rtrp/game/common/utils"
	"github.com/ahmetkoprulu/rtrp/game/internal/room"
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
	PlayerID string
	Conn     *websocket.Conn
	Server   *Server
	mu       sync.Mutex
	send     chan []byte
}

type Server struct {
	clients     map[string]*Client
	roomManager *room.RoomManager
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mu          sync.RWMutex
	handler     *MessageHandler
}

func NewServer() *Server {
	roomManager := room.NewRoomManager()
	roomManager.CreateRoom("room_1", "Default Room", 100, 5, 10, models.GameTypeHoldem)

	server := &Server{
		clients:     make(map[string]*Client),
		roomManager: roomManager,
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}

	server.handler = NewMessageHandler(server, roomManager)

	return server
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.PlayerID] = client
			s.mu.Unlock()

		case client := <-s.unregister:
			if _, ok := s.clients[client.PlayerID]; ok {
				s.mu.Lock()
				delete(s.clients, client.PlayerID)
				close(client.send)
				s.mu.Unlock()
			}

		case message := <-s.broadcast:
			s.mu.RLock()
			for _, client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client.PlayerID)
				}
			}
			s.mu.RUnlock()
		}
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusUnauthorized)
		return
	}

	userID, playerID, err := validateTokenAsString(token)
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

	// // Get the default room
	// for roomID := range s.rooms {
	// 	defaultRoomID = roomID
	// 	break
	// }

	client := &Client{
		ID:       userID,
		PlayerID: playerID,
		Conn:     conn,
		Server:   s,
		send:     make(chan []byte, 256),
	}

	s.register <- client

	go client.readPump()
	go client.writePump()
}

func (s *Server) GetRoom(roomID string) *models.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return nil
	}
	return room
}

func (s *Server) JoinRoom(roomID string, client *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	room.Players = append(room.Players, &models.Player{
		ID:       client.PlayerID,
		Username: "Guest",
		Picture:  "https://via.placeholder.com/150",
		Chips:    1000,
	})

	s.roomManager.RegisterRoom(room)
	return nil
}

func (s *Server) BroadcastToRoom(roomID string, message []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return
	}

	for _, player := range room.Players {
		if client, ok := s.clients[player.ID]; ok {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(s.clients, client.PlayerID)
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		// Remove player from game when disconnected
		room, err := c.Server.roomManager.GetRoomByPlayerID(c.PlayerID)
		if err != nil {
			return
		}

		if room != nil && room.Game != nil {
			if err := c.Server.handler.roomManager.LeaveRoom(room.ID, c.PlayerID); err != nil {
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

func (s *Server) BroadcastToGame(roomID string, message []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return
	}
	game := room.Game

	for _, player := range game.Players {
		// TODO: Add game ID to client struct and check if client is in the game
		if client, ok := s.clients[player.Player.ID]; ok {
			client.send <- message
		}
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

func validateTokenAsString(token string) (string, string, error) {
	claims, err := utils.ValidateJwTTokenWithClaims(token)
	if err != nil {
		return "", "", err
	}

	return claims.UserID, claims.PlayerID, nil
}
