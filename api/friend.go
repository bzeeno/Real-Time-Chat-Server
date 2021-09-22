package api

import (
	"fmt"
	"log"

	"github.com/bzeeno/RealTimeChat/database"
	"github.com/bzeeno/RealTimeChat/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Get all friends
func GetFriends(c *fiber.Ctx) error {
	user := GetUser(c) // Get user if authenticated

	friends := user.Friends // get friends list

	// return friends list and pending friends
	return c.JSON(fiber.Map{
		"friends": friends,
	})
}

// Get friend requests
func GetFriendReqs(c *fiber.Ctx) error {
	user := GetUser(c)
	requests := user.FriendReqs

	return c.JSON(fiber.Map{
		"requests": requests,
	})
}

func GetFriendInfo(c *fiber.Ctx) error {
	var data map[string]string
	var friend models.User

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	// Get friend
	friend = getFriend(data["friend_id"])

	if friend.UserName == "" { // if user not found:
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{ // send message
			"message": "Could not find user",
		})
	}

	return c.JSON(fiber.Map{
		"username":    friend.UserName,
		"profile_pic": friend.ProfilePic,
	})

}

func GetFriendChat(c *fiber.Ctx) error {
	var data map[string]string
	roomCollection := database.DB.Collection("rooms")
	user := GetUser(c)
	user_id := user.ID

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	// Get friend
	friend := getFriend(data["friend_id"])
	var room models.Room
	var friend_room_id primitive.ObjectID
	for _, room_id := range friend.Rooms {
		if err := roomCollection.FindOne(database.Context, bson.M{"_id": room_id}).Decode(&room); err != nil { // Get room with specified id
			return err
		}
		if room.FriendRoom == true {
			for _, person_id := range room.People {
				if person_id == user_id {
					friend_room_id = room_id
					break
				}
			}
		}
	}
	return c.JSON(fiber.Map{
		"room_id": friend_room_id,
	})
}

// Search for friends to add
func SearchUsers(c *fiber.Ctx) error {
	var data map[string]string
	userCollection := database.DB.Collection("users")
	//var user models.User
	var search_results []bson.M

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	// Make sure user who is searching is authenticated
	user := GetUser(c)
	if user.UserName == data["username"] {
		return c.JSON(fiber.Map{ // send message
			"message": "That's You!",
		})
	}

	cursor, err := userCollection.Find(database.Context, bson.M{"username": data["username"]})
	if err != nil { // Get friend who user is trying to add
		c.Status(fiber.StatusNotFound) // if user not found:
		return c.JSON(fiber.Map{       // send message
			"message": "User Not Found",
		})
	}

	if err := cursor.All(database.Context, &search_results); err != nil {
		c.Status(fiber.StatusNotFound) // if user not found:
		return c.JSON(fiber.Map{       // send message
			"message": "User Not Found",
		})
	}

	return c.JSON(search_results)
}

