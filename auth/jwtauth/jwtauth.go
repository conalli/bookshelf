package jwtauth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/conalli/bookshelf-backend/models/apiErrors"
	"github.com/golang-jwt/jwt/v4"
)

var signingKey = []byte(os.Getenv("SIGNING_SECRET"))

// CustomClaims represents the claims made in the JWT.
type CustomClaims struct {
	Name string
	jwt.StandardClaims
}

// NewToken creates a new token based on the CustomClaims and returns the token
// as a string signed with the secret.
func NewToken(name string) (string, apiErrors.ApiErr) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		Name: name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
			NotBefore: time.Now().Add(5 * time.Second).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "http://bookshelf-backend.jp",
			Subject:   name,
		},
	})

	tkn, err := token.SignedString(signingKey)
	if err != nil {
		log.Printf("error when trying to sign token %+v", token)
		return tkn, apiErrors.NewInternalServerError()
	}
	return tkn, nil
}

// Authorized reads the JWT from the incoming request and returns whether the user is authorized or not.
// TODO: improve validation/ cookie handling
func Authorized(name string) func(w http.ResponseWriter, r *http.Request) bool {
	return func(w http.ResponseWriter, r *http.Request) bool {
		cookies := r.Cookies()
		if len(cookies) < 1 {
			log.Println("error: no cookies in request")
			return false
		}
		bookshelfCookie := filterCookies("bookshelfjwt", cookies)
		if bookshelfCookie == nil {
			log.Println("error: did not find bookshelf cookie")
			return false
		}
		token, err := jwt.ParseWithClaims(bookshelfCookie.Value, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) { return signingKey, nil })
		if err != nil {
			log.Printf("%+v", err)
			log.Println("error: failed to parse with claims")
			return false
		}
		tkn, ok := token.Claims.(*CustomClaims)
		if !ok {
			log.Println("error: failed to convert token to CustomClaims")
			return false
		}
		if err = tkn.Valid(); err != nil || tkn.Name != name || tkn.Subject != name {
			log.Println("error: token not valid")
			return false
		}
		fmt.Printf("%+v", tkn)
		return true
	}
}

func filterCookies(name string, cookies []*http.Cookie) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}