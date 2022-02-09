package mongodb

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/conalli/bookshelf-backend/pkg/db"
	"github.com/conalli/bookshelf-backend/pkg/errors"
	"github.com/conalli/bookshelf-backend/pkg/password"
	"github.com/conalli/bookshelf-backend/pkg/user"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User represents the db fields associated with each user.
type User struct {
	ID        string            `json:"id" bson:"_id,omitempty"`
	Name      string            `json:"name" bson:"name"`
	Password  string            `json:"password,omitempty" bson:"password"`
	APIKey    string            `json:"APIKey" bson:"APIKey"`
	Bookmarks map[string]string `json:"bookmarks,omitempty" bson:"bookmarks"`
	Teams     map[string]string `json:"teams,omitempty" bson:"teams"`
}

// NewUser is a func.
func (m *Mongo) NewUser(ctx context.Context, requestData user.SignUpRequest) (user.User, errors.ApiErr) {
	reqCtx, cancelFunc := db.ReqContextWithTimeout(ctx)
	defer cancelFunc()
	m.Initialize()
	err := m.client.Connect(reqCtx)
	if err != nil {
		log.Printf("couldn't connect to db on new user, %+v", err)
	}
	defer m.client.Disconnect(reqCtx)
	collection := m.db.Collection(CollectionUsers)
	userExists := DataAlreadyExists(reqCtx, collection, "name", requestData.Name)
	if userExists {
		log.Println("user already exists")
		return user.User{}, errors.NewBadRequestError(fmt.Sprintf("error creating new user; user with name %v already exists", requestData.Name))
	}
	APIKey, err := GenerateAPIKey()
	if err != nil {
		log.Println("error generating uuid")
		return user.User{}, errors.NewInternalServerError()
	}
	hashedPassword, err := password.HashPassword(requestData.Password)
	if err != nil {
		log.Println("error hashing password")
		return user.User{}, errors.NewInternalServerError()
	}
	signUpData := User{
		Name:      requestData.Name,
		Password:  hashedPassword,
		APIKey:    APIKey,
		Bookmarks: map[string]string{},
		Teams:     map[string]string{},
	}
	res, err := collection.InsertOne(reqCtx, signUpData)
	if err != nil {
		log.Printf("error creating new user with data: \n username: %v\n password: %v", requestData.Name, requestData.Password)
		return user.User{}, errors.NewInternalServerError()
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Println("error getting objectID from newly inserted user")
		return user.User{}, errors.NewInternalServerError()
	}
	newUserData := user.User{
		ID:     oid.Hex(),
		Name:   requestData.Name,
		APIKey: APIKey,
	}
	return newUserData, nil
}

// GenerateAPIKey generates a random URL-safe string of random length for use as an API key.
func GenerateAPIKey() (string, error) {
	key, err := uuid.NewRandom()
	return key.String(), err
}

// LogIn checks the users credentials returns the user if password is correct.
func (m *Mongo) LogIn(ctx context.Context, requestData user.LogInRequest) (user.User, errors.ApiErr) {
	reqCtx, cancelFunc := db.ReqContextWithTimeout(ctx)
	defer cancelFunc()
	m.Initialize()
	err := m.client.Connect(reqCtx)
	if err != nil {
		log.Println("couldn't connect to db on login")
	}
	defer m.client.Disconnect(reqCtx)
	collection := m.db.Collection(CollectionUsers)
	currUser, err := GetUserByKey(reqCtx, collection, "name", requestData.Name)
	if err != nil || !password.CheckHashedPassword(currUser.Password, requestData.Password) {
		log.Printf("login getuserbykey %+v", err)
		return user.User{}, errors.NewApiError(http.StatusUnauthorized, errors.ErrWrongCredentials.Error(), "error: name or password incorrect")
	}
	return user.User{
		ID:        currUser.ID,
		Name:      currUser.Name,
		APIKey:    currUser.APIKey,
		Bookmarks: currUser.Bookmarks,
		Teams:     currUser.Teams,
	}, nil
}

// GetTeams uses user id to get all users teams from the db.
func (m *Mongo) GetTeams(ctx context.Context, APIKey string) ([]user.Team, errors.ApiErr) {
	reqCtx, cancelFunc := db.ReqContextWithTimeout(ctx)
	defer cancelFunc()
	m.Initialize()
	err := m.client.Connect(reqCtx)
	if err != nil {
		log.Println("couldn't connect to db on login")
	}
	defer m.client.Disconnect(reqCtx)
	res, err := m.SessionWithTransaction(reqCtx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		userCollection := m.db.Collection(CollectionUsers)
		currUser, err := GetUserByKey(sessCtx, userCollection, "APIKey", APIKey)
		if err != nil {
			log.Printf("error getting user by APIKey: %s -> %+v\n", APIKey, err)
			return nil, err
		}
		teamIDs, err := convertIDs(currUser.Teams)
		if err != nil {
			log.Printf("error converting teams to ids: error -> %+v\n", err)
			return nil, err
		}
		teamCollection := m.db.Collection(CollectionTeams)
		filter := bson.M{"_id": bson.M{"$in": teamIDs}}
		opts := options.Find()
		teamCursor, err := teamCollection.Find(sessCtx, filter, opts)
		defer teamCursor.Close(sessCtx)
		if err != nil {
			log.Printf("error converting teams to ids -> %+v\n", err)
			return nil, err
		}
		var teams []user.Team
		for teamCursor.Next(sessCtx) {
			var currTeam user.Team
			if err := teamCursor.Decode(&currTeam); err != nil {
				log.Printf("error could not get team from found teams -> %+v\n", err)
				return nil, err
			}
			teams = append(teams, currTeam)
		}
		return teams, nil
	})
	if err != nil {
		log.Printf("error could not get data from transaction -> %+v\n", err)
		return nil, errors.NewInternalServerError()
	}
	teams, ok := res.([]user.Team)
	if !ok {
		log.Println("error could not assert type []Team")
		return nil, errors.NewInternalServerError()
	}
	return teams, nil
}

