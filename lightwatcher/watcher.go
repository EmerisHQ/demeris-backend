package lightwatcher

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

const ackSuccess = "AQ==" // Packet ack value is true when ibc is success and contains error message in all other cases
const nonZeroCodeErrFmt = "non-zero code on chain %s: %s"

const (
	EventsTx       = "tm.event='Tx'"
	EventsBlock    = "tm.event='NewBlock'"
	defaultRPCPort = 26657
)

type DataHandler func(watcher *Watcher, event coretypes.ResultEvent)

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

type Watcher struct {
	Name         string
	DataChannel  chan coretypes.ResultEvent
	ErrorChannel chan error

	client           *client.WSClient
	l                *zap.SugaredLogger
	runContext       context.Context
	subs             []string
	stopReadChannel  chan struct{}
	stopErrorChannel chan struct{}
}

func NewWatcher(
	endpoint, chainName string,
	logger *zap.SugaredLogger,
	subscriptions []string,
) (*Watcher, error) {
	ws, err := client.NewWS(
		endpoint,
		"/websocket",
		client.ReadWait(30*time.Second),
	)

	if err != nil {
		return nil, err
	}

	ws.SetLogger(zapLogger{
		z:         logger,
		chainName: chainName,
	})

	if err := ws.OnStart(); err != nil {
		return nil, err
	}

	w := &Watcher{
		client:           ws,
		l:                logger,
		Name:             chainName,
		subs:             subscriptions,
		stopReadChannel:  make(chan struct{}),
		DataChannel:      make(chan coretypes.ResultEvent),
		stopErrorChannel: make(chan struct{}),
		ErrorChannel:     make(chan error),
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

					w.l.Panicw("error on tendermint websocket", "error", data.Error)
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
				if data.Query == "" {
					w.l.Infow(
						"got data from tendermint on empty query",
						"data",
						data.Data,
						"events",
						data.Events,
					)
					continue
				}
				w.l.Infow(
					"got data from tendermint websocket",
					"subscription",
					data.Query,
				)
			}
		}

	}
}
