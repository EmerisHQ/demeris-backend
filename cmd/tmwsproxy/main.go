package main

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/allinbits/emeris-utils/database"

	"github.com/allinbits/demeris-backend/tmwsproxy"

	"github.com/allinbits/emeris-utils/logging"
	"github.com/cssivision/reverseproxy"
	"github.com/gorilla/mux"
)

func main() {
	c, err := tmwsproxy.ReadConfig()
	if err != nil {
		panic(err)
	}

	l := logging.New(logging.LoggingConfig{
		Debug: c.Debug,
	})

	db, err := database.New(c.DatabaseConnectionURL)
	if err != nil {
		panic(err)
	}

	lnu, err := url.Parse(c.TendermintNode)
	if err != nil {
		panic(err)
	}

	proxy := reverseproxy.NewReverseProxy(lnu)

	relayerProxy := tmwsproxy.NewProxy(l)

	router := mux.NewRouter()

	router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			l.Infow("new request", "endpoint", r.URL.String(), "time", time.Now(), "method", r.Method)
			handler.ServeHTTP(rw, r)
		})
	})

	router.HandleFunc("/websocket", relayerProxy.WebsocketHandler)

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// redirect everything that comes on this endpoint to the standard tendermint endpoint
		l.Debugw("redirecting request to real tendermint node")
		proxy.ServeHTTP(w, r)
	})

	go func() {
		l.Panicw("http server panic-ed", "error", http.ListenAndServe(c.ListenAddr, router))
	}()

	tm, err := tmwsproxy.NewTendermintClient(c.TendermintNode, l, db)
	if err != nil {
		l.Panicw("real node connection error", "error", err)
	}

	for data := range tm.DataChannel {
		err = relayerProxy.SendMessage(data)
		if err != nil && !errors.Is(err, tmwsproxy.ErrOtherSideAbsent) {
			l.Panicw("cannot send data to relayer", "error", err)
		}
	}
}
