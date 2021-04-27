package database

import (
	"github.com/jmoiron/sqlx"
)

type Delegation struct {
	Id               uint64 `db:"id"`
	DelegatorAddress string `db:"delegator_address"`
	ValidatorAddress string `db:"validator_address"`
	Amount           string `db:"amount"`
}

func (d *Database) Delegations(address string) ([]Delegation, error) {
	var delegations []Delegation

	q, args, err := sqlx.In("SELECT * FROM tracelistener.delegations WHERE delegator_address IN (?);", []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return delegations, d.dbi.DB.Select(&delegations, q, args...)
}
