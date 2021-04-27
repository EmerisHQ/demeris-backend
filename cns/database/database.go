package database

import (
	"github.com/allinbits/demeris-backend/models"
	dbutils "github.com/allinbits/demeris-backend/utils/database"

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

func (i *Instance) AddChain(chain models.Chain) error {
	return i.d.Exec(insertChain, &chain, nil)
}

func (i *Instance) Chains() ([]models.Chain, error) {
	var c []models.Chain

	return c, i.d.Exec(getAllChains, nil, &c)
}
