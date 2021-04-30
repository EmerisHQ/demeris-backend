package models

import (
	"encoding/json"
	"errors"
)

// Chain represents CNS chain metadata row on the database.
type Chain struct {
	ID                uint64      `db:"id" json:"-"`
	ChainName         string      `db:"chain_name" binding:"required" json:"chain_name"`                 // the unique name of the chain
	CounterpartyNames DbStringMap `db:"counterparty_names" binding:"required" json:"counterparty_names"` // a mapping of client_id to chain names used to identify which chain a given client_id corresponds to
	PrimaryChannel    DbStringMap `db:"primary_channel" binding:"required" json:"primary_channel"`       // a mapping of chain name to primary channel
	NativeDenoms      DenomList   `db:"native_denoms" binding:"required" json:"native_denoms"`           // a list of denoms native to the chain
	FeeTokens         DenomList   `db:"fee_tokens" binding:"required" json:"fee_tokens"`                 // a list of denoms accepted as fee on the chain, fee tokens must be verified
	FeeAddress        string      `db:"fee_address" binding:"required" json:"fee_address"`               // the address on which we accept fee payments
	PriceModifier     float64     `db:"price_modifier" binding:"required" json:"price_modifier"`         // modifier (between 0 and 1) applied when estimating the price of a token hopping through the chain
	BaseIBCFee        float64     `db:"base_ibc_fee" binding:"required" json:"base_ibc_fee"`             // average cost (in dollar) to submit an IBC transaction to the chain
	GenesisHash       string      `db:"genesis_hash" binding:"required" json:"genesis_hash"`             // hash of the chain's genesis file
	NodeInfo          NodeInfo    `db:"node_info" binding:"required" json:"node_info"`                   // info required to query full-node (e.g. to submit tx)
}

// VerifiedFeeTokens returns a DenomList of fee tokens that are verified.
func (c Chain) VerifiedFeeTokens() DenomList {
	var ret DenomList
	for _, ft := range c.FeeTokens {
		if !ft.Verified {
			continue
		}

		ret = append(ret, ft)
	}

	return ret
}

// VerifiedNativeDenoms returns a DenomList of native denoms that are verified.
func (c Chain) VerifiedNativeDenoms() DenomList {
	var ret DenomList
	for _, ft := range c.NativeDenoms {
		if !ft.Verified {
			continue
		}

		ret = append(ret, ft)
	}

	return ret
}

// NodeInfo holds information useful to connect to a full node and broadcast transactions.
type NodeInfo struct {
	Endpoint string `binding:"required" json:"endpoint"`
	ChainID  string `binding:"required" json:"chain_id"`
}

// Scan is the sql.Scanner implementation for DbStringMap.
func (a *NodeInfo) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// Denom holds a token denomination and its verification status.
type Denom struct {
	Name     string `db:"name" binding:"required" json:"name"`
	Verified bool   `db:"verified" binding:"required" json:"verified"`
}

// DenomList represents a slice of Denom.
type DenomList []Denom

// Scan is the sql.Scanner implementation for DenomList.
func (a *DenomList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// DbStringMap represent a JSON database-enabled string map.
type DbStringMap map[string]string

// Scan is the sql.Scanner implementation for DbStringMap.
func (a *DbStringMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}
