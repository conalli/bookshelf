package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/conalli/bookshelf-backend/models/errors"
	"github.com/conalli/bookshelf-backend/models/requests"
	"github.com/conalli/bookshelf-backend/models/responses"
	"github.com/conalli/bookshelf-backend/models/team"
)

// AddTeamCmd is the handler for the team/addcmd endpoint. Checks credentials + JWT and if
// authorized sets new cmd.
func AddTeamCmd(w http.ResponseWriter, r *http.Request) {
	log.Println("SetCmd endpoint hit")
	var setCmdReq requests.AddTeamCmdRequest
	json.NewDecoder(r.Body).Decode(&setCmdReq)

	numUpdated, err := team.AddCmd(r.Context(), setCmdReq)
	if err != nil {
		log.Printf("error returned while trying to add a new cmd: %v", err)
		errors.APIErrorResponse(w, err)
		return
	}
	if numUpdated == 0 {
		log.Printf("could not update cmds... maybe %s:%s already exists?", setCmdReq.Cmd, setCmdReq.URL)
		err := errors.NewBadRequestError("error: could not update cmds")
		errors.APIErrorResponse(w, err)
		return
	}
	log.Printf("successfully set cmd: %s, url: %s", setCmdReq.Cmd, setCmdReq.URL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	res := responses.AddCmdResponse{
		CmdsSet: numUpdated,
		Cmd:     setCmdReq.Cmd,
		URL:     setCmdReq.URL,
	}
	json.NewEncoder(w).Encode(res)
	return
}