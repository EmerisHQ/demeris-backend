package database

import (
	"strings"

	dbutils "github.com/allinbits/demeris-backend/utils/database"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v4/stdlib"
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
			if ticker[0:1] == "U" {
				ticker = ticker[1:]
			}
			if ticker == "OSMO" {
				continue
			}
			if ticker == "REGEN" {
				continue
			}
			Whitelists = append(Whitelists, ticker)
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
