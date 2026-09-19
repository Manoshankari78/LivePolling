package repositories

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"live-polling-app/backend/models"
)

type PollRepository interface {
	Create(ctx context.Context, poll *models.Poll) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error)
	ListByCreator(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error)
	Update(ctx context.Context, id, creatorID primitive.ObjectID, question string, pollOptions []models.PollOption, expiresAt *time.Time) (*models.Poll, error)
	Delete(ctx context.Context, id, creatorID primitive.ObjectID) error
	Close(ctx context.Context, id, creatorID primitive.ObjectID) error
	IncrementOption(ctx context.Context, id primitive.ObjectID, optionID string) error
}

type MongoPollRepository struct{ collection *mongo.Collection }

func NewPollRepository(db *mongo.Database) *MongoPollRepository {
	return &MongoPollRepository{db.Collection("polls")}
}
func (r *MongoPollRepository) Create(ctx context.Context, poll *models.Poll) error {
	_, err := r.collection.InsertOne(ctx, poll)
	return err
}
func (r *MongoPollRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	var poll models.Poll
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, mongo.ErrNoDocuments
	}
	return &poll, err
}
func (r *MongoPollRepository) ListByCreator(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	cur, err := r.collection.Find(ctx, bson.M{"creatorId": creatorID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var polls []models.Poll
	if err := cur.All(ctx, &polls); err != nil {
		return nil, err
	}
	return polls, nil
}
func (r *MongoPollRepository) Update(ctx context.Context, id, creatorID primitive.ObjectID, question string, pollOptions []models.PollOption, expiresAt *time.Time) (*models.Poll, error) {
	update := bson.M{"$set": bson.M{"question": question, "options": pollOptions, "expiresAt": expiresAt, "updatedAt": time.Now()}}
	var poll models.Poll
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id, "creatorId": creatorID, "status": "active"}, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&poll)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, mongo.ErrNoDocuments
	}
	return &poll, err
}
func (r *MongoPollRepository) Delete(ctx context.Context, id, creatorID primitive.ObjectID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id, "creatorId": creatorID})
	if err == nil && res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return err
}
func (r *MongoPollRepository) Close(ctx context.Context, id, creatorID primitive.ObjectID) error {
	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "creatorId": creatorID, "status": "active"}, bson.M{"$set": bson.M{"status": "closed", "updatedAt": time.Now()}})
	if err == nil && res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return err
}
func (r *MongoPollRepository) IncrementOption(ctx context.Context, id primitive.ObjectID, optionID string) error {
	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "options.id": optionID, "status": "active"}, bson.M{"$inc": bson.M{"options.$.votes": 1}, "$set": bson.M{"updatedAt": time.Now()}})
	if err == nil && res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return err
}
