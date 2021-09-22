package chat

import (
	"log"
	"os"

	"github.com/bzeeno/RealTimeChat/database"
	"github.com/bzeeno/RealTimeChat/models"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var SECRET_KEY = os.Getenv("SECRET_KEY")

var current_pools []*Pool

func Connect(c *websocket.Conn) {
	room_id, _ := primitive.ObjectIDFromHex(c.Params("id")) // get room id (which will also be pool id)
	home_id, _ := primitive.ObjectIDFromHex("0")

	// Get user who sent message
	cookie := c.Cookies("jwt")
	var user models.User

	userCollection := database.DB.Collection("users")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) { // Get token
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return
	}

	claims := token.Claims.(*jwt.StandardClaims)
	objID, err := primitive.ObjectIDFromHex(claims.Issuer) // convert issuer in claims to mongo objectID
	if err != nil {
		return
	}

	if err := userCollection.FindOne(database.Context, bson.M{"_id": objID}).Decode(&user); err != nil { // Get user with specified id
		return
	}

	// loop through current_pools, if pool already exists: join. Else: create pool
	var found_pool = false
	var conn_pool *Pool
	for i, pool := range current_pools {
		if pool.ClientCount == 0 { // if there is an empty pool:
			pool = nil                                                        // assign pool object to nil
			current_pools = append(current_pools[:i], current_pools[i+1:]...) // remove it from the list
			continue
		}
		if pool.ID == room_id { // if pool already exists:
			conn_pool = pool // set conn_pool to pool
			found_pool = true
			break
		}
	}
	if !found_pool {
		conn_pool = NewPool(room_id) // create new pool
		current_pools = append(current_pools, conn_pool)
		go conn_pool.Listen() // go routine: Listen()
	}

	client := &Client{ID: user.ID, User: user.UserName, Conn: c, Pool: conn_pool} // create new client
	conn_pool.Register <- client                                                  // register client w/pool
	if room_id == home_id {
		client.ReadHome() // read
	} else {
		client.ReadMessage() // read
	}
}

func Reader(c *websocket.Conn) {
	log.Println("room id: ", c.Params("id")) // 123
	room_id, _ := primitive.ObjectIDFromHex(c.Params("id"))
	roomCollection := database.DB.Collection("rooms")
	var room models.Room

	if err := roomCollection.FindOne(database.Context, bson.M{"_id": room_id}).Decode(&room); err != nil { // Get room with specified id
		log.Println("Couldn't get room")
		return
	}

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Error in read: ", err)
			return
		}
		log.Println("received msg: ", string(msg))

		// Get user who sent message
		cookie := c.Cookies("jwt")

		userCollection := database.DB.Collection("users")
		var user models.User

		token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) { // Get token
			return []byte(SECRET_KEY), nil
		})
		if err != nil {
			return
		}

		claims := token.Claims.(*jwt.StandardClaims)
		objID, err := primitive.ObjectIDFromHex(claims.Issuer) // convert issuer in claims to mongo objectID
		if err != nil {
			return
		}

		if err := userCollection.FindOne(database.Context, bson.M{"_id": objID}).Decode(&user); err != nil { // Get user with specified id
			return
		}

		// set username to currently logged in user
		// set text to the received message
		return_message := models.Message{User: user.UserName, Text: string(msg)}

		// Add new message to database
		new_messages := append(room.Messages, return_message)
		update_field := bson.D{primitive.E{Key: "messages", Value: new_messages}}
		_, err = roomCollection.UpdateOne(database.Context, bson.M{"_id": room_id}, bson.D{
			{"$set", update_field},
		})
		if err != nil {
			log.Fatal(err)
		}

		if err := c.WriteJSON(return_message); err != nil { // write return message
			log.Println("Error in write: ", err)
			return
		}
	}
}