// Add friend (Takes in: ids for user and friend | Returns: message)
func AddFriend(c *fiber.Ctx) error {
	var data map[string]string
	var friend_is_pending bool
	userCollection := database.DB.Collection("users")
	roomCollection := database.DB.Collection("rooms")

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	user := GetUser(c)

	friend := getFriend(data["friend_id"])
	if friend.UserName == "" { // if user not found:
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{ // send message
			"message": "Could not find user",
		})
	}
	// Check if user has already sent a friend request
	for _, user_id := range friend.FriendReqs {
		if user_id == user.ID {
			return c.JSON(fiber.Map{
				"message": "You have already sent a friend request.",
			})
		}
	}
	// Check if friend is in user's list
	for _, friend_id := range user.Friends {
		if friend_id == friend.ID {
			return c.JSON(fiber.Map{
				"message": "You are already friends!",
			})
		}
	}

	for i, pending_friend := range user.FriendReqs { // for pending_friend in user's friend requests
		if pending_friend == friend.ID { // if friend is in pending requests
			// add friend to user's friend list
			user.Friends = append(user.Friends, friend.ID)
			update_field := bson.D{primitive.E{Key: "friends", Value: user.Friends}}
			_, err := userCollection.UpdateOne(database.Context, bson.M{"_id": user.ID}, bson.D{
				{"$set", update_field},
			})
			if err != nil {
				log.Fatal(err)
			}
			// add user to friend's friend list
			friend.Friends = append(friend.Friends, user.ID)
			update_field = bson.D{primitive.E{Key: "friends", Value: friend.Friends}}
			_, err = userCollection.UpdateOne(database.Context, bson.M{"_id": friend.ID}, bson.D{
				{"$set", update_field},
			})
			if err != nil {
				log.Fatal(err)
			}

			// remove friend from pending requests
			new_friend_reqs := append(user.FriendReqs[:i], user.FriendReqs[i+1:]...)
			update_field = bson.D{primitive.E{Key: "friend_reqs", Value: new_friend_reqs}}
			_, err = userCollection.UpdateOne(database.Context, bson.M{"_id": user.ID}, bson.D{
				{"$set", update_field},
			})
			if err != nil {
				log.Fatal(err)
			}

			/* Create chat room for just these 2 friends */
			new_room := models.Room{
				Name:       user.UserName + friend.UserName,
				People:     []primitive.ObjectID{user.ID, friend.ID},
				Messages:   []models.Message{},
				FriendRoom: true,
			}
			_, err = roomCollection.InsertOne(database.Context, new_room) // insert new room in database
			if err != nil {
				log.Fatal(err)
			}

			// insert room into both user's lists
			if err := roomCollection.FindOne(database.Context, bson.M{"name": user.UserName + friend.UserName}).Decode(&new_room); err != nil {
				return err
			}
			new_rooms := append(user.Rooms, new_room.ID)
			update_field = bson.D{primitive.E{Key: "rooms", Value: new_rooms}}
			_, err = userCollection.UpdateOne(database.Context, bson.M{"_id": user.ID}, bson.D{
				{"$set", update_field},
			})
			if err != nil {
				log.Fatal(err)
			}

			new_rooms = append(friend.Rooms, new_room.ID)
			update_field = bson.D{primitive.E{Key: "rooms", Value: new_rooms}}
			_, err = userCollection.UpdateOne(database.Context, bson.M{"_id": friend.ID}, bson.D{
				{"$set", update_field},
			})
			if err != nil {
				log.Fatal(err)
			}

			friend_is_pending = true
			return c.JSON(fiber.Map{
				"message": "Successfully added friend",
			})
		}
	}
	if !friend_is_pending { // if friend is not in pending request
		// add user to friend's pending list
		friend.FriendReqs = append(friend.FriendReqs, user.ID)
		update_field := bson.D{primitive.E{Key: "friend_reqs", Value: friend.FriendReqs}}

		_, err := userCollection.UpdateOne(database.Context, bson.M{"_id": friend.ID}, bson.D{
			{"$set", update_field},
		})
		if err != nil {
			log.Fatal(err)
		}
		return c.JSON(fiber.Map{
			"message": "Friend request sent",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Something went wrong",
	})

}

