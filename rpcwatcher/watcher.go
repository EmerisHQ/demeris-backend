package rpcwatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/allinbits/demeris-backend/rpcwatcher/database"

	"github.com/allinbits/demeris-backend/utils/store"
	gaia "github.com/cosmos/gaia/v5/app"
	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
	jsonrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	"github.com/tendermint/tendermint/types"
)

const ackSuccess = "AQ==" // Packet ack value is true when ibc is success and contains error message in all other cases
const nonZeroCodeErrFmt = "non-zero code on chain %s: %s"

const (
	EventsTx       = "tm.event='Tx'"
	EventsBlock    = "tm.event='NewBlock'"
	defaultRPCPort = 26657
	grpcPort       = 9090
)

type Watcher struct {
	Name             string
	apiUrl           string
	client           *client.WSClient
	d                *database.Instance
	l                *zap.SugaredLogger
	store            *store.Store
	runContext       context.Context
	endpoint         string
	subs             []string
	stopReadChannel  chan struct{}
	DataChannel      chan coretypes.ResultEvent
	stopErrorChannel chan struct{}
	ErrorChannel     chan *jsonrpctypes.RPCError
}

type WsResponse struct {
	Event coretypes.ResultEvent `json:"result"`
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

type Ack struct {
	Result string `json:"result"`
}

func NewWatcher(endpoint, chainName string, logger *zap.SugaredLogger, apiUrl string, db *database.Instance, s *store.Store, subscriptions []string) (*Watcher, error) {

	ws, err := client.NewWS(endpoint, "/websocket")
	if err != nil {
		return nil, err
	}

	if err := ws.OnStart(); err != nil {
		return nil, err
	}

	w := &Watcher{
		apiUrl:           apiUrl,
		d:                db,
		client:           ws,
		l:                logger,
		store:            s,
		Name:             chainName,
		endpoint:         endpoint,
		subs:             subscriptions,
		stopReadChannel:  make(chan struct{}),
		DataChannel:      make(chan coretypes.ResultEvent),
		stopErrorChannel: make(chan struct{}),
		ErrorChannel:     make(chan *jsonrpctypes.RPCError),
	}

	w.l.Debugw("creating rpcwatcher with config", "apiurl", apiUrl)

	for _, sub := range subscriptions {
		if err := w.client.Subscribe(context.Background(), sub); err != nil {
			return nil, fmt.Errorf("failed to subscribe, %w", err)
		}
	}

	go w.readChannel()

	go w.checkError()
	return w, nil
}

func Start(watcher *Watcher, ctx context.Context) {
	watcher.runContext = ctx
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
					go func() {
						w.l.Debugw("writing error to error channel", "error", data.Error)
						w.ErrorChannel <- data.Error
					}()

					// if we get any kind of error from tendermint, exit: the reconnection routine will take care of
					// getting us up to speed again
					return
				}

				e := coretypes.ResultEvent{}
				if err := tmjson.Unmarshal(data.Result, &e); err != nil {
					w.l.Errorw("cannot unmarshal data into resultevent", "error", err, "chain", w.Name)
					continue
				}

				go func() {
					w.DataChannel <- e
				}()
			}
		}
	}
}

func (w *Watcher) checkError() {
	for {
		select {
		case <-w.stopErrorChannel:
			return
		default:
			select {
			case err := <-w.ErrorChannel:
				if err != nil {
					storeErr := w.store.SetWithExpiry(w.Name, "false", 0)
					if err != nil {
						w.l.Errorw("unable to set chain name to false", "store error", storeErr,
							"error", err)
					}
					resubscribe(w)
					return
				}
			}
		}
	}
}

func resubscribe(w *Watcher) {
	count := 0
	for {
		err := w.store.SetWithExpiry(w.Name, "resubscribing", 0)
		if err != nil {
			w.l.Errorw("unable to set chain name with status resubscribing", "error", err)
		}

		time.Sleep(500 * time.Millisecond)
		count = count + 1
		w.l.Debugw("this is count", "count", count)

		ww, err := NewWatcher(w.endpoint, w.Name, w.l, w.apiUrl, w.d, w.store, w.subs)
		if err != nil {
			w.l.Errorw("cannot resubscribe to chain", "name", w.Name, "endpoint", w.endpoint, "error", err)
			continue
		}

		ww.runContext = w.runContext
		w = ww

		Start(w, w.runContext)
		err = w.store.SetWithExpiry(w.Name, "true", 0)
		if err != nil {
			w.l.Errorw("unable to set chain name as true", "error", err)
		}

		w.l.Infow("successfully reconnected", "name", w.Name, "endpoint", w.endpoint)
		return
	}
}

