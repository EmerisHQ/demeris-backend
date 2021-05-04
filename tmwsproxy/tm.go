package tmwsproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/allinbits/demeris-backend/utils/database"

	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"go.uber.org/zap"

	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

const (
	blEvents = "tm.event='NewBlock'"
	txEvent  = "tm.event='Tx'"
)

type TendermintClient struct {
	client      *client.WSClient
	l           *zap.SugaredLogger
	db          *database.Instance
	DataChannel chan json.RawMessage
}

type bankSendEvt struct {
	sender   string
	receiver string
	amount   string
}

func (t *TendermintClient) Subscribe(evtString string) error {
	return t.client.Subscribe(context.Background(), evtString)
}

func (t *TendermintClient) readChannel() {
	for data := range t.client.ResponsesCh {
		e := coretypes.ResultEvent{}
		err := json.Unmarshal(data.Result, &e)
		if err != nil {
			t.l.Errorw("cannot unmarshal data into resultevent", "error", err)
			return
		}

		if !t.shouldRelayEvents(e) {
			continue
		}

		mdata, err := json.Marshal(data)
		if err != nil {
			t.l.Debugw("cannot marshal rpc data", "error", err)
			continue
		}

		go func() {
			t.DataChannel <- mdata
		}()
	}
}

func (t *TendermintClient) queryFeeAddresses() (map[string]struct{}, error) {
	var addresses []string

	if err := t.db.DB.Select(&addresses, "SELECT fee_address FROM cns.chains"); err != nil {
		return nil, err
	}

	var ret = map[string]struct{}{}

	for _, a := range addresses {
		ret[a] = struct{}{}
	}

	return ret, nil
}

// TODO: add logic here
func (t *TendermintClient) shouldRelayEvents(evt coretypes.ResultEvent) bool {
	if evt.Query != txEvent {
		return false
	}

	e := evt.Events

	for k, v := range e {
		t.l.Debugw("event", "key", k, "value", v)
	}

	ibcSenders := map[string]struct{}{}

	// grab ibc transfers senders
	for _, s := range e["ibc_transfer.sender"] {
		ibcSenders[s] = struct{}{}
	}

	if len(ibcSenders) == 0 {
		return false // nothing to relay
	}

	packetData, ok := e["send_packet.packet_data"]
	if !ok { // no IBC stuff here
		return false
	}

	_ = packetData

	// grab fee addresses
	feeAddresses, err := t.queryFeeAddresses()
	if err != nil {
		t.l.Errorw("cannot query fee addresses, not relying events", "error", err)
	}

	// search for bank modules messages index
	var bankEvts []bankSendEvt
	for i, m := range e["message.module"] {
		if m == "bank" {
			// is this transfer initiated by someone who sent out an ibc transfer?
			if _, ok := ibcSenders[e["transfer.sender"][i]]; !ok {
				continue
			}

			// are we the recipient?
			recp := e["transfer.recipient"][i]
			if _, ok := feeAddresses[recp]; !ok {
				continue
			}

			// TODO: filter out empty or non ok amounts
			bankEvts = append(bankEvts, bankSendEvt{
				sender:   e["transfer.sender"][i],
				receiver: e["transfer.recipient"][i],
				amount:   e["transfer.amount"][i],
			})
		}
	}

	// filter out bank events of which

	t.l.Debugw("bank events found", "events", bankEvts)

	return true
}

func NewTendermintClient(endpoint string, logger *zap.SugaredLogger, db *database.Instance) (*TendermintClient, error) {
	// TODO: handle immediate resubscription, see newWSEvents() in tendermint codebase
	wsc, err := client.NewWS(endpoint, "/websocket")
	if err != nil {
		return nil, err
	}

	if err := wsc.Start(); err != nil {
		return nil, err
	}

	tm := &TendermintClient{
		client:      wsc,
		l:           logger,
		db:          db,
		DataChannel: make(chan json.RawMessage),
	}

	if err := tm.Subscribe(blEvents); err != nil {
		return nil, fmt.Errorf("real node subscription error, %w", err)
	}

	if err := tm.Subscribe(txEvent); err != nil {
		return nil, fmt.Errorf("real node subscription error, %w", err)
	}

	go tm.readChannel()

	return tm, nil
}
