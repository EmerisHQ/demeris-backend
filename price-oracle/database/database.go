package database

import (
	"strings"

	dbutils "github.com/allinbits/demeris-backend/utils/database"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

type Instance struct {
	d          *dbutils.Instance
	connString string
}

func New(connString string) (*Instance, error) {
	i, err := dbutils.New(connString)
	if err != nil {
		return nil, err
	}

	ii := &Instance{
		d:          i,
		connString: connString,
	}
	_, err = ii.Query("SHOW TABLES FROM oracle")
	if err != nil {
		ii.runMigrations()
	}
	//interim measures
	//_, err = ii.Query("SELECT * FROM oracle.coingecko")
	//if err != nil {
	//	ii.runMigrationsCoingecko()
	//}
	return ii, nil
}

func CnsTokenQuery(db *sqlx.DB) ([]string, error) {
	var Whitelists []string
	q, err := db.Queryx("SELECT  y.x->'ticker',y.x->'fetch_price' FROM cns.chains jt, LATERAL (SELECT json_array_elements(jt.denoms) x) y")
	if err != nil {
		return nil, err
	}
	for q.Next() {
		var ticker string
		var fetch_price bool
		err := q.Scan(&ticker, &fetch_price)
		if err != nil {
			return nil, err
		}
		if fetch_price == true {
			ticker = strings.TrimRight(ticker, "\"")
			ticker = strings.TrimLeft(ticker, "\"")
			Whitelists = append(Whitelists, ticker)
		}
	}
	return Whitelists, nil
}

func CnsPriceIdQuery(db *sqlx.DB) ([]string, error) {
	var Whitelists []string
	q, err := db.Queryx("SELECT  y.x->'price_id',y.x->'fetch_price' FROM cns.chains jt, LATERAL (SELECT json_array_elements(jt.denoms) x) y")
	if err != nil {
		return nil, err
	}
	for q.Next() {
		var price_id string
		var fetch_price bool
		err := q.Scan(&price_id, &fetch_price)
		if err != nil {
			return nil, err
		}
		if fetch_price == true {
			price_id = strings.TrimRight(price_id, "\"")
			price_id = strings.TrimLeft(price_id, "\"")
			Whitelists = append(Whitelists, price_id)
		}
	}
	return Whitelists, nil
}

func (i *Instance) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	q, err := i.d.DB.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (i *Instance) CnstokenQueryHandler() ([]string, error) {
	Whitelists, err := CnsTokenQuery(i.d.DB)
	if err != nil {
		return nil, err
	}
	return Whitelists, nil
}