func (w *Watcher) handleMessage(data coretypes.ResultEvent) {
	txHashSlice, exists := data.Events["tx.hash"]
	_, createPoolEventPresent := data.Events["create_pool.pool_name"]
	_, IBCSenderEventPresent := data.Events["ibc_transfer.sender"]
	_, IBCAckEventPresent := data.Events["fungible_token_packet.acknowledgement"]
	_, IBCReceivePacketEventPresent := data.Events["recv_packet.packet_sequence"]
	_, IBCTimeoutEventPresent := data.Events["timeout.refund_receiver"]

	if len(txHashSlice) == 0 {
		return
	}

	txHash := txHashSlice[0]
	chainName := w.Name
	eventTx := data.Data.(types.EventDataTx)
	height := eventTx.Height
	key := fmt.Sprintf("%s-%s", chainName, txHash)

	w.l.Debugw("got message to handle", "chain name", chainName, "key", key, "is create lp", createPoolEventPresent, "is ibc", IBCSenderEventPresent, "is ibc recv", IBCReceivePacketEventPresent,
		"is ibc ack", IBCAckEventPresent, "is ibc timeout", IBCTimeoutEventPresent)

	w.l.Debugw("is simple ibc transfer"+
		"", "is it", exists && !createPoolEventPresent && !IBCSenderEventPresent && !IBCReceivePacketEventPresent && w.store.Exists(key))

	if eventTx.Result.Code != 0 {
		logStr := fmt.Sprintf(nonZeroCodeErrFmt, chainName, eventTx.Result.Log)

		w.l.Debugw("transaction error", "chainName", chainName, "txHash", txHash, "log", eventTx.Result.Log)

		if err := w.store.SetFailedWithErr(key, logStr, height); err != nil {
			w.l.Errorw("cannot set failed with err", "chain name", chainName, "error", err,
				"txHash", txHash, "code", eventTx.Result.Code)
		}

		return
	}

	// Handle case where a simple non-IBC transfer is being used.
	if exists && !createPoolEventPresent && !IBCSenderEventPresent && !IBCReceivePacketEventPresent &&
		!IBCAckEventPresent && !IBCTimeoutEventPresent && w.store.Exists(key) {
		if err := w.store.SetComplete(key, height); err != nil {
			w.l.Errorw("cannot set complete", "chain name", chainName, "error", err)
		}
		return
	}

	// Handle case where an LP is being created on the Cosmos Hub
	if createPoolEventPresent && chainName == "cosmos-hub" {
		defer func() {
			if err := w.store.SetComplete(key, height); err != nil {
				w.l.Errorw("cannot set complete", "chain name", chainName, "error", err)
			}
		}()

		w.l.Debugw("is create lp", "is it", createPoolEventPresent)

		chain, err := w.d.Chain(chainName)

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

		err = w.d.UpdateDenoms(chain)

		if err != nil {
			w.l.Errorw("failed to update chain", "error", err)
			return
		}

		return
	}

	// Handle case where an IBC transfer is sent from the origin chain.
	if IBCSenderEventPresent {

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

		c, err := w.d.GetCounterParty(chainName, sendPacketSourceChannel[0])
		if err != nil {
			w.l.Errorw("unable to fetch counterparty chain from db", "error", err)
			return
		}

		if err := w.store.SetInTransit(key, c[0].Counterparty, sendPacketSourceChannel[0], sendPacketSequence[0], txHash, chainName, height); err != nil {
			w.l.Errorw("unable to set status as in transit for key", "key", key, "error", err)
		}
		return
	}

	// Handle case where IBC transfer is received by the receiving chain.
	if IBCReceivePacketEventPresent {
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

		packetAck, ok := data.Events["write_acknowledgement.packet_ack"]

		if !ok {
			w.l.Errorf("packet ack not found")
			return
		}

		key := fmt.Sprintf("%s-%s-%s", chainName, recvPacketSourceChannel[0], recvPacketSequence[0])
		var ack Ack
		if err := json.Unmarshal([]byte(packetAck[0]), &ack); err != nil {
			w.l.Errorw("unable to unmarshal packetAck", "err", err)
			return
		}

		if ack.Result != ackSuccess {
			if err := w.store.SetIbcFailed(key, txHash, chainName, height); err != nil {
				w.l.Errorw("unable to set status as failed for key", "key", key, "error", err)
			}
			return
		}

		if err := w.store.SetIbcReceived(key, txHash, chainName, height); err != nil {
			w.l.Errorw("unable to set status as ibc received for key", "key", key, "error", err)
		}
		return
	}

	if IBCTimeoutEventPresent {
		timeoutPacketSourceChannel, ok := data.Events["timeout_packet.packet_src_channel"]

		if !ok {
			w.l.Errorf("timeout_packet.packet_src_channel not found")
			return
		}

		timeoutPacketSequence, ok := data.Events["timeout_packet.packet_sequence"]

		if !ok {
			w.l.Errorf("timeout_packet.packet_sequence not found")
			return
		}

		c, err := w.d.GetCounterParty(chainName, timeoutPacketSourceChannel[0])
		if err != nil {
			w.l.Errorw("unable to fetch counterparty chain from db", "error", err)
			return
		}

		key := fmt.Sprintf("%s-%s-%s", c[0].Counterparty, timeoutPacketSourceChannel[0], timeoutPacketSequence[0])
		if err := w.store.SetIbcTimeoutUnlock(key, txHash, chainName, height); err != nil {
			w.l.Errorw("unable to set status as ibc timeout unlock for key", "key", key, "error", err)
		}
		return
	}

	if IBCAckEventPresent {
		ackPacketSourceChannel, ok := data.Events["acknowledge_packet.packet_src_channel"]

		if !ok {
			w.l.Errorf("acknowledge_packet.packet_src_channel not found")
			return
		}

		ackPacketSequence, ok := data.Events["acknowledge_packet.packet_sequence"]

		if !ok {
			w.l.Errorf("acknowledge_packet.packet_sequence not found")
			return
		}

		c, err := w.d.GetCounterParty(chainName, ackPacketSourceChannel[0])
		if err != nil {
			w.l.Errorw("unable to fetch counterparty chain from db", "error", err)
			return
		}

		key := fmt.Sprintf("%s-%s-%s", c[0].Counterparty, ackPacketSourceChannel[0], ackPacketSequence[0])
		_, ok = data.Events["fungible_token_packet.error"]
		if ok {
			if err := w.store.SetIbcAckUnlock(key, txHash, chainName, height); err != nil {
				w.l.Errorw("unable to set status as ibc ack unlock for key", "key", key, "error", err)
			}
			return
		}

	}

}

