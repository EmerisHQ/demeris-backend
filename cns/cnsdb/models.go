// Code generated by sqlc. DO NOT EDIT.

package cnsdb

import (
	"github.com/allinbits/demeris-backend/models"
)

type Chain struct {
	ID               int32
	Enabled          bool
	ChainName        string
	ValidBlockThresh models.Threshold
	Logo             string
	DisplayName      string
	PrimaryChannel   models.DbStringMap
	Denoms           models.DenomList
	DemerisAddresses []string
	GenesisHash      string
	NodeInfo         models.NodeInfo
	DerivationPath   string
	BlockExplorer    string
}
