package sql

import (
	"math/big"

	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-core/action/protocol"
	"github.com/iotexproject/iotex-core/state"
)

func StoreAccount(sm protocol.StateManager, addr address.Address, account *state.Account) error {
	accountProto := account.ToProto()
	votingWeight := new(big.Int).SetBytes(accountProto.VotingWeight).String()
	height, err := sm.Height()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO accounts (block_height, account_address, balance, voting_weight, is_contract, is_candidate) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (block_height,account_address) DO UPDATE SET balance = excluded.balance,voting_weight = excluded.voting_weight,is_contract = excluded.is_contract,is_candidate = excluded.is_candidate", height, addr.String(), account.Balance.String(), votingWeight, account.IsContract(), accountProto.IsCandidate)
	return err
}
