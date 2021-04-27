package models

import (
	"encoding/json"
	"errors"
)

// Chain represents CNS chain metadata row on the database.
type Chain struct {
	ID                uint64              `db:"id" json:"-"`
	ChainName         string              `db:"chain_name" binding:"required" json:"chain_name"`                 // the unique name of the chain
	CounterpartyNames CounterpartyMapping `db:"counterparty_names" binding:"required" json:"counterparty_names"` // a mapping of client_id to chain names used to identify which chain a given client_id corresponds to
	NativeDenoms      DenomList           `db:"native_denoms" binding:"required" json:"native_denoms"`           // a list of denoms native to the chain
	FeeTokens         DenomList           `db:"fee_tokens" binding:"required" json:"fee_tokens"`                 // a list of denoms accepted as fee on the chain, fee tokens must be verified
	PriceModifier     float64             `db:"price_modifier" binding:"required" json:"price_modifier"`         // modifier (between 0 and 1) applied when estimating the price of a token hopping through the chain
	BaseIBCFee        float64             `db:"base_ibc_fee" binding:"required" json:"base_ibc_fee"`             // average cost (in dollar) to submit an IBC transaction to the chain
	GenesisHash       string              `db:"genesis_hash" binding:"required" json:"genesis_hash"`             // hash of the chain's genesis file
	//nodeInfo nodeInfo // info required to query full-node (e.g. to submit tx)
}

// Denom holds a token denomination and its verification status.
type Denom struct {
	Name       string `db:"name" binding:"required" json:"name"`
	IsVerified bool   `db:"is_verified" binding:"required" json:"is_verified"`
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

// CounterpartyMapping represent a mapping between IBC client IDs and chain names.
type CounterpartyMapping map[string]string

// Scan is the sql.Scanner implementation for CounterpartyMapping.
func (a *CounterpartyMapping) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}
