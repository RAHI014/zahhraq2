/*
* Copyright 2020, Offchain Labs, Inc.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*    http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package arbostest

import (
	"github.com/offchainlabs/arbitrum/packages/arb-node-core/test"
	"github.com/offchainlabs/arbitrum/packages/arb-rpc-node/snapshot"
	"math/big"
	"testing"

	"github.com/offchainlabs/arbitrum/packages/arb-evm/evm"
	"github.com/offchainlabs/arbitrum/packages/arb-evm/message"
	"github.com/offchainlabs/arbitrum/packages/arb-util/common"
	"github.com/offchainlabs/arbitrum/packages/arb-util/inbox"
)

func TestWithdrawEth(t *testing.T) {
	addr := common.RandAddress()

	depositMsg := message.Eth{
		Dest:  addr,
		Value: big.NewInt(100),
	}

	withdrawValue := big.NewInt(100)
	withdrawDest := common.RandAddress()
	tx := withdrawEthTx(big.NewInt(0), withdrawValue, withdrawDest)

	chainTime := inbox.ChainTime{
		BlockNum:  common.NewTimeBlocksInt(0),
		Timestamp: big.NewInt(0),
	}

	inboxMessages := []inbox.InboxMessage{
		message.NewInboxMessage(initMsg(), chain, big.NewInt(0), chainTime),
		message.NewInboxMessage(message.Eth{
			Dest:  sender,
			Value: big.NewInt(1000),
		}, sender, big.NewInt(1), chainTime),
		message.NewInboxMessage(depositMsg, common.RandAddress(), big.NewInt(2), chainTime),
		message.NewInboxMessage(message.NewSafeL2Message(tx), sender, big.NewInt(3), chainTime),
	}

	logs, _, _, _ := runAssertion(t, inboxMessages, 2, 0)
	results := processResults(t, logs)

	sendRes, ok := results[0].(*evm.SendResult)
	if !ok {
		t.Fatal("not send res")
	}
	txRes, ok := results[1].(*evm.TxResult)
	if !ok {
		t.Fatal("not tx res")
	}

	succeededTxCheck(t, txRes)

	if len(txRes.EVMLogs) != 1 {
		t.Fatal("unexpected log count")
	}
	ev, err := snapshot.ParseEthWithdrawalEvent(txRes.EVMLogs[0])
	failIfError(t, err)
	if ev.Amount.Cmp(withdrawValue) != 0 {
		t.Error("unexpected withdrawal value")
	}
	if ev.DestAddr != withdrawDest.ToEthAddress() {
		t.Error("unexpected dest address")
	}

	withdrawEth, err := evm.NewWithdrawEthResultFromData(sendRes.Data)
	test.FailIfError(t, err)

	if withdrawEth.Destination != withdrawDest {
		t.Fatal("wrong withdraw sender")
	}

	if withdrawEth.Amount.Cmp(withdrawValue) != 0 {
		t.Fatal("wrong withdraw value", withdrawEth.Amount)
	}
}
