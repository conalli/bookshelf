package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/conalli/bookshelf-backend/internal/testutils"
	"github.com/conalli/bookshelf-backend/pkg/http/rest"
	"github.com/go-playground/validator/v10"
)

func TestGetCmds(t *testing.T) {
	t.Parallel()
	db := testutils.NewDB().AddDefaultUsers()
	r := rest.NewRouter(testutils.NewLogger(), validator.New(), db, testutils.NewCache())
	srv := httptest.NewServer(r.Handler())
	defer srv.Close()
	tc := []struct {
		name       string
		APIKey     string
		statusCode int
		res        map[string]string
	}{
		{
			name:       "Default user, correct request.",
			APIKey:     db.Users["1"].APIKey,
			statusCode: 200,
			res:        db.Users["1"].Cmds,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			res, err := testutils.RequestWithCookie("GET", srv.URL+"/api/user/cmd/"+c.APIKey, nil, c.APIKey, testutils.NewLogger())
			if err != nil {
				t.Fatalf("Couldn't create request to get cmds with cookie.")
			}
			want := 200
			if res.StatusCode != want {
				t.Errorf("Expected get cmd request to give status code %d: got %d", want, res.StatusCode)
			}
			defer res.Body.Close()
			var response map[string]string
			err = json.NewDecoder(res.Body).Decode(&response)
			if err != nil {
				t.Fatalf("Couldn't decode json body upon getting cmds.")
			}
			if fmt.Sprint(response) != fmt.Sprint(c.res) {
				t.Errorf("Expected commands to be %v: got %v", c.res, response)
			}
		})
	}
}
