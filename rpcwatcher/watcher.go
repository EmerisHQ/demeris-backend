package rpcwatcher

import (
	"context"
	"encoding/json"
	"fmt"

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
	Name        string
	client      *client.WSClient
	d           *database.Instance
	l           *zap.SugaredLogger
	store       *store.Store
	DataChannel chan coretypes.ResultEvent
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

func NewWatcher(endpoint string, logger *zap.SugaredLogger, db *database.Instance, s *store.Store, subscriptions []string) (*Watcher, error) {
	// TODO: handle immediate resubscription, see newWSEvents() in tendermint codebase
	ws, err := client.NewWS(endpoint, "/websocket")
	if err != nil {
		return nil, err
	}

	if err := ws.Start(); err != nil {
		return nil, err
	}

	w := &Watcher{
		client:      ws,
		l:           logger,
		store:       s,
		DataChannel: make(chan coretypes.ResultEvent),
	}

	for _, sub := range subscriptions {
		if err := w.client.Subscribe(context.Background(), sub); err != nil {
			return nil, fmt.Errorf("failed to subscribe, %w", err)
		}
	}

	go w.readChannel()

	return w, nil
}

func (w *Watcher) readChannel() {
	for data := range w.client.ResponsesCh {
		e := coretypes.ResultEvent{}
		err := json.Unmarshal(data.Result, &e)
		if err != nil {
			w.l.Errorw("cannot unmarshal data into resultevent", "error", err)
			return
		}

		go func() {
			w.DataChannel <- e
		}()
	}
}

func (w *Watcher) HandleMessage(data coretypes.ResultEvent) {

	txHash, exists := data.Events["tx.hash"]
	_, isIBC := data.Events["ibc_transfer.sender"]
	_, isIBCRecv := data.Events["recv_packet.packet_sequence"]

	// Handle case where a simple non-IBC transfer is being used.
	if key := fmt.Sprintf("%s-%s", w.Name, txHash); exists && !isIBC && !isIBCRecv && w.store.Exists(key) {
		w.store.SetComplete(key)
	}

	// Handle case where an IBC transfer is sent from the origin chain.
	if isIBC {
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

		var c models.ChannelQuery

		q, err := w.d.DB.PrepareNamed("select chain_name, mapping.* from cns.chains c, jsonb_each_text(primary_channel) mapping where chain_name=:chain_name and mapping.channel_name=:channel limit 1")
		if err != nil {
			w.l.Errorf(err.Error())
		}

		q.Select(&c, map[string]interface{}{
			"chain_name": w.Name,
			"channel":    sendPacketSourceChannel,
		})

		w.store.SetInTransit(fmt.Sprintf("%s-%s", w.Name, txHash), c.Counterparty, sendPacketSourceChannel[0], sendPacketSequence[0])
	}

	// Handle case where IBC transfer is received by the receiving chain.
	if isIBCRecv {
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

		key := fmt.Sprintf(w.Name, recvPacketSourceChannel[0], recvPacketSequence[0])

		w.store.SetIbcReceived(key)

	}

}
