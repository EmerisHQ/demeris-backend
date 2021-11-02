package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/allinbits/demeris-backend/price-oracle/types"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestRest(t *testing.T) {
	router, _, _, tDown := setup(t)
	defer tDown()

	s := NewServer(router.s.ri, router.s.l, router.s.d, router.s.c)
	ch := make(chan struct{})
	go func() {
		close(ch)
		if err := s.Serve(router.s.c.ListenAddr); err != nil {
			require.NoError(t, err)
		}
	}()
	<-ch // Wait for the goroutine to start. Still hack!!
	resp, err := http.Get(fmt.Sprintf("http://%s%s", router.s.c.ListenAddr, getAllPriceRoute))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var got struct {
		Data types.AllPriceResponse `json:"data"`
	}
	err = json.Unmarshal(body, &got)
	require.NoError(t, err)
	wantData := types.AllPriceResponse{
		Fiats: []types.FiatPriceResponse{
			{Symbol: "USDCHF", Price: 10},
			{Symbol: "USDEUR", Price: 20},
			{Symbol: "USDKRW", Price: 5},
		},
		Tokens: []types.TokenPriceResponse{
			{Price: 10, Symbol: "ATOMUSDT", Supply: 113563929433.0},
			{Price: 10, Symbol: "LUNAUSDT", Supply: 113563929433.0},
		},
	}
	require.Equal(t, got.Data, wantData)

	var testSetToken = map[string]struct {
		Tokens  types.SelectToken
		Status  int
		Message string
	}{
		"Token: Not whitelisted": {
			types.SelectToken{Tokens: []string{"DOTUSDT"}},
			http.StatusForbidden,
			"Not whitelisting asset",
		},
		"Token: No value": {
			types.SelectToken{Tokens: []string{}},
			http.StatusForbidden,
			"Not allow 0 asset",
		},
		"Token: Nil value": {
			types.SelectToken{Tokens: nil},
			http.StatusForbidden,
			"Not allow nil asset",
		},
		"Token: Exceeds limit": {
			types.SelectToken{Tokens: []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"}},
			http.StatusForbidden,
			"Not allow More than 10 asset",
		},
	}

	for tName, expected := range testSetToken {
		t.Run(tName, func(t *testing.T) {
			jsonBytes, err := json.Marshal(expected.Tokens)
			require.NoError(t, err)

			url := fmt.Sprintf("http://%s%s", router.s.c.ListenAddr, getselectTokensPricesRoute)
			resp, err = http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
			require.NoError(t, err)

			body, err = ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			var gotPost struct {
				Data    []types.TokenPriceResponse `json:"data"`
				Status  int                        `json:"status"`
				Message string                     `json:"message"`
			}

			err = json.Unmarshal(body, &gotPost)
			require.NoError(t, err)
			require.Equal(t, expected.Status, gotPost.Status)
			require.Equal(t, expected.Message, gotPost.Message)
		})
	}

	var testSetFiat = map[string]struct {
		Fiat    types.SelectFiat
		Status  int
		Message string
	}{
		"Fiat: Not whitelisted": {
			types.SelectFiat{Fiats: []string{"USDBDT"}},
			http.StatusForbidden,
			"Not whitelisting asset",
		},
		"Fiat: No value": {
			types.SelectFiat{Fiats: []string{}},
			http.StatusForbidden,
			"Not allow 0 asset",
		},
		"Fiat: Nil value": {
			types.SelectFiat{Fiats: nil},
			http.StatusForbidden,
			"Not allow nil asset",
		},
		"Fiat: Exceeds limit": {
			types.SelectFiat{Fiats: []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"}},
			http.StatusForbidden,
			"Not allow More than 10 asset",
		},
	}

	for tName, expected := range testSetFiat {
		t.Run(tName, func(t *testing.T) {
			jsonBytes, err := json.Marshal(expected.Fiat)
			require.NoError(t, err)

			url := fmt.Sprintf("http://%s%s", router.s.c.ListenAddr, getselectFiatsPricesRoute)
			resp, err = http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
			require.NoError(t, err)

			body, err = ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			var gotPost struct {
				Data    []types.FiatPriceResponse `json:"data"`
				Status  int                       `json:"status"`
				Message string                    `json:"message"`
			}

			err = json.Unmarshal(body, &gotPost)
			require.NoError(t, err)
			require.Equal(t, expected.Status, gotPost.Status)
			require.Equal(t, expected.Message, gotPost.Message)
		})
	}
}
