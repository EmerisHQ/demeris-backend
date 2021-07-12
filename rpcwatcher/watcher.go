package rpcwatcher

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/utils/database"
	"github.com/allinbits/demeris-backend/utils/store"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
	jsonrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"

	"github.com/tendermint/tendermint/types"
)

var (
	txEvent  = "tm.event='Tx'"
	subscriptions = []string{txEvent}
)

type Watcher struct {
	Name             string
	client           *client.WSClient
	d                *database.Instance
	l                *zap.SugaredLogger
	store            *store.Store
	stopReadChannel  chan struct{}
	DataChannel      chan coretypes.ResultEvent
	stopErrorChannel chan struct{}
	ErrorChannel     chan *jsonrpctypes.RPCError
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

func NewWatcher(endpoint, chainName string, logger *zap.SugaredLogger, db *database.Instance, s *store.Store, subscriptions []string) (*Watcher, error) {

	ws, err := client.NewWS(endpoint, "/websocket")
	if err != nil {
		return nil, err
	}

	if err := ws.OnStart(); err != nil {
		return nil, err
	}

	w := &Watcher{
		d:               db,
		client:          ws,
		l:               logger,
		store:           s,
		Name:            chainName,
		stopReadChannel: make(chan struct{}),
		DataChannel:     make(chan coretypes.ResultEvent),
		stopErrorChannel: make(chan struct{}),
		ErrorChannel:     make(chan *jsonrpctypes.RPCError),
	}

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
					w.l.Debugw("this is error", "error", data.Error)
					go func() {
						w.l.Debugw("error is being written")
						w.ErrorChannel <- data.Error
					}()
					continue
				}

				e := coretypes.ResultEvent{}
				err := tmjson.Unmarshal(data.Result, &e)
				if err != nil {
					w.l.Errorw("cannot unmarshal data into resultevent", "error", err)
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


func (w *Watcher) checkError() {
	for {
		select {
		case <-w.stopErrorChannel:
			return
		default:
			select {
			case err := <-w.ErrorChannel:
				if err != nil {
					resubscribe(w)
					continue
				}
			}
		}
	}
}

func resubscribe(w *Watcher) {
	count := 0
	for {
		time.Sleep(500*time.Millisecond)
		count = count + 1
		w.l.Debugw("this is count", "count", count)
		for _, sub := range subscriptions {
			err := w.client.Subscribe(context.Background(), sub)
			if err != nil {
				if err.Error() == "RPC error -32603 - Internal error: already subscribed"{
					w.l.Debugw("unable to subscribe thisss", "error", err)
					<-w.ErrorChannel
					return
				}
				w.l.Debugw("unable to subscribe", "error", err)
			}
		}
	}
}


func (w *Watcher) handleMessage(data coretypes.ResultEvent) {
	txHashSlice, exists := data.Events["tx.hash"]
	_, isIBC := data.Events["ibc_transfer.sender"]
	_, isIBCRecv := data.Events["recv_packet.packet_sequence"]

	if len(txHashSlice) == 0 {
		return
	}

	txHash := txHashSlice[0]

	key := fmt.Sprintf("%s-%s", w.Name, txHash)

	w.l.Debugw("got message to handle", "chain name", w.Name, "key", key, "is ibc", isIBC, "is ibc recv", isIBCRecv)

	w.l.Debugw("is simple ibc transfer", "is it", exists && !isIBC && !isIBCRecv && w.store.Exists(key))
	// Handle case where a simple non-IBC transfer is being used.
	if exists && !isIBC && !isIBCRecv && w.store.Exists(key){
		eventTx := data.Data.(types.EventDataTx)

		if eventTx.Result.Code == 0 {
			if err := w.store.SetComplete(key); err != nil {
				w.l.Errorw("cannot set complete", "chain name", w.Name, "error", err)
			}
			return
		}

		if err := w.store.SetFailedWithErr(key, eventTx.Result.Log); err != nil {
			w.l.Errorw("cannot set failed with err", "chain name", w.Name, "error", err)
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

		var c []models.ChannelQuery

		q, err := w.d.DB.PrepareNamed("select chain_name, json_data.* from cns.chains, jsonb_each_text(primary_channel) as json_data where chain_name=:chain_name and value=:channel limit 1;")
		if err != nil {
			w.l.Errorw("cannot prepare statement", "error", err)
			return
		}

		if err := q.Select(&c, map[string]interface{}{
			"chain_name": w.Name,
			"channel":    sendPacketSourceChannel[0],
		}); err != nil {
			w.l.Errorw("cannot query chain", "error", err)
			return
		}

		if len(c) == 0 {
			w.l.Errorw("cannot query chain, database query returned 0 rows")
			return
		}

		w.store.SetInTransit(fmt.Sprintf("%s-%s", w.Name, txHash), c[0].Counterparty, sendPacketSourceChannel[0], sendPacketSequence[0])
		return
	}

	// Handle case where IBC transfer is received by the receiving chain.
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
				w.handleMessage(data)
			}
		}

	}
}
