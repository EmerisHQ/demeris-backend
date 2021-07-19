package rpcwatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/utils/database"
	"github.com/allinbits/demeris-backend/utils/store"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
	jsonrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	"github.com/tendermint/tendermint/types"
)

const ackSuccess = "AQ==" // Packet ack value is true when ibc is success and contains error message in all other cases

type Watcher struct {
	Name             string
	runContext       context.Context
	endpoint         string
	subs             []string
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

type Ack struct {
	Result string `json:"result"`
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
		d:                db,
		endpoint:         endpoint,
		subs:             subscriptions,
		client:           ws,
		l:                logger,
		store:            s,
		Name:             chainName,
		stopReadChannel:  make(chan struct{}),
		DataChannel:      make(chan coretypes.ResultEvent),
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
		time.Sleep(500 * time.Millisecond)
		count = count + 1
		w.l.Debugw("this is count", "count", count)

		ww, err := NewWatcher(w.endpoint, w.Name, w.l, w.d, w.store, w.subs)
		if err != nil {
			w.l.Errorw("cannot resubscribe to chain", "name", w.Name, "endpoint", w.endpoint, "error", err)
			continue
		}

		ww.runContext = w.runContext
		w = ww

		Start(w, w.runContext)

		w.l.Infow("successfully reconnected", "name", w.Name, "endpoint", w.endpoint)
		return
	}
}

func (w *Watcher) handleMessage(data coretypes.ResultEvent) {
	txHashSlice, exists := data.Events["tx.hash"]
	_, IBCSenderEventPresent := data.Events["ibc_transfer.sender"]
	_, IBCAckEventPresent := data.Events["fungible_token_packet.acknowledgement"]
	_, IBCReceivePacketEventPresent := data.Events["recv_packet.packet_sequence"]
	_, IBCTimeoutEventPresent := data.Events["timeout.refund_receiver"]

	if len(txHashSlice) == 0 {
		return
	}

	txHash := txHashSlice[0]
	eventTx := data.Data.(types.EventDataTx)

	key := fmt.Sprintf("%s-%s", w.Name, txHash)

	w.l.Debugw("got message to handle", "chain name", w.Name, "key", key, "is ibc", IBCSenderEventPresent, "is ibc recv", IBCReceivePacketEventPresent,
		"is ibc ack", IBCAckEventPresent, "is ibc timeout", IBCTimeoutEventPresent)

	w.l.Debugw("is simple ibc transfer"+
		"", "is it", exists && !IBCSenderEventPresent && !IBCReceivePacketEventPresent && w.store.Exists(key))
	// Handle case where a simple non-IBC transfer is being used.
	if exists && !IBCSenderEventPresent && !IBCReceivePacketEventPresent &&
		!IBCAckEventPresent && !IBCTimeoutEventPresent && w.store.Exists(key) {

		if eventTx.Result.Code == 0 {
			if err := w.store.SetComplete(key); err != nil {
				w.l.Errorw("cannot set complete", "chain name", w.Name, "error", err)
			}
			return
		}

		if err := w.store.SetFailedWithErr(key, eventTx.Result.Log); err != nil {
			w.l.Errorw("cannot set failed with err", "chain name", w.Name, "error", err,
				"txHash", txHash, "code", eventTx.Result.Code)
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

		c, err := w.GetCounterParty(sendPacketSourceChannel[0])
		if err != nil {
			w.l.Errorw("unable to fetch counterparty chain from db", err)
			return
		}

		if err := w.store.SetInTransit(key, c[0].Counterparty, sendPacketSourceChannel[0], sendPacketSequence[0], eventTx.Height); err != nil {
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

		key := fmt.Sprintf("%s-%s-%s", w.Name, recvPacketSourceChannel[0], recvPacketSequence[0])
		var ack Ack
		if err := json.Unmarshal([]byte(packetAck[0]), &ack); err != nil {
			w.l.Errorw("unable to unmarshal packetAck", "err", err)
			return
		}

		if ack.Result != ackSuccess {
			if err := w.store.SetIbcFailed(key, eventTx.Height); err != nil {
				w.l.Errorw("unable to set status as failed for key", "key", key, "error", err)
			}
			return
		}

		if err := w.store.SetIbcReceived(key, eventTx.Height); err != nil {
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

		c, err := w.GetCounterParty(timeoutPacketSourceChannel[0])
		if err != nil {
			w.l.Errorw("unable to fetch counterparty chain from db", err)
			return
		}

		key := fmt.Sprintf("%s-%s-%s", c[0].Counterparty, timeoutPacketSourceChannel[0], timeoutPacketSequence[0])
		if err := w.store.SetIbcTimeoutUnlock(key, eventTx.Height); err != nil {
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

		c, err := w.GetCounterParty(ackPacketSourceChannel[0])
		if err != nil {
			w.l.Errorw("unable to fetch counterparty chain from db", err)
			return
		}

		key := fmt.Sprintf("%s-%s-%s", c[0].Counterparty, ackPacketSourceChannel[0], ackPacketSequence[0])
		_, ok = data.Events["fungible_token_packet.error"]
		if ok {
			if err := w.store.SetIbcAckUnlock(key, eventTx.Height); err != nil {
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
				w.handleMessage(data)
			}
		}

	}
}
