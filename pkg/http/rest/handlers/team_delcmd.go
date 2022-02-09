package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/conalli/bookshelf-backend/pkg/errors"
	"github.com/conalli/bookshelf-backend/pkg/team"
	"github.com/gorilla/mux"
)

type delTeamCmdResponse struct {
	NumDeleted int    `json:"numDeleted"`
	Cmd        string `json:"cmd"`
}

// DelTeamCmd is the handler for the team/delcmd endpoint. Checks credentials + JWT and if
// authorized deletes given cmd.
func DelTeamCmd(t team.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("DelCmd endpoint hit")
		vars := mux.Vars(r)
		APIKey := vars["APIKey"]
		var delCmdReq team.DelTeamCmdRequest
		json.NewDecoder(r.Body).Decode(&delCmdReq)

		result, err := t.DelCmdFromTeam(r.Context(), delCmdReq, APIKey)
		if err != nil {
			log.Printf("error returned while trying to remove a cmd: %v", err)
			errors.APIErrorResponse(w, err)
			return
		}
		if result == 0 {
			log.Printf("could not remove cmd... maybe %s doesn't exists?", delCmdReq.Cmd)
			err := errors.NewBadRequestError("error: could not remove cmd")
			errors.APIErrorResponse(w, err)
			return
		}
		log.Printf("successfully updates cmds: %s, removed %d", delCmdReq.Cmd, result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		res := delTeamCmdResponse{
			NumDeleted: result,
			Cmd:        delCmdReq.Cmd,
		}
		json.NewEncoder(w).Encode(res)
		return
	}
}