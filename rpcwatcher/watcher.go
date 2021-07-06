package rpcwatcher

import (
	"context"
	"encoding/json"
	"fmt"

	cnsdb "github.com/allinbits/demeris-backend/cns/database"
	// tldb "github.com/allinbits/demeris-backend/api/database"
	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/utils/database"

	"github.com/allinbits/demeris-backend/utils/store"
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
	client          *client.WSClient
	d               *database.Instance
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

func NewWatcher(endpoint, chainName string, logger *zap.SugaredLogger, db *database.Instance, cnsdb *cnsdb.Instance, s *store.Store, subscriptions []string) (*Watcher, error) {

	ws, err := client.NewWS(endpoint, "/websocket")
	if err != nil {
		return nil, err
	}

	if err := ws.Start(); err != nil {
		return nil, err
	}

	w := &Watcher{
		d:               db,
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

	w.l.Debugw("got message to handle", "chain name", w.Name, "key", key, "is ibc", isIBC, "is ibc recv", isIBCRecv,
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

		for _, token := range chain.Denoms {
			if token.Name == poolCoinDenom[0] {
				token.Verified = true

				poolName, ok := data.Events["create_pool.pool_name"]

				if !ok {
					w.l.Errorw("no field create_pool.pool_name in Events", "error", err)
					return
				}

				token.DisplayName = fmt.Sprintf("[AMM] %s LP", poolName[0])

				// todo: use simplified display name for IBC tokens
				break
			}
		}

		err = w.cns.AddChain(chain)

		if err != nil {

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
