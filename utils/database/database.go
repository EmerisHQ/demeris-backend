package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
)

const (
	DriverPGX = "pgx"
	DriverPQ  = "postgres"
)

// New returns an Instance connected to the database pointed by connString.
func New(connString string) (*pgxpool.Pool, error) {
	return pgxpool.Connect(context.Background(), connString)
}

//// NewWithDriver returns an Instance connected to the database pointed by connString with the given driver.
//func NewWithDriver(connString string, driver string) (*Instance, error) {
//	db, err := sqlx.Connect(driver, connString)
//	if err != nil {
//		return nil, err
//	}
//
//	i := &Instance{
//		DB: db,
//	}
//
//	if err := i.DB.Ping(); err != nil {
//		return nil, fmt.Errorf("cannot ping db, %w", err)
//	}
//
//	i.DB.DB.SetMaxOpenConns(25)
//	i.DB.DB.SetMaxIdleConns(25)
//	i.DB.DB.SetConnMaxLifetime(5 * time.Minute)
//
//	return i, nil
//}

//// Exec executes query with the given params.
//// If params is nil, query is assumed to be of the `SELECT` kind, and the resulting data will be written in dest.
//func (i *Instance) Exec(query string, params interface{}, dest interface{}) error {
//	return crdbsqlx.ExecuteTx(context.Background(), i.DB, nil, func(tx *sqlx.Tx) error {
//		if dest != nil {
//			if params != nil {
//				return tx.Select(dest, query, params)
//			}
//
//			return tx.Select(dest, query)
//		}
//
//		res, err := tx.NamedExec(query, params)
//		if err != nil {
//			return fmt.Errorf("transaction named exec error, %w", err)
//		}
//
//		re, err := res.RowsAffected()
//		if err != nil {
//			return fmt.Errorf("transaction named exec error, %w", err)
//		}
//
//		if re == 0 {
//			return fmt.Errorf("affected rows are zero")
//		}
//
//		return nil
//	})
//}