func (w *Watcher) startChain(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.stopReadChannel <- struct{}{}
			w.stopErrorChannel <- struct{}{}
			w.l.Infof("watcher %s has been canceled", w.Name)
			return
		default:
			select {
			case data := <-w.DataChannel:
				switch data.Query {
				case EventsTx:
					w.handleMessage(data)
				case EventsBlock:
					w.handleBlock(data.Data)
				}
			}
		}

	}
}

func (w *Watcher) handleBlock(data types.TMEventData) {
	w.l.Debugw("called handleBlock")
	realData, ok := data.(types.EventDataNewBlock)
	if !ok {
		panic("rpc returned data which is not of expected type")
	}

	time.Sleep(750 * time.Millisecond) // to handle the time gap between block production and event broadcast
	newHeight := realData.Block.Header.Height

	u := fmt.Sprintf("http://%s:%d", w.Name, defaultRPCPort)

	ru, err := url.Parse(u)
	if err != nil {
		w.l.Errorw("cannot parse url", "url_string", u, "error", err)
		return
	}

	vals := url.Values{}
	vals.Set("height", strconv.FormatInt(newHeight, 10))

	w.l.Debugw("asking for block", "height", newHeight)

	ru.Path = "block_results"
	ru.RawQuery = vals.Encode()

	res := bytes.Buffer{}

	resp, err := http.Get(ru.String())
	if err != nil {
		w.l.Errorw("cannot query node for block data", "error", err, "height", newHeight)
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.l.Errorw("endpoint returned non-200 code", "code", resp.StatusCode, "height", newHeight)
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	read, err := res.ReadFrom(resp.Body)
	if err != nil {
		w.l.Errorw("cannot read block data resp body into buffer", "height", newHeight, "error", err)
		return
	}

	if read == 0 {
		w.l.Errorw("read zero bytes from response body", "height", newHeight)
	}

	bs := store.NewBlocks(w.store)
	err = bs.Add(res.Bytes(), newHeight)
	if err != nil {
		w.l.Errorw("cannot set block to cache", "error", err, "height", newHeight)
		return
	}

	grpcConn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", w.Name, grpcPort),
		grpc.WithInsecure(),
	)

	liquidityQuery := liquiditytypes.NewQueryClient(grpcConn)
	poolsRes, err := liquidityQuery.LiquidityPools(context.Background(), &liquiditytypes.QueryLiquidityPoolsRequest{})
	if err != nil {
		w.l.Errorw("cannot get liquidity pools in blocks", "error", err, "height", newHeight)
	}

	cdc, _ := gaia.MakeCodecs()
	bz, err := cdc.MarshalJSON(poolsRes)
	if err != nil {
		w.l.Errorw("cannot marshal liquidity pools", "error", err, "height", newHeight)
	}

	err = w.store.SetWithExpiry("pools", string(bz), 0)
	if err != nil {
		w.l.Errorw("cannot set liquidity pools", "error", err, "height", newHeight)
	}

	paramsRes, err := liquidityQuery.Params(context.Background(), &liquiditytypes.QueryParamsRequest{})
	if err != nil {
		w.l.Errorw("cannot get liquidity params", "error", err, "height", newHeight)
	}

	bz, err = cdc.MarshalJSON(paramsRes)
	if err != nil {
		w.l.Errorw("cannot unmarshal liquidity params", "error", err, "height", newHeight)
	}

	err = w.store.SetWithExpiry("params", string(bz), 0)
	if err != nil {
		w.l.Errorw("cannot set liquidity params", "error", err, "height", newHeight)
	}

	return
}
