package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/conalli/bookshelf-backend/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Names for each MongoDB collection used.
const (
	CollectionUsers = "users"
	CollectionTeams = "teams"
)

// Mongo represents a Mongodb client and database.
type Mongo struct {
	client *mongo.Client
	db     *mongo.Database
}

func resolveEnv(envType string) string {
	switch envType {
	case "uri":
		if l := os.Getenv("LOCAL"); l == "production" || l == "atlas" {
			return os.Getenv("MONGO_URI")
		}
		return os.Getenv("LOCAL_MONGO_URI")
	case "db":
		if os.Getenv("LOCAL") == "production" {
			return os.Getenv("DB_NAME")
		}
		if os.Getenv("LOCAL") == "atlas" {
			return os.Getenv("TEST_DB_NAME")
		}
		return os.Getenv("LOCAL_DB_NAME")
	default:
		return ""
	}
}

func New() *Mongo {
	return &Mongo{}
}

func (m *Mongo) Initialize() {
	mongoURI := resolveEnv("uri")
	db := resolveEnv("db")
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	m.client = client
	m.db = m.client.Database(db)
}

// Collection uses the DB_NAME env var, and returns a collection based on the collectionName and client.
func (m *Mongo) Collection(collectionName string) *mongo.Collection {
	collection := m.db.Collection(collectionName)
	return collection
}

// SessionWithTransaction takes a context and transaction func and returns the result of the transaction.
func (m *Mongo) SessionWithTransaction(ctx context.Context, transactionFunc func(sessCtx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	sess, err := m.client.StartSession(opts)
	defer sess.EndSession(ctx)
	if err != nil {
		log.Println("could not start db session")
		return nil, errors.NewInternalServerError()
	}
	txnOpts := options.Transaction().SetReadPreference(readpref.Primary())
	res, err := sess.WithTransaction(ctx, transactionFunc, txnOpts)
	return res, err
}

// DataAlreadyExists attempts to find a user based on a given key-value pair, returning wether they
// already exist in the db or not.
func DataAlreadyExists(ctx context.Context, collection *mongo.Collection, key, value string) bool {
	var result bson.M
	err := collection.FindOne(ctx, bson.D{primitive.E{Key: key, Value: value}}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false
		}
		log.Printf("already exists %+v", err)
	}
	return true
}

// GetByID finds and returns user data based on a the users _id.
func GetByID(ctx context.Context, collection *mongo.Collection, userID string) (User, error) {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return User{}, err
	}
	var result User
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return result, mongo.ErrNoDocuments
		}
		return result, err
	}
	return result, nil
}

// GetUserByKey finds and returns user data based on a key-value pair.
func GetUserByKey(ctx context.Context, collection *mongo.Collection, reqKey, reqValue string) (User, error) {
	var result User
	err := collection.FindOne(ctx, bson.D{primitive.E{Key: reqKey, Value: reqValue}}).Decode(&result)
	if err != nil {
		log.Printf("get user: %+v", err)
		if err == mongo.ErrNoDocuments {
			return result, mongo.ErrNoDocuments
		}
		return result, err
	}
	return result, nil
}

type UpdateEmbedOptions struct {
	FilterKey, FilterValue, Embedded, Key, Value string
}

func UpdateEmbedByField(ctx context.Context, collection *mongo.Collection, data UpdateEmbedOptions) (*mongo.SingleResult, error) {
	options := options.FindOneAndUpdate().SetUpsert(true)
	var filter primitive.M
	if data.FilterKey == "_id" {
		userID, err := primitive.ObjectIDFromHex(data.FilterValue)
		if err != nil {
			return nil, err
		}
		filter = bson.M{data.FilterKey: userID}
	} else {
		filter = bson.M{data.FilterKey: data.FilterValue}
	}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: fmt.Sprintf("%s.%s", data.Embedded, data.Key), Value: data.Value}}}}
	return collection.FindOneAndUpdate(ctx, filter, update, options), nil
}

func DecodeUser(res *mongo.SingleResult) (User, error) {
	var user User
	err := res.Decode(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func DecodeTeam(res *mongo.SingleResult) (Team, error) {
	var team Team
	err := res.Decode(&team)
	if err != nil {
		return Team{}, err
	}
	return team, nil
}
