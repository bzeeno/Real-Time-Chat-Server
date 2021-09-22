package main

import (
	"os"

	"github.com/bzeeno/RealTimeChat/database"
	"github.com/bzeeno/RealTimeChat/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

func main() {
	mongo_client := database.Connect()              // connect to database
	defer mongo_client.Disconnect(database.Context) // defer disconnect to database

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	routes.Setup(app) // setup routes

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	app.Listen(":" + port)
}
