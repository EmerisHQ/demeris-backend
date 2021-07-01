package database

import (
	"fmt"
	"strings"

	"github.com/allinbits/demeris-backend/models"
)

func (d *Database) DenomTrace(chain string, hash string) (models.IBCDenomTraceRow, error) {
	hash = strings.ToLower(hash)
	var denomTraces []models.IBCDenomTraceRow

	q := "SELECT * FROM tracelistener.denom_traces WHERE chain_name=? and hash=? and base_denom != '' limit 1;"

	q = d.dbi.DB.Rebind(q)

	if err := d.dbi.DB.Select(&denomTraces, q, chain, hash); err != nil {
		return models.IBCDenomTraceRow{}, err
	}

	if len(denomTraces) == 0 {
		return models.IBCDenomTraceRow{}, fmt.Errorf("query done but returned no result")
	}

	return denomTraces[0], nil
}
