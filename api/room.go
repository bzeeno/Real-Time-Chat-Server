package api

import (
	"log"

	"github.com/bzeeno/RealTimeChat/database"
	"github.com/bzeeno/RealTimeChat/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Get all rooms
func GetRooms(c *fiber.Ctx) error {
	user := GetUser(c) // Get user if authenticated
	roomCollection := database.DB.Collection("rooms")

	rooms := user.Rooms // get friends list

	var rooms_list []primitive.ObjectID
	var room models.Room
	for _, room_id := range rooms {
		if err := roomCollection.FindOne(database.Context, bson.M{"_id": room_id}).Decode(&room); err != nil { // Get room with specified id
			return err
		}
		if room.FriendRoom == true {
			continue
		} else {
			rooms_list = append(rooms_list, room.ID)
		}
	}

	// return friends list and pending friends
	return c.JSON(fiber.Map{
		"rooms": rooms_list,
	})
}

func GetRoomInfo(c *fiber.Ctx) error {
	var data map[string]string
	var room models.Room
	roomCollection := database.DB.Collection("rooms")

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	// Get room
	objID, _ := primitive.ObjectIDFromHex(data["room_id"])
	if err := roomCollection.FindOne(database.Context, bson.M{"_id": objID}).Decode(&room); err != nil { // Get room with specified id
		return err
	}

	return c.JSON(fiber.Map{
		"room_name": room.Name,
		"room_pic":  room.RoomPic,
		"users":     room.People,
	})

}

func AddToRoom(c *fiber.Ctx) error {
	var data map[string]string
	var room models.Room
	var friend models.User
	userCollection := database.DB.Collection("users")
	roomCollection := database.DB.Collection("rooms")

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	friend_id, _ := primitive.ObjectIDFromHex(data["friend_id"])
	room_id, _ := primitive.ObjectIDFromHex(data["room_id"])

	// Get room
	if err := roomCollection.FindOne(database.Context, bson.M{"_id": room_id}).Decode(&room); err != nil { // Get room with specified id
		return err
	}
	// Get friend
	if err := userCollection.FindOne(database.Context, bson.M{"_id": friend_id}).Decode(&friend); err != nil { // Get room with specified id
		return err
	}
	for _, friend_room := range friend.Rooms {
		if friend_room == room.ID {
			return c.JSON(fiber.Map{
				"message": "Friend is already in room",
			})
		}
	}
	// add room to friend's list
	new_rooms := append(friend.Rooms, room_id)
	update_field := bson.D{primitive.E{Key: "rooms", Value: new_rooms}}
	_, err := userCollection.UpdateOne(database.Context, bson.M{"_id": friend_id}, bson.D{
		{"$set", update_field},
	})
	if err != nil {
		log.Fatal(err)
	}

	// add friend to room
	new_people := append(room.People, friend_id)
	update_field = bson.D{primitive.E{Key: "people", Value: new_people}}
	_, err = roomCollection.UpdateOne(database.Context, bson.M{"_id": room_id}, bson.D{
		{"$set", update_field},
	})
	if err != nil {
		log.Fatal(err)
	}
	return c.JSON(fiber.Map{
		"message": "Successfully added friend to room",
	})
}

// Create Room
func CreateRoom(c *fiber.Ctx) error {
	var data map[string]string
	userCollection := database.DB.Collection("users")
	roomCollection := database.DB.Collection("rooms")

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	user := GetUser(c)

	roomName := data["name"]

	new_room := models.Room{
		Name:       roomName,
		People:     []primitive.ObjectID{user.ID},
		Messages:   []models.Message{},
		FriendRoom: false,
		RoomPic:    "default_room.jpeg",
	}

	// Create new room w/user in it
	_, err := roomCollection.InsertOne(database.Context, new_room) // insert new room in database
	if err != nil {
		log.Fatal(err)
	}

	// Add room to user's list
	var room models.Room
	if err := roomCollection.FindOne(database.Context, bson.M{"name": roomName}).Decode(&room); err != nil { // Get room with specified id
		return err
	}
	user.Rooms = append(user.Rooms, room.ID)
	update_field := bson.D{primitive.E{Key: "rooms", Value: user.Rooms}}
	_, err = userCollection.UpdateOne(database.Context, bson.M{"_id": user.ID}, bson.D{
		{"$set", update_field},
	})
	if err != nil {
		log.Fatal(err)
	}

	return c.JSON(fiber.Map{
		"message": "Successfully created room!",
	})

}

// Leave Room
func LeaveRoom(c *fiber.Ctx) error {
	var data map[string]string
	user := GetUser(c)

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	room_id, _ := data["room_id"]
	room_objID, _ := primitive.ObjectIDFromHex(room_id)

	userCollection := database.DB.Collection("users")
	roomCollection := database.DB.Collection("rooms")

	// remove room from user's list
	for i, curr_id := range user.Rooms {
		if curr_id == room_objID {
			user.Rooms = append(user.Rooms[:i], user.Rooms[i+1:]...)
			break
		}
	}
	update_field := bson.D{primitive.E{Key: "rooms", Value: user.Rooms}}
	_, err := userCollection.UpdateOne(database.Context, bson.M{"_id": user.ID}, bson.D{
		{"$set", update_field},
	})
	if err != nil {
		log.Fatal(err)
	}

	// remove user from room's list
	var room models.Room
	if err := roomCollection.FindOne(database.Context, bson.M{"_id": room_objID}).Decode(&room); err != nil { // Get room with specified id
		return err
	}
	for i, curr_id := range room.People {
		if curr_id == user.ID {
			room.People = append(room.People[:i], room.People[i+1:]...)
			break
		}
	}
	update_field = bson.D{primitive.E{Key: "people", Value: room.People}}
	_, err = roomCollection.UpdateOne(database.Context, bson.M{"_id": room.ID}, bson.D{
		{"$set", update_field},
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(room.People) == 0 {
		_, err := roomCollection.DeleteOne(database.Context, bson.M{"_id": room.ID})
		if err != nil {
			log.Fatal(err)
		}
	}

	return c.JSON(fiber.Map{
		"message": "Successfully left room!",
	})

}

// Invite to room
// Get Room Messages
func GetMessages(c *fiber.Ctx) error {
	var data map[string]string
	var room models.Room
	var userInRoom = false
	user := GetUser(c)
	roomCollection := database.DB.Collection("rooms")

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	// Get room id
	room_id := data["room_id"]
	room_objID, _ := primitive.ObjectIDFromHex(room_id)

	// If user is unauthenticated: return message
	if user.UserName == "" {
		return c.JSON(fiber.Map{
			"message": "Not authenticated",
		})
	}

	// Get room
	if err := roomCollection.FindOne(database.Context, bson.M{"_id": room_objID}).Decode(&room); err != nil { // Get room with specified id
		return err
	}
	// If user is not in room: return message
	for _, person_id := range room.People {
		if person_id == user.ID {
			userInRoom = true
			break
		}
	}
	if !userInRoom {
		return c.JSON(fiber.Map{
			"message": "You are not in this room",
		})
	}

	// Otherwise: Return messages
	return c.JSON(fiber.Map{
		"messages": room.Messages,
	})
}

// Get people in room
