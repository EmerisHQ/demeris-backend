package rpcwatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tldb "github.com/allinbits/demeris-backend/api/database"
	cnsdb "github.com/allinbits/demeris-backend/cns/database"
	"github.com/allinbits/demeris-backend/models"
	dbutils "github.com/allinbits/demeris-backend/utils/database"

	"github.com/allinbits/demeris-backend/utils/store"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"go.uber.org/zap"
)

const (
	blEvents = "tm.event='NewBlock'"
	txEvent  = "tm.event='Tx'"
)

type Watcher struct {
	Name            string
	apiUrl          string
	client          *client.WSClient
	db              *tldb.Database
	d               *dbutils.Instance
	cns             *cnsdb.Instance
	l               *zap.SugaredLogger
	store           *store.Store
	stopReadChannel chan struct{}
	DataChannel     chan coretypes.ResultEvent
}

type WsResponse struct {
	Event coretypes.ResultEvent `json:"result"`
}

type IbcTransitData struct {
	SourceChain              string `json:"sourceChain"`
	DestChain                string `json:"destChain"`
	SendPacketSourceChannel  string `json:"sourceChannel"`  // send_packet.packet_src_channel
	SendPacketPacketSequence string `json:"packetSequence"` // send_packet.packet_sequence
}

type IbcReceiveData struct {
	SourceChain              string `json:"sourceChain"`
	DestChain                string `json:"destChain"`
	RecvPacketSourceChannel  string `json:"sourceChannel"`  // recv_packet.packet_src_channel
	RecvPacketPacketSequence string `json:"packetSequence"` // write_acknowledgement.packet_sequence
}

type Events map[string][]string

func NewWatcher(endpoint, chainName string, logger *zap.SugaredLogger, apiUrl string, db *dbutils.Instance, tldb *tldb.Database, cnsdb *cnsdb.Instance, s *store.Store, subscriptions []string) (*Watcher, error) {

	ws, err := client.NewWS(endpoint, "/websocket")
	if err != nil {
		return nil, err
	}

	if err := ws.Start(); err != nil {
		return nil, err
	}

	w := &Watcher{
		apiUrl:          apiUrl,
		d:               db,
		db:              tldb,
		cns:             cnsdb,
		client:          ws,
		l:               logger,
		store:           s,
		Name:            chainName,
		stopReadChannel: make(chan struct{}),
		DataChannel:     make(chan coretypes.ResultEvent),
	}

	for _, sub := range subscriptions {
		if err := w.client.Subscribe(context.Background(), sub); err != nil {
			return nil, fmt.Errorf("failed to subscribe, %w", err)
		}
	}

	go w.readChannel()

	return w, nil
}

func Start(watcher *Watcher, ctx context.Context) {
	go watcher.startChain(ctx)
}

func (w *Watcher) readChannel() {
	/*
		This thing uses nested selects because when we read from tendermint data channel, we should check first if
		the cancellation function has been called, and if yes we should return.

		Only after having done such check we can process the tendermint data.
	*/
	for {
		select {
		case <-w.stopReadChannel:
			return
		default:
			select {
			case data := <-w.client.ResponsesCh:
				if data.Error != nil {
					w.l.Errorw("error from tendermint rpc", "error", data.Error.Error(), "chain", w.Name)
					continue
				}

				e := coretypes.ResultEvent{}
				err := json.Unmarshal(data.Result, &e)
				if err != nil {
					w.l.Errorw("cannot unmarshal data into resultevent", "error", err, "chain", w.Name)
					continue
				}

				w.l.Debugw("got message to handle", "chain name", w.Name)

				go func() {
					w.DataChannel <- e
				}()
			}
		}
	}
}

