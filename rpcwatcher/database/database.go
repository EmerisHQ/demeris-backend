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

	return ii, nil
}

func (i *Instance) UpdateDenoms(chain models.Chain) error {
	n, err := i.d.DB.PrepareNamed(`UPDATE cns.chains 
	SET denoms=:denoms 
	WHERE chain_name=:chain_name;`)
	if err != nil {
		return err
	}

	res, err := n.Exec(chain)

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("database update statement had no effect")
	}

	return nil
}

func (i *Instance) Chain(chain string) (models.Chain, error) {
	var c models.Chain

	err := i.d.DB.Get(&c, fmt.Sprintf("SELECT * FROM cns.chains WHERE chain_name='%s' limit 1;", chain))

	return c, err
}

func (i *Instance) Chains() ([]models.Chain, error) {
	var c []models.Chain

	return c, i.d.Exec("SELECT * FROM cns.chains", nil, &c)
}

func (i *Instance) GetCounterParty(chain, srcChannel string) ([]models.ChannelQuery, error) {
	var c []models.ChannelQuery

	q, err := i.d.DB.PrepareNamed("select chain_name, json_data.* from cns.chains, jsonb_each_text(primary_channel) as json_data where chain_name=:chain_name and value=:channel limit 1;")
	if err != nil {
		return []models.ChannelQuery{}, err
	}

	if err := q.Select(&c, map[string]interface{}{
		"chain_name": chain,
		"channel":    srcChannel,
	}); err != nil {
		return []models.ChannelQuery{}, err
	}

	if len(c) == 0 {
		return nil, fmt.Errorf("no counterparty found for chain %s on channel %s", chain, srcChannel)
	}

	return c, nil
}
