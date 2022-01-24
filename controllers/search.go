package controllers

import (
	"context"
	"fmt"

	"github.com/conalli/bookshelf-backend/db"
	"github.com/conalli/bookshelf-backend/models"
	"github.com/conalli/bookshelf-backend/models/apiErrors"
)

// GetURL takes in an apiKey and cmd and returns either a correctly formatted url from the db,
// or a google search url for the cmd based on whether the cmd could be found or not.
func GetURL(reqCtx context.Context, apiKey, cmd string) (string, apiErrors.ApiErr) {
	ctx, cancelFunc := db.ReqContextWithTimeout(reqCtx)
	defer cancelFunc()

	cache := db.NewRedisClient()
	url, err := cache.GetSearchData(ctx, apiKey, cmd)
	if err != nil {
		client := db.NewMongoClient(ctx)
		defer client.DB.Disconnect(ctx)
		collection := client.MongoCollection("users")

		user, err := models.GetUserByKey(ctx, &collection, "apiKey", apiKey)
		defaultSearch := fmt.Sprintf("http://www.google.com/search?q=%s", cmd)
		if err != nil {
			return defaultSearch, apiErrors.ParseGetUserError(apiKey, err)
		}

		cache.SetCacheCmds(ctx, apiKey, user.Bookmarks)

		url, found := user.Bookmarks[cmd]
		if !found {
			return defaultSearch, apiErrors.NewBadRequestError("error: command: " + cmd + " not registered")
		}
		return models.FormatURL(url), nil
	}
	return models.FormatURL(url), nil
}