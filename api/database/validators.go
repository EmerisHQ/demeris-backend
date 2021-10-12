package database

import "github.com/allinbits/demeris-backend/models"

func (d *Database) GetValidators(chain string) ([]models.ValidatorRow, error) {
	var validators []models.ValidatorRow

	q := `
	SELECT *
	FROM tracelistener.validators 
	WHERE chain_name=?;
	`

	q = d.dbi.DB.Rebind(q)

	return validators, d.dbi.DB.Select(&validators, q, chain)
}
