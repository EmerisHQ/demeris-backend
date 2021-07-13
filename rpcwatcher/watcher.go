package rpcwatcher

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/utils/database"
	"github.com/allinbits/demeris-backend/utils/store"
	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"github.com/tendermint/tendermint/types"
)

const (
	blEvents   = "tm.event='NewBlock'"
	txEvent    = "tm.event='Tx'"
	ackSuccess = "AQ==" // Packet ack value is true when ibc is success and contains error message in all other cases
)

type Watcher struct {
	Name            string
	client          *client.WSClient
	d               *database.Instance
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

func NewWatcher(endpoint, chainName string, logger *zap.SugaredLogger, db *database.Instance, s *store.Store, subscriptions []string) (*Watcher, error) {

	ws, err := client.NewWS(endpoint, "/websocket")
	if err != nil {
		return nil, err
	}

	if err := ws.Start(); err != nil {
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
				if err := tmjson.Unmarshal(data.Result, &e); err != nil {
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
	_, IBCSenderEventPresent := data.Events["ibc_transfer.sender"]
	_, IBCAckEventPresent := data.Events["fungible_token_packet.acknowledgement"]
	_, IBCReceivePacketEventPresent := data.Events["recv_packet.packet_sequence"]
	_, IBCTimeoutEventPresent := data.Events["timeout.refund_receiver"]

	if len(txHashSlice) == 0 {
		return
	}

	txHash := txHashSlice[0]

	key := fmt.Sprintf("%s-%s", w.Name, txHash)

	w.l.Debugw("got message to handle", "chain name", w.Name, "key", key, "is ibc", IBCSenderEventPresent, "is ibc recv", IBCReceivePacketEventPresent,
		"is ibc ack", IBCAckEventPresent, "is ibc timeout", IBCTimeoutEventPresent)

	w.l.Debugw("is simple ibc transfer"+
		"", "is it", exists && !IBCSenderEventPresent && !IBCReceivePacketEventPresent && w.store.Exists(key))
	// Handle case where a simple non-IBC transfer is being used.
	if exists && !IBCSenderEventPresent && !IBCReceivePacketEventPresent && w.store.Exists(key) {
		eventTx := data.Data.(types.EventDataTx)

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

		counterparty, ok := data.Events["send_packet.packet_dst_channel"]
		if !ok {
			w.l.Errorf("send_packet.packet_dst_channel not found for key", "key", key)
			return
		}

		if err := w.store.SetInTransit(key, counterparty[0], sendPacketSourceChannel[0], sendPacketSequence[0]); err != nil {
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

		if packetAck[0] != ackSuccess {
			if err := w.store.SetIbcFailed(key); err != nil {
				w.l.Errorw("unable to status as failed for key", "key", key, "error", err)
			}
			return
		}

		key := fmt.Sprintf("%s-%s-%s", w.Name, recvPacketSourceChannel[0], recvPacketSequence[0])
		if err := w.store.SetIbcReceived(key); err != nil {
			w.l.Errorw("unable to status as ibc received for key", "key", key, "error", err)
		}
		return
	}

	if IBCTimeoutEventPresent {
		_, ok := data.Events["timeout.refund_receiver"]
		if !ok {
			w.l.Errorw("refund receiver not found for key", "key", key)
			return
		}

		if err := w.store.SetIbcTimeoutUnlock(key); err != nil {
			w.l.Errorw("unable to status as ibc timeout unlock for key", "key", key, "error", err)
		}
		return
	}

	if IBCAckEventPresent {
		_, ok := data.Events["fungible_token_packet.error"]
		if ok {
			if err := w.store.SetIbcAckUnlock(key); err != nil {
				w.l.Errorw("unable to status as ibc ack unlock for key", "key", key, "error", err)
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