func convertIDs(teams map[string]string) ([]primitive.ObjectID, error) {
	output := make([]primitive.ObjectID, len(teams))
	for id := range teams {
		res, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		output = append(output, res)
	}
	return output, nil
}

// GetAllCmds uses req info to get all users current cmds from the db.
func (m *Mongo) GetAllCmds(ctx context.Context, APIKey string) (map[string]string, errors.ApiErr) {
	reqCtx, cancelFunc := db.ReqContextWithTimeout(ctx)
	defer cancelFunc()
	m.Initialize()
	err := m.client.Connect(reqCtx)
	if err != nil {
		log.Println("couldn't connect to db on login")
	}
	defer m.client.Disconnect(reqCtx)
	collection := m.db.Collection(CollectionUsers)

	user, err := GetUserByKey(reqCtx, collection, "APIKey", APIKey)
	if err != nil {
		return nil, errors.ParseGetUserError(APIKey, err)
	}
	return user.Bookmarks, nil
}

// AddCmd attempts to either add or update a cmd for the user, returning the number
// of updated cmds.
func (m *Mongo) AddCmd(reqCtx context.Context, requestData user.AddCmdRequest, APIKey string) (int, errors.ApiErr) {
	ctx, cancelFunc := db.ReqContextWithTimeout(reqCtx)
	defer cancelFunc()
	m.Initialize()
	err := m.client.Connect(ctx)
	if err != nil {
		log.Println("couldn't connect to db on login")
	}
	defer m.client.Disconnect(reqCtx)
	collection := m.db.Collection(CollectionUsers)

	result, err := AddCmdToUser(ctx, collection, requestData)
	if err != nil {
		return 0, errors.NewInternalServerError()
	}
	var numUpdated int
	if int(result.UpsertedCount) >= int(result.ModifiedCount) {
		numUpdated = int(result.UpsertedCount)
	} else {
		numUpdated = int(result.ModifiedCount)
	}
	return numUpdated, nil
}

// AddCmdToUser takes a given username along with the cmd and URL to set and adds the data to their bookmarks.
func AddCmdToUser(ctx context.Context, collection *mongo.Collection, requestData user.AddCmdRequest) (*mongo.UpdateResult, error) {
	opts := options.Update().SetUpsert(true)
	filter, err := primitive.ObjectIDFromHex(requestData.ID)
	if err != nil {
		return nil, err
	}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: fmt.Sprintf("bookmarks.%s", requestData.Cmd), Value: requestData.URL}}}}
	result, err := collection.UpdateByID(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DelCmd attempts to either rempve a cmd from the user, returning the number
// of updated cmds.
func (m *Mongo) DelCmd(ctx context.Context, requestData user.DelCmdRequest, APIKey string) (int, errors.ApiErr) {
	reqCtx, cancelFunc := db.ReqContextWithTimeout(ctx)
	defer cancelFunc()
	m.Initialize()
	err := m.client.Connect(reqCtx)
	if err != nil {
		log.Println("couldn't connect to db on login")
	}
	defer m.client.Disconnect(reqCtx)
	collection := m.db.Collection(CollectionUsers)
	result, err := RemoveCmdFromUser(reqCtx, collection, requestData.ID, requestData.Cmd)
	if err != nil {
		return 0, errors.NewInternalServerError()
	}
	return int(result.ModifiedCount), nil
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

// Delete attempts to delete a user from the db, returning the number of deleted users.
func (m *Mongo) Delete(reqCtx context.Context, requestData user.DelUserRequest, APIKey string) (int, errors.ApiErr) {
	ctx, cancelFunc := db.ReqContextWithTimeout(reqCtx)
	defer cancelFunc()
	m.Initialize()
	err := m.client.Connect(ctx)
	if err != nil {
		log.Println("couldn't connect to db on login")
	}
	defer m.client.Disconnect(reqCtx)
	collection := m.db.Collection(CollectionUsers)
	userData, err := GetByID(ctx, collection, requestData.ID)
	if err != nil {
		log.Printf("error deleting user: couldn't find user -> %v", err)
		return 0, errors.NewBadRequestError("could not find user to delete")
	}
	ok := password.CheckHashedPassword(userData.Password, requestData.Password)
	if !ok {
		log.Printf("error deleting user: password incorrect -> %v", err)
		return 0, errors.NewWrongCredentialsError("password incorrect")
	}
	result, err := DeleteUserFromDB(ctx, collection, requestData.ID)
	if err != nil {
		log.Printf("error deleting user: error -> %v", err)
		return 0, errors.NewInternalServerError()
	}
	if result.DeletedCount == 0 {
		log.Printf("could not remove user... maybe user:%s doesn't exists?", requestData.Name)
		return 0, errors.NewBadRequestError("error: could not remove cmd")
	}
	return int(result.DeletedCount), nil
}

// DeleteUserFromDB takes a given userID and removes the user from the database.
func DeleteUserFromDB(ctx context.Context, collection *mongo.Collection, userID string) (*mongo.DeleteResult, error) {
	opts := options.Delete().SetCollation(&options.Collation{
		Locale:    "en_US",
		Strength:  1,
		CaseLevel: false,
	})
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	result, err := collection.DeleteOne(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}