package api

import (
	"log"
	"os"
	"time"

	"github.com/bzeeno/RealTimeChat/database"
	"github.com/bzeeno/RealTimeChat/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var SECRET_KEY = os.Getenv("SECRET_KEY")

func Register(c *fiber.Ctx) error {
	var data map[string]string
	userCollection := database.DB.Collection("users")

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	new_user := models.User{
		UserName:   data["username"],
		Email:      data["email"],
		Password:   password,
		Friends:    []primitive.ObjectID{},
		FriendReqs: []primitive.ObjectID{},
		Rooms:      []primitive.ObjectID{},
		ProfilePic: "default_pic.jpeg",
	}

	_, err := userCollection.InsertOne(database.Context, new_user) // insert new user in database

	if err != nil {
		log.Fatal(err)
	}

	return c.JSON(data)

}

func Login(c *fiber.Ctx) error {
	var data map[string]string
	userCollection := database.DB.Collection("users")
	var user models.User

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	if err := userCollection.FindOne(database.Context, bson.M{"email": data["email"]}).Decode(&user); err != nil { // Get user with specified email
		c.Status(fiber.StatusNotFound) // if user not found:
		return c.JSON(fiber.Map{       // send message
			"message": "User Not Found",
		})
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil { // compare user password to entered password
		c.Status(fiber.StatusBadRequest) // if incorrect password:
		return c.JSON(fiber.Map{         // send message
			"message": "Incorrect Password",
		})
	}

	// set jwt
	expire_time := time.Now().Add(time.Hour * 24) // token expires in a day
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.ID.Hex(),      // use user ID as issuer (convert to string hex)
		ExpiresAt: expire_time.Unix(), // set expiration date
	})

	token, err := claims.SignedString([]byte(SECRET_KEY))

	if err != nil {
		c.Status(fiber.StatusInternalServerError) // if couldn't get token
		return c.JSON(fiber.Map{                  // send message
			"message": "Could Not Login",
		})
	}

	// set fiber cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "jwt"
	cookie.Value = token
	cookie.Expires = expire_time
	cookie.HTTPOnly = true

	c.Cookie(cookie) // set cookie in fiber context

	return c.JSON(user)

}

func GetUserAuth(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt") // get cookie
	userCollection := database.DB.Collection("users")
	var user models.User

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		c.Status(fiber.StatusUnauthorized) // if couldn't get token
		return c.JSON(fiber.Map{           // send message
			"message": "Unauthenticated",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)
	objID, err := primitive.ObjectIDFromHex(claims.Issuer) // convert issuer in claims to mongo objectID
	if err != nil {
		log.Fatal(err)
	}

	if err := userCollection.FindOne(database.Context, bson.M{"_id": objID}).Decode(&user); err != nil { // Get user with specified id
		c.Status(fiber.StatusUnauthorized) // if couldn't get token
		return c.JSON(fiber.Map{           // send message
			"message": "Couldn't get cookie",
		})
	}
	return c.JSON(user)
}

func Logout(c *fiber.Ctx) error {
	// delete fiber cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "jwt"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour)
	cookie.HTTPOnly = true

	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"message": "",
	})
}

// Function for getting model of user who is currently logged in. If user is not logged in: return empty user model
func GetUser(c *fiber.Ctx) models.User {
	cookie := c.Cookies("jwt") // get cookie
	userCollection := database.DB.Collection("users")
	var user models.User

	// Get token
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return user
	}

	claims := token.Claims.(*jwt.StandardClaims)
	objID, err := primitive.ObjectIDFromHex(claims.Issuer) // convert issuer in claims to mongo objectID
	if err != nil {
		return user
	}

	if err := userCollection.FindOne(database.Context, bson.M{"_id": objID}).Decode(&user); err != nil { // Get user with specified id
		return user
	}

	return (user)
}
