package models

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DelCmdReq represents the expected fields needed for the delcmd request to be completed.
type DelCmdReq struct {
	ID  string `json:"id" bson:"_id"`
	Cmd string `json:"cmd" bson:"cmd"`
}

// DelCmdRes represents the fields returned by the delcmd endpoint.
type DelCmdRes struct {
	NumDeleted int    `json:"numDeleted"`
	Cmd        string `json:"cmd"`
}

// RemoveCmdFromUser takes a given username along with the cmd and removes the cmd from their bookmarks.
func RemoveCmdFromUser(ctx context.Context, collection *mongo.Collection, userID, cmd string) (*mongo.UpdateResult, error) {
	opts := options.Update().SetUpsert(true)
	filter, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	update := bson.D{primitive.E{Key: "$unset", Value: bson.D{primitive.E{Key: fmt.Sprintf("bookmarks.%s", cmd), Value: ""}}}}
	result, err := collection.UpdateByID(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}