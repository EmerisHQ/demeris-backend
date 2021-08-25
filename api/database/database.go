package database

import (
	"github.com/allinbits/demeris-backend/api/config"
	"github.com/allinbits/demeris-backend/utils/database"
)

type Database struct {
	dbi           *database.Instance
	connectionURL string
}

// Init initializes a connection to the database.
func Init(c *config.Config) (*Database, error) {
	i, err := database.NewWithDriver(c.DatabaseConnectionURL, database.DriverPQ)
	if err != nil {
		return nil, err
	}

	return &Database{
		dbi:           i,
		connectionURL: c.DatabaseConnectionURL,
	}, nil
}

// Close closes the connections to the database.
func (d *Database) Close() error {
	return d.dbi.Close()
}

// Q queries the DB.
func (d *Database) Q(sql string, dest interface{}, args ...interface{}) error {
	return d.dbi.Exec(sql, args, dest)
}
