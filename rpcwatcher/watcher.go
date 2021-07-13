package rpcwatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

type VerifyTraceResponse struct {
	VerifyTrace struct {
		IbcDenom  string `json:"ibc_denom"`
		BaseDenom string `json:"base_denom"`
		Verified  bool   `json:"verified"`
		Path      string `json:"path"`
		Trace     []struct {
			Channel          string `json:"channel"`
			Port             string `json:"port"`
			ChainName        string `json:"chain_name"`
			CounterpartyName string `json:"counterparty_name"`
		} `json:"trace"`
	} `json:"verify_trace"`
}

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

	w.l.Debugw("api url", "url", apiUrl)

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
	_, createPoolEventPresent := data.Events["create_pool.pool_name"]
	_, ibcTransferEventPresent := data.Events["ibc_transfer.sender"]
	_, ibcSuccessEventPresent := data.Events["fungible_token_packet.success"]
	_, ibcReciveEventPresent := data.Events["recv_packet.packet_sequence"]
	_, ibcTimeoutEventPresent := data.Events["timeout.refund_receiver"]

	if len(txHashSlice) == 0 {
		return
	}

	txHash := txHashSlice[0]

	key := fmt.Sprintf("%s-%s", w.Name, txHash)

	w.l.Debugw("got message to handle", "chain name", w.Name, "key", key, "is create lp", createPoolEventPresent, "is ibc", ibcTransferEventPresent, "is ibc recv", ibcReciveEventPresent,
		"is ibc success", ibcSuccessEventPresent, "is ibc timeout", ibcTimeoutEventPresent)

	w.l.Debugw("is simple ibc transfer", "is it", exists && !createPoolEventPresent && !ibcTransferEventPresent && !ibcReciveEventPresent && w.store.Exists(key))
	// Handle case where a simple non-IBC transfer is being used.
	if exists && !createPoolEventPresent && !ibcTransferEventPresent && !ibcReciveEventPresent && w.store.Exists(key) {
		if err := w.store.SetComplete(key); err != nil {
			w.l.Errorw("cannot set complete", "chain name", w.Name, "error", err)
		}
		return
	}

	w.l.Debugw("is create lp", "is it", createPoolEventPresent)

	// Handle case where an LP is being created on the Cosmos Hub

	if createPoolEventPresent && w.Name == "cosmos-hub" {

		chain, err := w.cns.Chain(w.Name)

		if err != nil {
			w.l.Errorw("can't find chain cosmos-hub", "error", err)
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
	if ibcTransferEventPresent {

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
	if ibcSuccessEventPresent {
		if ibcReciveEventPresent {
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

	if ibcTimeoutEventPresent {
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

type denomInfo struct {
	displayName string
	denom       string
	baseDenom   string
	ticker      string
	verified    bool
}

func (d denomInfo) isPoolCoin() bool {
	if len(d.denom) < 4 {
		return false
	}
	return d.denom[:4] == "pool"
}

func (d denomInfo) isIBCToken() bool {
	if len(d.denom) < 4 {
		return false
	}
	return d.denom[:4] == "ibc/"
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
	cosmoshub, err := w.cns.Chain("cosmos-hub")

	if err != nil {
		return d, err
	}

	var denomInfos []denomInfo

	for _, coin := range coins {
		denom := denomInfo{
			denom: coin.Denom,
		}

		if denom.isIBCToken() {

			verifiedTrace := VerifyTraceResponse{}
			w.l.Debugw("querying verified trace for coin", "coin", denom.denom)

			endpoint := fmt.Sprintf("%s/chain/%s/denom/verify_trace/%s", "http://api-server:8000", "cosmos-hub", denom.denom[4:])

			resp, err := http.Get(endpoint)

			if err != nil {
				return d, err
			}

			if resp.StatusCode == 200 {
				resp, err = http.Get(endpoint)

				if err != nil {
					return d, err
				}
			}

			dc := json.NewDecoder(resp.Body)

			err = dc.Decode(&verifiedTrace)

			if err != nil {
				return d, err
			}

			b, err := json.Marshal(verifiedTrace)

			if err != nil {
				return d, err
			}

			w.l.Debugw("got trace", "trace", string(b))

			if !verifiedTrace.VerifyTrace.Verified {
				return d, fmt.Errorf("not a verified denom")
			}
			if l := len(verifiedTrace.VerifyTrace.Trace); l != 1 {
				return d, fmt.Errorf("trace too long, expected 1, got %d", l)
			}

			denom.baseDenom = verifiedTrace.VerifyTrace.BaseDenom

			sourceChainName := verifiedTrace.VerifyTrace.Trace[0].CounterpartyName
			w.l.Debugw("checking base denom in chain", "denom", denom.baseDenom, "chain", sourceChainName)

			sourceChain, err := w.cns.Chain(sourceChainName)

			if err != nil {
				return d, err
			}

			found := false

			for _, dd := range sourceChain.Denoms {
				if dd.Name == denom.baseDenom {

					if !dd.Verified {
						return d, fmt.Errorf("denom not verified in source chain")
					}
					denom.displayName = dd.DisplayName

					if dd.Ticker == "" {
						denom.ticker = dd.DisplayName
					} else {
						denom.ticker = dd.Ticker
					}

					denom.verified = dd.Verified

					found = true
				}
			}

			if !found {
				return d, fmt.Errorf("denom not found")
			}
		} else {
			// check if token exists & is verified on cosmos hub
			denom.baseDenom = coin.Denom

			for _, dd := range cosmoshub.Denoms {
				if dd.Name == denom.baseDenom {

					if !dd.Verified {
						return d, fmt.Errorf("denom not verified in source chain")
					}
					denom.displayName = dd.DisplayName
					if dd.Ticker == "" {
						denom.ticker = dd.DisplayName
					} else {
						denom.ticker = dd.Ticker
					}
					denom.verified = dd.Verified

				}
			}

		}

		w.l.Debugw("adding verified denom", "denom", denom.displayName)

		denomInfos = append(denomInfos, denom)

	}

	if denomInfos[0].isPoolCoin() || denomInfos[0].isPoolCoin() {
		d.DisplayName = fmt.Sprintf("GDEX %s LP", poolCoinDenom[0])
		d.Ticker = fmt.Sprintf("G-%s", poolCoinDenom[0])
		d.Verified = true
	} else {
		d.DisplayName = fmt.Sprintf("GDEX %s/%s LP", denomInfos[0].displayName, denomInfos[1].displayName)
		d.Ticker = fmt.Sprintf("G-%s-%s", denomInfos[0].ticker, denomInfos[1].ticker)
		d.Verified = true
	}

	w.l.Debugw("verified lp denom", "displayname", d.DisplayName, "ticker", d.Ticker)

	return d, nil
}
