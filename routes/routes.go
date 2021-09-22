package routes

import (
	"github.com/bzeeno/RealTimeChat/api"
	"github.com/bzeeno/RealTimeChat/chat"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func Setup(app *fiber.App) {
	// Authentication
	app.Post("/api/register", api.Register)
	app.Post("/api/login", api.Login)
	app.Get("/api/getuser", api.GetUserAuth)
	app.Post("/api/logout", api.Logout)
	// Friends
	app.Get("/api/get-friends", api.GetFriends)
	app.Get("/api/get-friend-reqs", api.GetFriendReqs)
	app.Post("/api/get-friend-chat", api.GetFriendChat)
	app.Post("/api/get-friend-info", api.GetFriendInfo)
	app.Post("/api/add-friend", api.AddFriend)
	app.Post("/api/remove-friend", api.RemoveFriend)
	app.Post("/api/search-friend", api.SearchUsers)
	app.Post("/api/check-friend", api.CheckIfFriends)
	// Rooms
	app.Get("/api/get-rooms", api.GetRooms)
	app.Post("/api/get-messages", api.GetMessages)
	app.Post("/api/get-room-info", api.GetRoomInfo)
	app.Post("/api/add-to-room", api.AddToRoom)
	app.Post("/api/create-room", api.CreateRoom)
	app.Post("/api/leave-room", api.LeaveRoom)
	// Websocket
	app.Get("/ws/:id", websocket.New(chat.Connect))
	app.Get("/ws/", websocket.New(chat.Connect))
}
