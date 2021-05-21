package database

import (
	"github.com/allinbits/demeris-backend/models"
	"github.com/jmoiron/sqlx"
)

func (d *Database) Delegations(address string) ([]models.DelegationRow, error) {
	var delegations []models.DelegationRow

	q, args, err := sqlx.In("SELECT * FROM tracelistener.delegations WHERE delegator_address IN (?) and chain_name in (select chain_name from cns.chains where enabled=true);", []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return delegations, d.dbi.DB.Select(&delegations, q, args...)
}
