package sql

import (
	"math/big"

	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-core/action/protocol"
	"github.com/iotexproject/iotex-core/state"
)

/*
DROP TABLE IF EXISTS "public"."accounts";
CREATE TABLE "public"."accounts" (

	"id" serial8 NOT NULL,
	"block_height" int8 NOT NULL,
	"account_address" varchar(42) NOT NULL DEFAULT '',
	"balance" numeric(42,0) NOT NULL DEFAULT 0,
	"voting_weight" numeric(42,0) NOT NULL DEFAULT 0,
	"is_contract" bool NOT NULL DEFAULT false,
	"is_candidate" bool NOT NULL DEFAULT false

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

CREATE UNIQUE INDEX idx_accounts_block_height_account_address
ON accounts(block_height, account_address);

CREATE INDEX "idx_accounts_block_height" ON accounts("block_height");

ALTER TABLE "public"."accounts" ADD CONSTRAINT "account_balance_pkey" PRIMARY KEY ("id");
*/
func StoreAccount(sm protocol.StateManager, addr address.Address, account *state.Account) error {
	if !dbOpened {
		return nil
	}

	accountProto := account.ToProto()
	votingWeight := new(big.Int).SetBytes(accountProto.GetVotingWeight()).String()
	height, err := sm.Height()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO accounts (block_height, account_address, balance, voting_weight, is_contract, is_candidate) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (block_height,account_address) DO UPDATE SET balance = excluded.balance,voting_weight = excluded.voting_weight,is_contract = excluded.is_contract,is_candidate = excluded.is_candidate", height, addr.String(), account.Balance.String(), votingWeight, account.IsContract(), accountProto.GetIsCandidate())
	return err
}
