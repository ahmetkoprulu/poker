package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ahmetkoprulu/rtrp/game/common/utils"
	"github.com/ahmetkoprulu/rtrp/game/internal/api"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, this should be more restrictive
	},
}

type Server struct {
	clients        map[string]*Client
	roomManager    *RoomManager
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	mu             sync.RWMutex
	handler        *MessageHandler
	IdlePlayerTime time.Duration
	ApiService     *api.ApiService
}

func NewServer() *Server {
	apiService := api.NewApiService()
	roomManager := NewRoomManager()
	roomManager.CreateRoom("room_1", "Default Room", 100, 5, 10, GameTypeHoldem)

	server := &Server{
		clients:        make(map[string]*Client),
		roomManager:    roomManager,
		broadcast:      make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		IdlePlayerTime: 600 * time.Second,
		ApiService:     apiService,
	}

	server.handler = NewMessageHandler(server, roomManager)

	return server
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.User.Player.ID] = client
			s.mu.Unlock()

		case client := <-s.unregister:
			if _, ok := s.clients[client.User.Player.ID]; ok {
				s.mu.Lock()
				if client.CurrentRoom != nil {
					client.CurrentRoom.RemovePlayer(client.User.Player.ID)
				}
				delete(s.clients, client.User.Player.ID)
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
					delete(s.clients, client.User.Player.ID)
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

	_, _, err := validateTokenAsString(token)
	if err != nil {
		log.Printf("Failed to validate token: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := s.ApiService.AuthService.GetUser(token)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		User:           user,
		authToken:      token,
		Conn:           conn,
		IpAddress:      r.RemoteAddr,
		ConnectionTime: time.Now(),
		ConnectCount:   0,
		mu:             sync.Mutex{},
		Server:         s,
		send:           make(chan []byte, 256),
	}
	client.Touch()
	s.register <- client

	go client.writePump()
	go client.readPump()
}

func (s *Server) HandleRoomList(w http.ResponseWriter, r *http.Request) {
	gameType := r.URL.Query().Get("game_type")
	var gameTypeInt int
	if gameType == "" {
		gameTypeInt = int(GameTypeHoldem)
	}

	if gameTypeInt == 0 {
		http.Error(w, "Invalid game type", http.StatusBadRequest)
		return
	}

	rooms := s.roomManager.GetRoomsByGameType(GameType(gameTypeInt))
	json.NewEncoder(w).Encode(rooms)
}

func (s *Server) GetRoom(roomID string) *Room {
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

	room.AddPlayer(client)

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
		if client, ok := s.clients[player.User.Player.ID]; ok {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(s.clients, client.User.Player.ID)
			}
		}
	}
}

func (s *Server) BroadcastToAll(message []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, client := range s.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(s.clients, client.User.Player.ID)
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

	for _, gamePlayer := range game.Players {
		if client, ok := s.clients[gamePlayer.Client.User.Player.ID]; ok {
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
