package database

import (
	"fmt"
	"net"

	"github.com/allinbits/navigator-backend/config"
	"github.com/jackc/pgx"
	"golang.org/x/crypto/ssh"
)

type Database struct {
	pool      *pgx.ConnPool
	sshClient *ssh.Client
}

var DB *Database

// Initializes a connection pool with an ssh connection
func Init(c *config.Config) error {
	// TODO: toggle ssh with UseSsh var in config
	if sshcon, err := NewSshConnection(c); err == nil {
		fmt.Println("Connected to server")
		connPoolConfig := pgx.ConnPoolConfig{
			ConnConfig: pgx.ConnConfig{
				Host: "localhost",
				Port: uint16(26257),
				User: "root",
				Dial: func(network, addr string) (net.Conn, error) {
					return sshcon.Dial(network, addr)
				},
			},
			MaxConnections: 20,
		}
		pool, err := pgx.NewConnPool(connPoolConfig)
		if err != nil {
			return err
		}
		fmt.Printf("Connected to the db\n")
		db := &Database{
			pool:      pool,
			sshClient: sshcon,
		}

		DB = db

		return nil
	} else {
		panic(err)
	}
}

// Closes the connections
func Close() {
	DB.pool.Close()
	DB.sshClient.Close()
}

// Queries the DB
func Q(sql string, args ...interface{}) (*pgx.Rows, error) {
	return DB.pool.Query(sql, args...)
}
