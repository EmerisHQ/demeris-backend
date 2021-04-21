package database

import (
	"fmt"
	"strings"
)

type Balance struct {
	Id        uint64 `db:"id"`
	ChainName string `db:"chain_name"`
	Address   string `db:"address"`
	Amount    uint64 `db:"amount"`
	Denom     string `db:"denom"`
	Height    uint32 `db:"height"`
}

func (d *Database) Balances(addresses []string) ([]Balance, error) {
	balances := []Balance{}

	for i, address := range addresses {
		addresses[i] = fmt.Sprintf("'%s'", address)
	}

	q := fmt.Sprintf("SELECT * FROM tracelistener.balances WHERE address IN (%s)", strings.Join(addresses, ","))

	return balances, d.dbi.Exec(q, nil, &balances)
}
