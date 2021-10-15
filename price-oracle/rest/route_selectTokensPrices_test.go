package rest

import (
	"bytes"
	"encoding/json"
	"github.com/allinbits/demeris-backend/price-oracle/types"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestSelectTokensPrices(t *testing.T) {
	router, ctx, w, tDown := setup(t)
	defer tDown()

	ctx.Request = &http.Request{
		Header: make(http.Header),
	}
	ctx.Request.Method = "POST" // or PUT
	ctx.Request.Header.Set("Content-Type", "application/json")

	fiats := types.SelectToken{
		Tokens: []string{"ATOMUSDT", "LUNAUSDT"},
	}
	jsonBytes, err := json.Marshal(fiats)
	require.NoError(t, err)
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(jsonBytes))

	_, handler := router.getselectTokensPrices()
	handler(ctx)

	var got struct {
		Data []types.TokenPriceResponse `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &got)
	require.NoError(t, err)

	want := []types.TokenPriceResponse{
		{Price: 10, Symbol: "ATOMUSDT", Supply: 113563929433.0},
		{Price: 10, Symbol: "LUNAUSDT", Supply: 113563929433.0},
	}

	require.Equal(t, want, got.Data)
}
