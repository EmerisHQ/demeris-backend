package tmwsproxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

type jsonRpcMsg struct {
	Version string      `json:"jsonrpc"`
	Id      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

var (
	jok                = []byte(`{   "jsonrpc": "2.0",   "id": 1,   "result": {} }`)
	ErrOtherSideAbsent = fmt.Errorf("other side absent")
)

type conn struct {
	ws *websocket.Conn
	m  *sync.Mutex
}

func (c *conn) setWs(ws *websocket.Conn) {
	c.m.Lock()
	defer c.m.Unlock()
	c.ws = ws
}

type proxy struct {
	conn conn
	u    websocket.Upgrader
	l    *zap.SugaredLogger
}

func NewProxy(logger *zap.SugaredLogger) *proxy {
	return &proxy{
		u: websocket.Upgrader{
			HandshakeTimeout: 1 * time.Second,
		},
		l: logger,
		conn: conn{
			m: &sync.Mutex{},
		},
	}
}

func (p *proxy) SendMessage(msg []byte) error {
	if p.conn.ws == nil {
		return ErrOtherSideAbsent
	}

	return p.conn.ws.WriteMessage(websocket.TextMessage, msg)
}

func (p *proxy) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := p.u.Upgrade(w, r, nil)
	if err != nil {
		p.l.Errorw("connection upgrade error", "error", err)
		return
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		p.l.Errorw("message reading error", "error", err)
		return
	}

	p.l.Debugw("websocket proxy in message", "msg", string(message))
	var msg jsonRpcMsg

	if err := json.Unmarshal(message, &msg); err != nil {
		p.l.Errorw("unmarshaling error", "error", err)
		return
	}

	if msg.Method == "subscribe" {
		p.l.Debug("received subscribe request")
		err = c.WriteMessage(websocket.TextMessage, jok)
		if err != nil {
			p.l.Errorw("subscribe response writing error", "error", err)
			_ = c.Close()
			return
		}

		if p.conn.ws == nil {
			p.conn.setWs(c)
		}
	}

}
