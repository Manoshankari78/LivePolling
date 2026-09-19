package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo(uri, databaseName string) (*mongo.Client, *mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, err
	}
	db := client.Database(databaseName)
	if err := ensureIndexes(ctx, db); err != nil {
		return nil, nil, err
	}
	return client, db, nil
}

func ensureIndexes(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	_, err = db.Collection("polls").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "creatorId", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}},
	})
	if err != nil {
		return err
	}
	_, err = db.Collection("votes").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "pollId", Value: 1}}},
		{Keys: bson.D{{Key: "pollId", Value: 1}, {Key: "voterId", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	return err
}
