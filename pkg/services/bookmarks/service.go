package bookmarks

import (
	"context"
	"net/http"

	"github.com/conalli/bookshelf-backend/pkg/errors"
	"github.com/conalli/bookshelf-backend/pkg/http/request"
	"github.com/conalli/bookshelf-backend/pkg/logs"
	"github.com/go-playground/validator/v10"
)

type Service interface {
	GetAllBookmarks(ctx context.Context, APIKey string) (*Folder, errors.APIErr)
	GetBookmarksFolder(ctx context.Context, path, APIKey string) (*Folder, errors.APIErr)
	AddBookmark(ctx context.Context, requestData request.AddBookmark, APIKey string) (int, errors.APIErr)
	AddBookmarksFromFile(ctx context.Context, r *http.Request, APIKey string) (int, errors.APIErr)
	DeleteBookmark(ctx context.Context, requestData request.DeleteBookmark, APIKey string) (int, errors.APIErr)
}

type Repository interface {
	GetAllBookmarks(ctx context.Context, APIKey string) ([]Bookmark, errors.APIErr)
	GetBookmarksFolder(ctx context.Context, path, APIKey string) ([]Bookmark, errors.APIErr)
	AddBookmark(ctx context.Context, requestData request.AddBookmark, APIKey string) (int, errors.APIErr)
	AddManyBookmarks(ctx context.Context, bookmarks []Bookmark) (int, errors.APIErr)
	DeleteBookmark(ctx context.Context, requestData request.DeleteBookmark, APIKey string) (int, errors.APIErr)
}

type service struct {
	log      logs.Logger
	validate *validator.Validate
	db       Repository
}

func NewService(l logs.Logger, v *validator.Validate, db Repository) *service {
	return &service{l, v, db}
}

func (s *service) GetAllBookmarks(ctx context.Context, APIKey string) (*Folder, errors.APIErr) {
	reqCtx, cancelFunc := request.CtxWithDefaultTimeout(ctx)
	defer cancelFunc()
	validateErr := s.validate.Var(APIKey, "uuid")
	if validateErr != nil {
		s.log.Errorf("Could not validate GET ALL BOOKMARKS request: %v", validateErr)
		return nil, errors.NewBadRequestError("request format incorrect.")
	}
	books, err := s.db.GetAllBookmarks(reqCtx, APIKey)
	folder := organizeBookmarks(books, "", BookmarksBasePath, BookmarksBasePath, BookmarksBasePath)
	return folder, err
}

func (s *service) GetBookmarksFolder(ctx context.Context, path, APIKey string) (*Folder, errors.APIErr) {
	reqCtx, cancelFunc := request.CtxWithDefaultTimeout(ctx)
	defer cancelFunc()
	validateErr := s.validate.Var(APIKey, "uuid")
	if validateErr != nil {
		s.log.Errorf("Could not validate GET BOOKMARKS FOLDER request: %v", validateErr)
		return nil, errors.NewBadRequestError("request format incorrect.")
	}
	books, err := s.db.GetBookmarksFolder(reqCtx, path, APIKey)
	folder := organizeBookmarks(books, "", BookmarksBasePath, BookmarksBasePath, BookmarksBasePath)
	return folder, err
}

// AddBookmark adds a bookmark for an account.
func (s *service) AddBookmark(ctx context.Context, requestData request.AddBookmark, APIKey string) (int, errors.APIErr) {
	reqCtx, cancelFunc := request.CtxWithDefaultTimeout(ctx)
	defer cancelFunc()
	validateReqErr := s.validate.Struct(requestData)
	validateAPIKeyErr := s.validate.Var(APIKey, "uuid")
	if validateReqErr != nil || validateAPIKeyErr != nil {
		s.log.Errorf("Could not validate ADD BOOKMARK request: %v - %v", validateReqErr, validateAPIKeyErr)
		return 0, errors.NewBadRequestError("request format incorrect.")
	}
	numUpdated, err := s.db.AddBookmark(reqCtx, requestData, APIKey)
	return numUpdated, err
}

func (s *service) AddBookmarksFromFile(ctx context.Context, r *http.Request, APIKey string) (int, errors.APIErr) {
	reqCtx, cancelFunc := request.CtxWithDefaultTimeout(ctx)
	defer cancelFunc()
	header, ok := r.MultipartForm.File[BookmarksFileKey]
	if !ok || len(header) != 1 {
		s.log.Error("Could not find bookmarks_file in request")
		return 0, errors.NewBadRequestError("no bookmark file in request")
	}
	file, err := header[0].Open()
	defer file.Close()
	if err != nil {
		s.log.Error("Could not open open bookmarks_file")
		return 0, errors.NewInternalServerError()
	}
	bookmarks, err := NewHTMLBookmarkParser(file, APIKey).parseBookmarkFileHTML()
	if err != nil {
		s.log.Error("Could not parse bookmarks_file")
		return 0, errors.NewBadRequestError("could not parse bookmark file")
	}
	numAdded, apierr := s.db.AddManyBookmarks(reqCtx, bookmarks)
	if err != nil {
		s.log.Error("Could not add bookmarks to db")
		return 0, apierr
	}
	return numAdded, nil
}

// DeleteBookmark removes a bookmark from an account.
func (s *service) DeleteBookmark(ctx context.Context, requestData request.DeleteBookmark, APIKey string) (int, errors.APIErr) {
	reqCtx, cancelFunc := request.CtxWithDefaultTimeout(ctx)
	defer cancelFunc()
	validateReqErr := s.validate.Struct(requestData)
	validateAPIKeyErr := s.validate.Var(APIKey, "uuid")
	if validateReqErr != nil || validateAPIKeyErr != nil {
		s.log.Errorf("Could not validate DELETE BOOKMARK request: %v - %v", validateReqErr, validateAPIKeyErr)
		return 0, errors.NewBadRequestError("request format incorrect")
	}
	numUpdated, err := s.db.DeleteBookmark(reqCtx, requestData, APIKey)
	return numUpdated, err
}
