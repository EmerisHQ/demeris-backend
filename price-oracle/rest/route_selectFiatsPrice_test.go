package rest

import "testing"

func TestSelectFiatsPrice(t *testing.T) {
	router, ctx, w, tDown := setup(t)
	defer tDown()

	_, handler := router.getselectFiatsPrices()
	handler(ctx)
	gotFiats := getFiatsFromResponseWriter(t, w)

	t.Logf("Fiats %v\n", gotFiats)

}
