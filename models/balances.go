package models

type Balance struct {
	Id        uint64 `db:"id"`
	ChainName string `db:"chain_name"`
	Address   string `db:"address"`
	Amount    string `db:"amount"`
	Denom     string `db:"denom"`
	Height    uint32 `db:"height"`
}