func (w *Watcher) handleMessage(data coretypes.ResultEvent) {
	txHashSlice, exists := data.Events["tx.hash"]
	_, isCreateLP := data.Events["create_pool.pool_name"]
	_, isIBC := data.Events["ibc_transfer.sender"]
	_, isIBCSuccess := data.Events["fungible_token_packet.success"]
	_, isIBCRecv := data.Events["recv_packet.packet_sequence"]
	_, isIBCTimeout := data.Events["timeout.refund_receiver"]

	if len(txHashSlice) == 0 {
		return
	}

	txHash := txHashSlice[0]

	key := fmt.Sprintf("%s-%s", w.Name, txHash)

	w.l.Debugw("got message to handle", "chain name", w.Name, "key", key, "id create lp", isCreateLP, "is ibc", isIBC, "is ibc recv", isIBCRecv,
		"is ibc success", isIBCSuccess, "is ibc timeout", isIBCTimeout)

	w.l.Debugw("is simple ibc transfer", "is it", exists && !isCreateLP && !isIBC && !isIBCRecv && w.store.Exists(key))
	// Handle case where a simple non-IBC transfer is being used.
	if exists && !isCreateLP && !isIBC && !isIBCRecv && w.store.Exists(key) {
		if err := w.store.SetComplete(key); err != nil {
			w.l.Errorw("cannot set complete", "chain name", w.Name, "error", err)
		}
		return
	}

	// Handle case where an LP is being created on the Cosmos Hub

	if isCreateLP && w.Name == "cosmos-hub" {

		chain, err := w.cns.Chain(w.Name)

		if err != nil {
			w.l.Errorw("can't find chain", "error", err)
			return
		}

		poolCoinDenom, ok := data.Events["create_pool.pool_coin_denom"]

		if !ok {
			w.l.Errorw("no field create_pool.pool_coin_denom in Events", "error", err)
			return
		}

		dd, err := formatDenom(w, data)

		if err != nil {
			w.l.Errorw("failed to format denom", "error", err)
			return
		}

		found := false

		for _, token := range chain.Denoms {
			if token.Name == poolCoinDenom[0] {
				token = dd
				found = true
			}
		}

		if !found {
			chain.Denoms = append(chain.Denoms, dd)
		}

		err = w.cns.AddChain(chain)

		if err != nil {
			w.l.Errorw("failed to update chain", "error", err)
		}

		return

	}

	// Handle case where an IBC transfer is sent from the origin chain.
	if isIBC {

		sendPacketSourcePort, ok := data.Events["send_packet.packet_src_port"]

		if !ok {
			w.l.Errorf("send_packet.packet_src_port not found")
			return
		}

		if sendPacketSourcePort[0] != "transfer" {
			w.l.Errorf("port is not 'transfer', ignoring")
			return
		}

		sendPacketSourceChannel, ok := data.Events["send_packet.packet_src_channel"]

		if !ok {
			w.l.Errorf("send_packet.packet_src_channel not found")
			return
		}

		sendPacketSequence, ok := data.Events["send_packet.packet_sequence"]

		if !ok {
			w.l.Errorf("send_packet.packet_sequence not found")
			return
		}

		counterparty, ok := data.Events["send_packet.packet_dst_channel"]
		if !ok {
			w.l.Errorf("send_packet.packet_dst_channel not found")
			return
		}

		w.store.SetInTransit(key, counterparty[0], sendPacketSourceChannel[0], sendPacketSequence[0])
		return
	}

	// Handle case where IBC transfer is received by the receiving chain.
	if isIBCSuccess {
		if isIBCRecv {
			recvPacketSourcePort, ok := data.Events["recv_packet.packet_src_port"]

			if !ok {
				w.l.Errorf("recv_packet.packet_src_port not found")
				return
			}

			if recvPacketSourcePort[0] != "transfer" {
				w.l.Errorf("port is not 'transfer', ignoring")
				return
			}

			recvPacketSourceChannel, ok := data.Events["recv_packet.packet_src_channel"]

			if !ok {
				w.l.Errorf("recv_packet.packet_src_channel not found")
				return
			}

			recvPacketSequence, ok := data.Events["recv_packet.packet_sequence"]

			if !ok {
				w.l.Errorf("recv_packet.packet_sequence not found")
				return
			}

			key := fmt.Sprintf("%s-%s-%s", w.Name, recvPacketSourceChannel[0], recvPacketSequence[0])
			w.store.SetIbcReceived(key)
			return
		}

		successAck, ok := data.Events["fungible_token_packet.success"]
		if !ok {
			w.l.Errorf("success ack not found")
			return
		}

		if successAck[0] == "false" {
			w.store.SetIbcFailed(key)
			return
		}
	}

	if isIBCTimeout {
		_, ok := data.Events["timeout.refund_receiver"]
		if !ok {
			w.l.Errorf("refund receiver not found")
			return
		}

		w.store.SetIbcTimeout(key)
		return
	}

}

