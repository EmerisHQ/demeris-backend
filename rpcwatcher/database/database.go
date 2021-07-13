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

func (i *Instance) AddChain(chain models.Chain) error {
	n, err := i.d.DB.PrepareNamed(insertChain)
	if err != nil {
		return err
	}

	res, err := n.Exec(chain)

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("database delete statement had no effect")
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

	return c, i.d.Exec(getAllChains, nil, &c)
}

type channelsBetweenChain struct {
	ChainAName             string `db:"chain_a_chain_name"`
	ChainAChannelID        string `db:"chain_a_channel_id"`
	ChainACounterChannelID string `db:"chain_a_counter_channel_id"`
	ChainAChainID          string `db:"chain_a_chain_id"`
	ChainBName             string `db:"chain_b_chain_name"`
	ChainBChannelID        string `db:"chain_b_channel_id"`
	ChainBCounterChannelID string `db:"chain_b_counter_channel_id"`
	ChainBChainID          string `db:"chain_b_chain_id"`
}

func (i *Instance) ChannelsBetweenChains(source, destination, chainID string) (map[string]string, error) {

	var c []channelsBetweenChain

	n, err := i.d.DB.PrepareNamed(channelsBetweenChains)
	if err != nil {
		return map[string]string{}, err
	}

	if err := n.Select(&c, map[string]interface{}{
		"source":      source,
		"destination": destination,
		"chainID":     chainID,
	}); err != nil {
		return map[string]string{}, err
	}

	ret := map[string]string{}

	for _, cc := range c {
		ret[cc.ChainAChannelID] = cc.ChainBChannelID
	}

	return ret, nil
}
