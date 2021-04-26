package database

import (
	"github.com/jmoiron/sqlx"
)

// TODO: move this stuff to the database package

type Balance struct {
	Id        uint64 `db:"id"`
	ChainName string `db:"chain_name"`
	Address   string `db:"address"`
	Amount    string `db:"amount"`
	Denom     string `db:"denom"`
	Height    uint32 `db:"height"`
}

func (d *Database) Balances(addresses []string) ([]Balance, error) {
	var balances []Balance

	q, args, err := sqlx.In("SELECT * FROM tracelistener.balances WHERE address IN (?);", addresses)
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return balances, d.dbi.DB.Select(&balances, q, args...)
}
