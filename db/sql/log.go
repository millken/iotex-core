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
	"account_address" varchar(42) COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::character varying,
	"balance" numeric(42,0) NOT NULL DEFAULT 0,
	"voting_weight" numeric(42,0) NOT NULL DEFAULT 0,
	"is_contract" bool NOT NULL DEFAULT false,
	"is_candidate" bool NOT NULL DEFAULT false

)
;
ALTER TABLE "public"."accounts" OWNER TO "postgres";

-- ----------------------------
-- Indexes structure for table accounts
-- ----------------------------
CREATE UNIQUE INDEX "accounts_block_height_account_address_idx" ON "public"."accounts" USING btree (

	"block_height" "pg_catalog"."int8_ops" ASC NULLS LAST,
	"account_address" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST

);
CREATE INDEX "accounts_block_height_idx" ON "public"."accounts" USING btree (

	"block_height" "pg_catalog"."int8_ops" ASC NULLS LAST

);

-- ----------------------------
-- Primary Key structure for table accounts
-- ----------------------------
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
