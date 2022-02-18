package team

// NewTeamRequest represents the expected JSON request for the /team POST endpoint.
type NewTeamRequest struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	TeamPassword string `json:"password"`
	ShortName    string `json:"shortName"`
}

// DelTeamRequest represents the expected JSON request for the /team DELETE endpoint.
type DelTeamRequest struct {
	ID           string `json:"id"`
	TeamID       string `json:"teamId"`
	TeamPassword string `json:"password"`
}

// AddMemberRequest represents the expected JSON request for the /team/addmember endpoint.
type AddMemberRequest struct {
	ID         string `json:"id"`
	TeamID     string `json:"teamId"`
	MemberName string `json:"memberName"`
	Role       string `json:"role"`
}

// DelSelfRequest represents the expected JSON request for the /team/delself endpoint.
type DelSelfRequest struct {
	ID     string `json:"id"`
	TeamID string `json:"teamId"`
}

// DelMemberRequest represents the expected JSON request for the /team/delmember endpoint.
type DelMemberRequest struct {
	ID         string `json:"id"`
	TeamID     string `json:"teamId"`
	MemberName string `jsong:"memberName"`
	Role       string `json:"role"`
}

// AddTeamCmdRequest represents the expected JSON request for the /team/addcmd endpoint.
type AddTeamCmdRequest struct {
	ID       string `json:"id"`
	MemberID string `json:"memberId"`
	Cmd      string `json:"cmd"`
	URL      string `json:"url"`
}

// DelTeamCmdRequest represents the expected JSON request for the /team/addcmd endpoint.
type DelTeamCmdRequest struct {
	ID       string `json:"id"`
	MemberID string `json:"memberId"`
	Cmd      string `json:"cmd"`
}
