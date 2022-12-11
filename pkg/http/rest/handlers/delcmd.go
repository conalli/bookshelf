package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/conalli/bookshelf-backend/pkg/errors"
	"github.com/conalli/bookshelf-backend/pkg/http/request"
	"github.com/conalli/bookshelf-backend/pkg/logs"
	"github.com/conalli/bookshelf-backend/pkg/services/accounts"
	"github.com/gorilla/mux"
)

// DeleteCmdResponse represents the data returned upon successfully deleting a cmd.
type DeleteCmdResponse struct {
	NumDeleted int    `json:"numDeleted"`
	Cmd        string `json:"cmd"`
}

// DeleteCmd is the handler for the delcmd endpoint. Checks credentials + JWT and if
// authorized deletes given cmd.
func DeleteCmd(u accounts.UserService, log logs.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		APIKey := vars["APIKey"]
		delCmdReq, parseErr := request.DecodeJSONRequest[request.DeleteCmd](r.Body)
		if parseErr != nil {
			errRes := errors.NewBadRequestError("could not parse request body")
			errors.APIErrorResponse(w, errRes)
		}
		result, err := u.DeleteCmd(r.Context(), delCmdReq, APIKey)
		if err != nil {
			log.Errorf("error returned while trying to remove a cmd: %v", err)
			errors.APIErrorResponse(w, err)
			return
		}
		if result == 0 {
			log.Errorf("could not remove cmd... maybe %s doesn't exists?", delCmdReq.Cmd)
			err := errors.NewBadRequestError("error: could not remove cmd")
			errors.APIErrorResponse(w, err)
			return
		}
		log.Infof("successfully updates cmds: %s, removed %d", delCmdReq.Cmd, result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		res := DeleteCmdResponse{
			NumDeleted: result,
			Cmd:        delCmdReq.Cmd,
		}
		json.NewEncoder(w).Encode(res)
	}
}
