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
	q, err := db.Queryx("SELECT  y.x->'display_name',y.x->'fetch_price' FROM cns.chains jt, LATERAL (SELECT json_array_elements(jt.denoms) x) y")
	if err != nil {
		return nil, err
	}
	for q.Next() {
		var display_name string
		var fetch_price bool
		err := q.Scan(&display_name, &fetch_price)
		if err != nil {
			return nil, err
		}
		if fetch_price == true {
			display_name = strings.TrimRight(display_name, "\"")
			display_name = strings.TrimLeft(display_name, "\"")
			if display_name[0:1] == "U" {
				display_name = display_name[1:]
			}
			Whitelists = append(Whitelists, display_name)
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
