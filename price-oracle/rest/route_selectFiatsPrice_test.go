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

func TestSelectFiatsPrice(t *testing.T) {
	router, ctx, w, tDown := setup(t)
	defer tDown()

	ctx.Request = &http.Request{
		Header: make(http.Header),
	}
	ctx.Request.Method = "POST" // or PUT
	ctx.Request.Header.Set("Content-Type", "application/json")

	fiats := types.SelectFiat{
		Fiats: []string{"USDEUR", "USDKRW"},
	}
	jsonBytes, err := json.Marshal(fiats)
	require.NoError(t, err)
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(jsonBytes))

	_, handler := router.getselectFiatsPrices()
	handler(ctx)

	var got struct {
		Data []types.FiatPriceResponse `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &got)
	require.NoError(t, err)

	want := []types.FiatPriceResponse{
		{Symbol: "USDEUR", Price: 20},
		{Symbol: "USDKRW", Price: 5},
	}

	require.Equal(t, want, got.Data)
}