// Remove friend
func RemoveFriend(c *fiber.Ctx) error {
	var data map[string]string
	var user, friend models.User
	userCollection := database.DB.Collection("users")
	roomCollection := database.DB.Collection("rooms")

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	// Get user
	user = GetUser(c)
	if user.UserName == "" { // if user not found:
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{ // send message
			"message": "You Are Not Logged In",
		})
	}

	// Get friend
	friend = getFriend(data["friend_id"])
	if friend.UserName == "" { // if user not found:
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{ // send message
			"message": "You Are Not Logged In",
		})
	}

	// Remove friend from user's list
	for i, friend_id := range user.Friends {
		if friend_id == friend.ID {
			user.Friends = append(user.Friends[:i], user.Friends[i+1:]...)
			break
		}
	}
	update_field := bson.D{primitive.E{Key: "friends", Value: user.Friends}}
	_, err := userCollection.UpdateOne(database.Context, bson.M{"_id": user.ID}, bson.D{
		{"$set", update_field},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Remove user from friend's list
	for i, user_id := range friend.Friends {
		if user_id == user.ID {
			friend.Friends = append(friend.Friends[:i], friend.Friends[i+1:]...)
			break
		}
	}
	update_field = bson.D{primitive.E{Key: "friends", Value: friend.Friends}}
	_, err = userCollection.UpdateOne(database.Context, bson.M{"_id": friend.ID}, bson.D{
		{"$set", update_field},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Delete room from both users' lists
	// 2 possible room names
	chat_name1 := friend.UserName + user.UserName
	chat_name2 := user.UserName + friend.UserName
	var room models.Room
	var room_id primitive.ObjectID
	var i int

	// delete room from user's list
	for i, room_id = range user.Rooms {
		roomCollection.FindOne(database.Context, bson.M{"_id": room_id}).Decode(&room) // Get current room
		if (room.Name == chat_name1) || (room.Name == chat_name2) {
			user.Rooms = append(user.Rooms[:i], user.Rooms[i+1:]...)
			break
		}
	}
	update_field = bson.D{primitive.E{Key: "rooms", Value: user.Rooms}}
	_, err = userCollection.UpdateOne(database.Context, bson.M{"_id": user.ID}, bson.D{
		{"$set", update_field},
	})
	if err != nil {
		log.Fatal(err)
	}

	// delete room from friend's list
	for j, friend_room_id := range friend.Rooms {
		if friend_room_id == room_id {
			friend.Rooms = append(friend.Rooms[:j], friend.Rooms[j+1:]...)
			break
		}
	}
	update_field = bson.D{primitive.E{Key: "rooms", Value: friend.Rooms}}
	_, err = userCollection.UpdateOne(database.Context, bson.M{"_id": friend.ID}, bson.D{
		{"$set", update_field},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Delete chat room
	fmt.Println("chat names: ", chat_name1, " ", chat_name2)
	if _, err = roomCollection.DeleteOne(database.Context, bson.M{"name": chat_name1}); err != nil { // if name == chat_name1, delete
		return c.JSON(fiber.Map{
			"message": err,
		})
	}
	if _, err = roomCollection.DeleteOne(database.Context, bson.M{"name": chat_name2}); err != nil { // if name == chat_name2, delete
		return c.JSON(fiber.Map{
			"message": err,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Friend has been removed",
	})
}

// Check if current user and requested user are friends
func CheckIfFriends(c *fiber.Ctx) error {
	var data map[string]string
	var user, friend models.User

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	// Get user
	user = GetUser(c)

	// Get friend
	friend = getFriend(data["friend_id"])

	if friend.UserName == "" { // if user not found:
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{ // send message
			"message": "Could not find user",
		})
	}

	// Check if friend is in user's list
	friend_in_user := false
	for _, friend_id := range user.Friends {
		if friend_id == friend.ID {
			friend_in_user = true
		}
	}

	// if friend not in user's list, they are not friends
	if !friend_in_user {
		return c.JSON(fiber.Map{
			"message": "false",
		})
	}

	// check if user is in friend's list
	for _, user_id := range friend.Friends {
		if user_id == user.ID {
			return c.JSON(fiber.Map{
				"message": "true",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "false",
	})
}

// Get messages w/friend

// Get user helper function
func getFriend(user_id string) models.User {
	var user models.User
	userCollection := database.DB.Collection("users")

	objID, _ := primitive.ObjectIDFromHex(user_id)

	userCollection.FindOne(database.Context, bson.M{"_id": objID}).Decode(&user) // Get user who is adding friend with specified id
	return (user)
}
