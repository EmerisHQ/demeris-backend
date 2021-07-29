package database

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

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

type ChannelMapping struct {
	ChannelID        string
	CounterChannelID string
}

type ByOldestChannel []ChannelMapping

func (a ByOldestChannel) Len() int      { return len(a) }
func (a ByOldestChannel) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByOldestChannel) Less(i, j int) bool {
	chanI, err := strconv.Atoi(strings.TrimPrefix(a[i].ChannelID, "channel-"))
	if err != nil {
		panic(err)
	}

	chanJ, err := strconv.Atoi(strings.TrimPrefix(a[j].ChannelID, "channel-"))
	if err != nil {
		panic(err)
	}
	return chanI < chanJ
}

func (i *Instance) ChannelsBetweenChains(source, destination, chainID string) ([]ChannelMapping, error) {

	var c []channelsBetweenChain

	n, err := i.d.DB.PrepareNamed(channelsBetweenChains)
	if err != nil {
		return nil, err
	}

	if err := n.Select(&c, map[string]interface{}{
		"source":      source,
		"destination": destination,
		"chainID":     chainID,
	}); err != nil {
		return nil, err
	}

	var ret []ChannelMapping

	for _, cc := range c {
		// channel ID destination => channel ID on source
		ret = append(ret, ChannelMapping{
			ChannelID:        cc.ChainAChannelID,
			CounterChannelID: cc.ChainBChannelID,
		})
	}

	sort.Sort(ByOldestChannel(ret))

	return ret, nil
}

func (i *Instance) ChainAmount() (int, error) {
	var ret int
	return ret, i.d.DB.Get(&ret, "select count(id) from cns.chains")
}

type ClientChannelAssociation struct {
	ChainAName      string `db:"chain_a_chain_name"`
	ChainAChannelID string `db:"chain_a_channel_id"`
	ChainAChainID   string `db:"chain_a_chain_id"`
	ChainAClientID  string `db:"chain_a_client_id"`
	ChainBName      string `db:"chain_b_chain_name"`
	ChainBChannelID string `db:"chain_b_channel_id"`
	ChainBChainID   string `db:"chain_b_chain_id"`
	ChainBClientID  string `db:"chain_b_client_id"`
}

func (i *Instance) ClientByChannelName(chainName, channelName, chainID string) (ClientChannelAssociation, error) {
	var ret ClientChannelAssociation

	q := clientIDsOnChannel

	n, err := i.d.DB.PrepareNamed(q)
	if err != nil {
		return ClientChannelAssociation{}, err
	}

	if err := n.Get(&ret, map[string]interface{}{
		"source":    chainName,
		"channelID": channelName,
		"chainID":   chainID,
	}); err != nil {
		return ClientChannelAssociation{}, err
	}

	return ret, nil
}

func (i *Instance) ClientByID(chainName, clientID string) (models.IBCClientStateRow, error) {
	var ret models.IBCClientStateRow

	q := queryClientByID

	n, err := i.d.DB.PrepareNamed(q)
	if err != nil {
		return ret, err
	}

	if err := n.Get(&ret, map[string]interface{}{
		"chainName": chainName,
		"clientID":  clientID,
	}); err != nil {
		return ret, err
	}

	return ret, nil
}