func (w *Watcher) startChain(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.stopReadChannel <- struct{}{}
			w.l.Infof("watcher %s has been canceled", w.Name)
			return
		default:
			select {
			case data := <-w.DataChannel:
				w.handleMessage(data)
			}
		}

	}
}

func paths(path string) ([]string, error) {
	numSlash := strings.Count(path, "/")
	if numSlash == 1 {
		return []string{path}, nil
	}

	if numSlash%2 == 0 {
		return nil, fmt.Errorf("malformed path")
	}

	spl := strings.Split(path, "/")

	var paths []string
	pathBuild := ""

	for i, e := range spl {
		if i%2 != 0 {
			pathBuild = pathBuild + "/" + e
			paths = append(paths, pathBuild)
			pathBuild = ""
		} else {
			pathBuild = e
		}
	}

	return paths, nil
}

func formatDenom(w *Watcher, data coretypes.ResultEvent) (models.Denom, error) {
	d := models.Denom{}

	poolCoinDenom, ok := data.Events["create_pool.pool_coin_denom"]

	if !ok {
		return d, fmt.Errorf("failed to read pool coin denom")
	}

	d.Name = poolCoinDenom[0]

	depositCoins, ok := data.Events["create_pool.deposit_coins"]
	if !ok {
		return d, fmt.Errorf("failed to read deposit coins")
	}

	coins, err := sdktypes.ParseCoinsNormalized(depositCoins[0])
	chain, err := w.cns.Chain("cosmos-hub")

	if err != nil {
		return d, err
	}

	coinADenom := coins[0].Denom
	coinBDenom := coins[1].Denom
	d.Name = poolCoinDenom[0]
	if "pool" == coinADenom[:4] || "pool" == coinBDenom[:4] {
		d.DisplayName = fmt.Sprintf("GDEX %s LP", poolCoinDenom[0])
		d.Verified = false
		d.Ticker = fmt.Sprintf("G-%s", poolCoinDenom[0])

		return d, nil
	}

	var tokenTickers []string

	if "ibc/" == coinADenom[:4] {
		// verify trace
		denomTrace, err := w.db.DenomTrace("cosmos-hub", coinADenom[4:])
		if err != nil {
			return d, err
		}

		pathsElements, err := paths(denomTrace.Path)
		if err != nil {
			return d, err
		}

		if len(pathsElements) == 1 {
			// only verify single hop tokens

			denomTicker := getDisplayTicker(chain, denomTrace.BaseDenom)
			tokenTickers = append(tokenTickers, denomTicker)
			d.Verified = true

		} else {
			d.Verified = false
		}

	} else {
		denomTicker := getDisplayTicker(chain, coinADenom)
		tokenTickers = append(tokenTickers, denomTicker)

	}

	if "ibc/" == coinBDenom[:4] {
		// verify trace
		denomTrace, err := w.db.DenomTrace("cosmos-hub", coinBDenom[4:])
		if err != nil {
			return d, err
		}

		pathsElements, err := paths(denomTrace.Path)
		if err != nil {
			return d, err
		}

		if len(pathsElements) == 1 {
			// only verify single hop tokens

			denomTicker := getDisplayTicker(chain, denomTrace.BaseDenom)
			tokenTickers = append(tokenTickers, denomTicker)
			d.Verified = true

		} else {
			d.Verified = false
		}

	} else {
		denomTicker := getDisplayTicker(chain, coinBDenom)
		tokenTickers = append(tokenTickers, denomTicker)
	}

	d.DisplayName = fmt.Sprintf("GDEX %s/%s LP", tokenTickers[0], tokenTickers[1])
	d.Ticker = fmt.Sprintf("G-%s-%s", tokenTickers[0], tokenTickers[1])

	return d, nil
}

func getDisplayTicker(c models.Chain, denom string) string {
	for _, d := range c.Denoms {
		if d.Name == denom {
			if d.DisplayName != "" {
				return d.DisplayName
			}
		}
	}
	return denom
}
