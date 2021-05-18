package database

import (
	"fmt"

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

func (i *Instance) DeleteChain(chain string) error {
	n, err := i.d.DB.PrepareNamed(deleteChain)
	if err != nil {
		return err
	}

	res, err := n.Exec(map[string]interface{}{
		"chain_name": chain,
	})

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("database delete statement had no effect")
	}

	return nil
}

func (i *Instance) Chains() ([]models.Chain, error) {
	var c []models.Chain

	return c, i.d.Exec(getAllChains, nil, &c)
}
