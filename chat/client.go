package chat

import (
	"log"

	"github.com/bzeeno/RealTimeChat/database"
	"github.com/bzeeno/RealTimeChat/models"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Client struct {
	ID   primitive.ObjectID
	User string
	Conn *websocket.Conn
	Pool *Pool
}

type Request struct {
	SenderID  string `json:"sender_id" bson:"sender_id"`
	FriendID  string `json:"friend_id" bson:"friend_id"`
	RoomID    string `json:"room_id" bson:"room_id"`
	InPending string `json:"in_pending" bson:"in_pending"`
	Request   string `json:"req" bson:"req"`
}

func (this *Client) ReadHome() {
	defer func() {
		this.Pool.Unregister <- this
		this.Conn.Close()
	}()

	for {
		// Read request from client 1
		var req Request
		//_, msg, err := this.Conn.ReadMessage()
		err := this.Conn.ReadJSON(&req)
		//err = json.Unmarshal(msg, &req)
		//err := this.Conn.ReadJSON(req)
		if err != nil {
			log.Println(err)
		}
		log.Println("received request from home: ", req)
		if err != nil {
			log.Println(err)
			return
		}

		usr_id_str := this.ID.Hex()

		new_req := Request{SenderID: req.SenderID, FriendID: req.FriendID, InPending: "false", Request: req.Request} // request to send to clients, set friend_id to client who sent request

		// send request to client 1
		if err := this.Conn.WriteJSON(new_req); err != nil { // send request to friend
			log.Println(err)
		}

		new_req = Request{SenderID: req.SenderID, FriendID: usr_id_str, InPending: req.InPending, Request: req.Request, RoomID: req.RoomID} // request to send to clients, set friend_id to client who sent request
		log.Println("new_req: ", new_req)
		// send request to client 2
		friend_objID, _ := primitive.ObjectIDFromHex(req.FriendID)
		for client, _ := range this.Pool.Clients { // loop through clients connected to homepage
			if client.ID == friend_objID { // if friend is client
				if err := client.Conn.WriteJSON(new_req); err != nil { // send request to friend
					log.Println(err)
					return
				}
			}
		}
	}
}

func (this *Client) ReadMessage() {
	defer func() {
		this.Pool.Unregister <- this
		this.Conn.Close()
	}()

	for {
		_, msg, err := this.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		new_message := models.Message{User: this.User, Text: string(msg)} // message to broadcast
		this.Pool.Broadcast <- new_message                                // broadcast message
		// save message to database
		roomCollection := database.DB.Collection("rooms")
		room_id := this.Pool.ID
		var room models.Room

		if err := roomCollection.FindOne(database.Context, bson.M{"_id": room_id}).Decode(&room); err != nil { // Get room with specified id
			log.Println("Couldn't get room")
			return
		}
		new_messages := append(room.Messages, new_message)                        // new messages list
		update_field := bson.D{primitive.E{Key: "messages", Value: new_messages}} // update messages in database
		_, err = roomCollection.UpdateOne(database.Context, bson.M{"_id": room_id}, bson.D{
			{"$set", update_field},
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Message received in client read method: ", new_message)
	}
}
