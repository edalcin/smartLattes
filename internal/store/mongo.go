package store

import (
	"context"
	"fmt"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/v2/bson"
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

type CVSummary struct {
	LattesID string `json:"lattesId"`
	Name     string `json:"name"`
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func (m *MongoDB) SearchCVs(ctx context.Context, query string) ([]CVSummary, error) {
	collection := m.database.Collection("curriculos")

	var filter bson.M
	if isAllDigits(query) {
		filter = bson.M{"_id": bson.M{"$regex": "^" + query}}
	} else {
		filter = bson.M{"curriculo-vitae.dados-gerais.nome-completo": bson.M{"$regex": query, "$options": "i"}}
	}

	projection := bson.M{
		"_id": 1,
		"curriculo-vitae.dados-gerais.nome-completo": 1,
	}
	opts := options.Find().SetLimit(20).SetProjection(projection)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []CVSummary
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		summary := CVSummary{LattesID: fmt.Sprintf("%v", doc["_id"])}
		if cv, ok := doc["curriculo-vitae"].(bson.M); ok {
			if dg, ok := cv["dados-gerais"].(bson.M); ok {
				if name, ok := dg["nome-completo"].(string); ok {
					summary.Name = name
				}
			}
		}
		results = append(results, summary)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (m *MongoDB) GetCV(ctx context.Context, lattesID string) (map[string]interface{}, error) {
	collection := m.database.Collection("curriculos")

	var doc map[string]interface{}
	err := collection.FindOne(ctx, bson.M{"_id": lattesID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CV não encontrado")
		}
		return nil, err
	}

	return doc, nil
}

func (m *MongoDB) UpsertSummary(ctx context.Context, lattesID, summary, provider, model string) error {
	collection := m.database.Collection("resumos")

	doc := bson.M{
		"_id":    lattesID,
		"resumo": summary,
		"_metadata": bson.M{
			"generatedAt": time.Now().UTC(),
			"provider":    provider,
			"model":       model,
		},
	}

	filter := bson.M{"_id": lattesID}
	opts := options.Replace().SetUpsert(true)

	_, err := collection.ReplaceOne(ctx, filter, doc, opts)
	return err
}

type SummaryMetadata struct {
	GeneratedAt time.Time `bson:"generatedAt"`
	Provider    string    `bson:"provider"`
	Model       string    `bson:"model"`
}

type SummaryDoc struct {
	ID       string          `bson:"_id"`
	Resumo   string          `bson:"resumo"`
	Metadata SummaryMetadata `bson:"_metadata"`
}

func (m *MongoDB) GetSummary(ctx context.Context, lattesID string) (*SummaryDoc, error) {
	collection := m.database.Collection("resumos")

	var doc SummaryDoc
	err := collection.FindOne(ctx, bson.M{"_id": lattesID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("resumo não encontrado")
		}
		return nil, err
	}

	return &doc, nil
}
