package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	UserName   string               `json:"username" bson:"username"`
	Email      string               `json:"email" bson:"email"`
	Password   []byte               `json:"password" bson:"password"`
	Friends    []primitive.ObjectID `json:"friends" bson:"friends"`
	FriendReqs []primitive.ObjectID `json:"friend_reqs" bson:"friend_reqs"`
	Rooms      []primitive.ObjectID `json:"rooms" bson:"rooms"`
	ProfilePic string               `json:"profile_pic" bson:"profile_pic"`
}
