package chat

import (
	"log"

	"github.com/bzeeno/RealTimeChat/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Pool struct {
	ID          primitive.ObjectID
	ClientCount int
	Clients     map[*Client]bool
	Register    chan *Client
	Unregister  chan *Client
	Broadcast   chan models.Message
}

// Create new Pool with default vals
func NewPool(pool_id primitive.ObjectID) *Pool {
	return &Pool{
		ID:          pool_id,
		ClientCount: 0,
		Clients:     make(map[*Client]bool),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Broadcast:   make(chan models.Message),
	}
}

// Listen for all events related to Pool
func (this *Pool) Listen() {
	for {
		select {
		case client := <-this.Register: // if client trying to register to pool
			this.Clients[client] = true // add client to pool
			this.ClientCount++
			break
		case client := <-this.Unregister: // if client is trying to unregister from pool
			delete(this.Clients, client) // remove client from pool
			this.ClientCount--
			break
		case message := <-this.Broadcast: // if client is trying to broadcast message to this pool
			for client, _ := range this.Clients { // loop through clients in pool
				if err := client.Conn.WriteJSON(message); err != nil { // write message to current client
					log.Println(err)
					return
				}
			}
			break
		}
	}
}
