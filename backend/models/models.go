package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type PollOption struct {
	ID    string `bson:"id" json:"id"`
	Text  string `bson:"text" json:"text"`
	Votes int64  `bson:"votes" json:"votes"`
}

type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatorID primitive.ObjectID `bson:"creatorId" json:"creatorId"`
	Question  string             `bson:"question" json:"question"`
	Options   []PollOption       `bson:"options" json:"options"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	ExpiresAt *time.Time         `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
}

type Vote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID    primitive.ObjectID `bson:"pollId" json:"pollId"`
	OptionID  string             `bson:"optionId" json:"optionId"`
	VoterID   string             `bson:"voterId" json:"-"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

type Result struct {
	OptionID string `json:"optionId"`
	Votes    int64  `json:"votes"`
}

type PollResults struct {
	PollID     string   `json:"pollId"`
	Results    []Result `json:"results"`
	TotalVotes int64    `json:"totalVotes"`
}
