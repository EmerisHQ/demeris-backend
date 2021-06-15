package database

import "github.com/allinbits/demeris-backend/utils/database"

const createDatabase = `
CREATE DATABASE oracle;
`

const createTableBinance = `
CREATE TABLE oracle.binance (symbol STRING PRIMARY KEY, price FLOAT, updatedat INT);
`
const createTableCoinmarketcap = `
CREATE TABLE oracle.coinmarketcap (symbol STRING PRIMARY KEY, price FLOAT, updatedat INT);
`
const createTableFixer = `
CREATE TABLE oracle.fixer (symbol STRING PRIMARY KEY, price FLOAT, updatedat INT);
`
const createTableTokens = `
CREATE TABLE oracle.tokens (symbol STRING PRIMARY KEY, price FLOAT);
`
const createTableFiats = `
CREATE TABLE oracle.fiats (symbol STRING PRIMARY KEY, price FLOAT);
`

var migrationList = []string{
	createDatabase,
	createTableBinance,
	createTableCoinmarketcap,
	createTableFixer,
	createTableTokens,
	createTableFiats,
}

func (i *Instance) runMigrations() {
	if err := database.RunMigrations(i.connString, migrationList); err != nil {
		panic(err)
	}
}
