package repositories

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"live-polling-app/backend/models"
)

type VoteRepository interface {
	Create(ctx context.Context, vote *models.Vote) error
	Results(ctx context.Context, poll *models.Poll) models.PollResults
}

type MongoVoteRepository struct{ collection *mongo.Collection }

func NewVoteRepository(db *mongo.Database) *MongoVoteRepository {
	return &MongoVoteRepository{db.Collection("votes")}
}
func (r *MongoVoteRepository) Create(ctx context.Context, vote *models.Vote) error {
	_, err := r.collection.InsertOne(ctx, vote)
	return err
}
func (r *MongoVoteRepository) Results(ctx context.Context, poll *models.Poll) models.PollResults {
	results := make([]models.Result, 0, len(poll.Options))
	var total int64
	for _, option := range poll.Options {
		results = append(results, models.Result{OptionID: option.ID, Votes: option.Votes})
		total += option.Votes
	}
	return models.PollResults{PollID: poll.ID.Hex(), Results: results, TotalVotes: total}
}

func IsDuplicateKey(err error) bool {
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		for _, item := range writeErr.WriteErrors {
			if item.Code == 11000 {
				return true
			}
		}
	}
	return false
}

func NewVote(pollID primitive.ObjectID, optionID, voterID string) *models.Vote {
	return &models.Vote{PollID: pollID, OptionID: optionID, VoterID: voterID, CreatedAt: time.Now()}
}

func (r *MongoVoteRepository) CountByPoll(ctx context.Context, pollID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"pollId": pollID})
}
