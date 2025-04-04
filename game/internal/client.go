package internal

import (
	"log"
	"sync"
	"time"

	"github.com/ahmetkoprulu/rtrp/game/models"
	"github.com/gorilla/websocket"
)

type Client struct {
	User           *models.User
	authToken      string
	IpAddress      string
	ConnectionTime time.Time
	ConnectCount   int
	DisconnectTime time.Time
	IdleTime       time.Time
	Conn           *websocket.Conn
	Server         *Server
	mu             sync.Mutex
	CurrentRoom    *Room
	CurrentGame    *Game
	IsDisconnected bool
	send           chan []byte
}

func (c *Client) readPump() {
	defer func() {
		// Remove player from game when disconnected
		room, err := c.Server.roomManager.GetRoomByPlayerID(c.User.Player.ID)
		if err != nil {
			return
		}

		if room != nil && room.Game != nil {
			if err := c.Server.handler.roomManager.LeaveRoom(room.ID, c.User.Player.ID); err != nil {
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

func (c *Client) Touch() {
	c.IdleTime = time.Now().Add(c.Server.IdlePlayerTime)
}
