package database

import (
	"encoding/json"
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

func (i *Instance) Chain(chain string) (models.Chain, error) {
	var c models.Chain

	err := i.d.DB.Get(&c, fmt.Sprintf("SELECT * FROM cns.chains WHERE chain_name='%s' limit 1;", chain))

	return c, err
}

func (i *Instance) Chains() ([]models.Chain, error) {
	var c []models.Chain

	return c, i.d.Exec(getAllChains, nil, &c)
}

func (i *Instance) UpdatePrimaryChannel(sourceChain, destChain, channel string) error {

	res, err := i.d.DB.Exec(fmt.Sprintf(`
	UPDATE cns.chains
	SET primary_channel = primary_channel || jsonb_build_object('%s' , '%s')
	WHERE chain_name='%s'
	`, destChain, channel, sourceChain))

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("update failed")
	}

	return nil
}

func (i *Instance) GetDenoms(chain string) (models.DenomList, error) {

	var l models.DenomList

	return l, i.d.Exec("select json_array_elements(denoms) from cns.chains where chain_name=:chain;", map[string]interface{}{
		"chain": chain,
	}, &l)
}

func (i *Instance) UpdateDenoms(chain string, denoms models.DenomList) error {

	b, err := json.Marshal(denoms)

	if err != nil {
		return err
	}

	res, err := i.d.DB.Exec(fmt.Sprintf(`
	UPDATE cns.chains
	SET denoms = '%s'::jsonb
	WHERE chain_name='%s'
	`, string(b), chain))

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("update failed")
	}

	return nil
}

type channelsBetweenChain struct {
	ChainAName             string `db:"chain_a_chain_name"`
	ChainAChannelID        string `db:"chain_a_channel_id"`
	ChainACounterChannelID string `db:"chain_a_counter_channel_id"`
	ChainAChainID          string `db:"chain_a_chain_id"`
	ChainAState            int    `db:"chain_a_state"`
	ChainBName             string `db:"chain_b_chain_name"`
	ChainBChannelID        string `db:"chain_b_channel_id"`
	ChainBCounterChannelID string `db:"chain_b_counter_channel_id"`
	ChainBChainID          string `db:"chain_b_chain_id"`
	ChainBState            int    `db:"chain_b_state"`
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
		// channel ID destination => channel ID on source
		ret[cc.ChainAChannelID] = cc.ChainBChannelID
	}

	return ret, nil
}

func (i *Instance) ChainAmount() (int, error) {
	var ret int
	return ret, i.d.DB.Get(&ret, "select count(id) from cns.chains")
}
