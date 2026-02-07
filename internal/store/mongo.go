package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

func Connect(uri, databaseName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return &MongoDB{client: client, database: client.Database(databaseName)}, err
	}

	return &MongoDB{client: client, database: client.Database(databaseName)}, nil
}

type UpsertResult struct {
	Updated bool
}

func (m *MongoDB) UpsertCV(ctx context.Context, doc map[string]interface{}, lattesID, originalFilename string, fileSize int64) (*UpsertResult, error) {
	collection := m.database.Collection("curriculos")

	doc["_id"] = lattesID
	doc["_metadata"] = map[string]interface{}{
		"uploadedAt":       time.Now().UTC(),
		"originalFilename": originalFilename,
		"fileSize":         fileSize,
	}

	filter := map[string]interface{}{"_id": lattesID}
	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, doc, opts)
	if err != nil {
		return nil, err
	}

	return &UpsertResult{Updated: result.MatchedCount > 0}, nil
}

func (m *MongoDB) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

func (m *MongoDB) Disconnect(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
