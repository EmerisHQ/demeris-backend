package database

import "github.com/allinbits/navigator-utils/database"

const createDatabase = `
CREATE DATABASE IF NOT EXISTS cns;
`

const createTableChains = `
CREATE TABLE IF NOT EXISTS cns.chains (
	id serial not null,
	client_id string not null,
	chain_name string not null,
	chain_id string not null,
	native_token string not null,
	unique(client_id)
)
`

const insertChain = `
UPSERT INTO cns.chains 
	(client_id, chain_name, chain_id, native_token) 
VALUES
	(:client_id, :chain_name, :chain_id, :native_token)
`

const getAllChains = `
SELECT * FROM cns.chains
`

var migrationList = []string{
	createDatabase,
	createTableChains,
}

func (i *Instance) runMigrations() {
	if err := database.RunMigrations(i.connString, migrationList); err != nil {
		panic(err)
	}
}
