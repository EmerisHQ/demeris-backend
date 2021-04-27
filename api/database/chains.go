package database

import "github.com/allinbits/demeris-backend/models"

func (d *Database) Chain(name string) (models.Chain, error) {
	var c models.Chain

	n, err := d.dbi.DB.PrepareNamed("select * from cns.chains where chain_name=:name limit 1")
	if err != nil {
		return models.Chain{}, err
	}

	return c, n.Get(&c, map[string]interface{}{
		"name": name,
	})
}

func (d *Database) Chains() ([]models.Chain, error) {
	var c []models.Chain
	return c, d.dbi.Exec("select * from cns.chains", nil, &c)
}
