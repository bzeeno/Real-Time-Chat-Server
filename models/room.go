package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var mongoURI = "mongodb://localhost:27017"

type Room struct {
	ID         primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string               `json:"name" bson:"name"`
	People     []primitive.ObjectID `json:"people" bson:"people"`
	Messages   []Message            `json:"messages" bson:"messages"`
	FriendRoom bool                 `json:"friend_room" bson:"friend_room"`
	RoomPic    string               `json:"room_pic" bson:"room_pic"`
}

type Message struct {
	ID   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	User string             `json:"user" bson:"user"`
	Text string             `json:"text" bson:"text"`
}
