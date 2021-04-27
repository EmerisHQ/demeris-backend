package database

import (
	dbutils "github.com/allinbits/demeris-backend/utils/database"

	navigator_cns "github.com/allinbits/demeris-backend/cns"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Instance struct {
	d          *dbutils.Instance
	connString string
}

func New(connString string) (*Instance, error) {
	i, err := dbutils.New(connString)
	if err != nil {
		return nil, err
	}

	ii := &Instance{
		d:          i,
		connString: connString,
	}

	ii.runMigrations()
	return ii, nil
}

func (i *Instance) AddChain(chain navigator_cns.Chain) error {
	return i.d.Exec(insertChain, &chain, nil)
}

func (i *Instance) Chains() ([]navigator_cns.Chain, error) {
	var c []navigator_cns.Chain

	return c, i.d.Exec(getAllChains, nil, &c)
}
