package balancehistory

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/iotexproject/iotex-address/address"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

var (
	sdb             *sql.DB
	dbOpened        bool
	currentHeight   uint64
	accountBalances = make(map[string]*big.Int)
)

type Database struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func (cfg *Database) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port)
}

func Open(cfg *Database) (*sql.DB, error) {
	var err error
	dbOpened = true
	sdb, err = sql.Open("postgres", cfg.DSN())
	return sdb, err
}

func Close() error {
	dbOpened = false
	return sdb.Close()
}

/*
DROP TABLE IF EXISTS "public"."accounts";
CREATE TABLE "public"."accounts" (

	"id" serial8 NOT NULL,
	"block_height" int8 NOT NULL,
	"address" varchar(42) NOT NULL DEFAULT '',
	"balance" numeric(42,0) NOT NULL DEFAULT 0

)PARTITION BY RANGE (block_height);

CREATE TABLE accounts_0_5000000 PARTITION OF accounts FOR VALUES FROM (0) TO (5000000);
CREATE TABLE accounts_5000000_10000000 PARTITION OF accounts FOR VALUES FROM (5000000) TO (10000000);
CREATE TABLE accounts_10000000_15000000 PARTITION OF accounts FOR VALUES FROM (10000000) TO (15000000);
CREATE TABLE accounts_15000000_20000000 PARTITION OF accounts FOR VALUES FROM (15000000) TO (20000000);
CREATE TABLE accounts_20000000_25000000 PARTITION OF accounts FOR VALUES FROM (20000000) TO (25000000);
CREATE TABLE accounts_25000000_30000000 PARTITION OF accounts FOR VALUES FROM (25000000) TO (30000000);
CREATE TABLE accounts_30000000_35000000 PARTITION OF accounts FOR VALUES FROM (30000000) TO (35000000);
CREATE TABLE accounts_35000000_40000000 PARTITION OF accounts FOR VALUES FROM (35000000) TO (40000000);
CREATE TABLE accounts_40000000_45000000 PARTITION OF accounts FOR VALUES FROM (40000000) TO (45000000);
CREATE TABLE accounts_45000000_50000000 PARTITION OF accounts FOR VALUES FROM (45000000) TO (50000000);

CREATE INDEX "idx_accounts_block_height" ON accounts("block_height");
*/
func sqlStoreAccount(height uint64, accounts map[string]*big.Int) error {
	if !dbOpened || len(accounts) == 0 {
		return nil
	}
	valueStrings := make([]string, 0, len(accounts))
	valueArgs := make([]interface{}, 0, len(accounts)*3)
	i := 0
	for addr, balance := range accounts {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, height, addr, balance.String())
		i++
	}
	stmt := fmt.Sprintf("INSERT INTO accounts (block_height, address, balance) VALUES %s", strings.Join(valueStrings, ","))
	_, err := sdb.Exec(stmt, valueArgs...)
	return err
}

func StoreAccountBalance(height uint64, addr address.Address, balance *big.Int) error {
	if height != currentHeight {
		if err := sqlStoreAccount(currentHeight, accountBalances); err != nil {
			return errors.Wrap(err, "failed to store account balance")
		}
		accountBalances = make(map[string]*big.Int)
		currentHeight = height
	}

	accountBalances[addr.String()] = balance
	return nil
}
