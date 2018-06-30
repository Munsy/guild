package models

//import bnet "github.com/mitchellh/go-bnet"

// BnetUser provides the battletag and ID number for a Battle.net account.
type BnetUser struct {
	ID		int	`json:"id"`
	BattleTag	string	`"json:battletag"`
}

