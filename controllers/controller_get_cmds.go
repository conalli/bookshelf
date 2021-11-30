package controllers

import (
	"github.com/conalli/bookshelf-backend/db"
	"github.com/conalli/bookshelf-backend/models"
	"github.com/conalli/bookshelf-backend/models/apiErrors"
)

// GetAllCmds uses req info to get all users current cmds from the db.
func GetAllCmds(userName string) (map[string]string, apiErrors.ApiErr) {
	ctx, cancelFunc := db.MongoContext()
	client := db.MongoClient(ctx)
	defer cancelFunc()
	defer client.Disconnect(ctx)

	collection := db.MongoCollection(client, "users")
	user, err := models.GetUserByKey(ctx, &collection, "name", userName)
	if err != nil {
		return nil, apiErrors.ParseGetUserError(userName, err)
	}
	return user.Bookmarks, nil
}
