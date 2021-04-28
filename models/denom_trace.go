package models

type DenomTrace struct {
	Id        uint64 `db:"id"`
	ChainName string `db:"chain_name"`
	Path      string `db:"path"`
	BaseDenom string `db:"baseDenom"`
	Hash      string `db:"hash"`
}
