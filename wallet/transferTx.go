// TODO : License related notes

package wallet

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
	"github.com/btcsuite/btcwallet/wallet/txrules"
	"github.com/btcsuite/btcwallet/walletdb"
	"github.com/btcsuite/btcwallet/wtxmgr"
)

// TODO : Write summary
func (w *Wallet) txTransferToOutputs(address string, txHash chainhash.Hash, account uint32,
	minconf int32, feeSatPerKb btcutil.Amount) (tx *txauthor.AuthoredTx, err error) {
	fmt.Printf("txTransferToOutputs")

	chainClient, err := w.requireChainClient()
	if err != nil {
		return nil, err
	}

	// Find tx to be transferred
	var txToBoTransferred *wtxmgr.Credit
	// Open a database read transaction and executes the function f
	// Used to find transaction with hash 'txHash'
	err = walletdb.View(w.db, func(dbtx walletdb.ReadTx) error {
		// Get current block's height and hash.
		bs, err := chainClient.BlockStamp()
		if err != nil {
			return err
		}

		// Find only the transaction with hash 'txHash' that belong to 'account'
		// If not found any, or the found one isn't eligible, throw a relevant exception
		// Eligible if : has 'minconf' confirmation & unspent
		// Eventually will return only post-dated cheques
		txToBoTransferred, err = w.findTheTransaction(dbtx, txHash, account, minconf, bs)
		return err
	})
	if err != nil {
		return nil, err
	}

	amount := txToBoTransferred.Amount
	// Make outputs for tx to be transferred
	redeemOutput, err := makeOutput(address, amount, w.ChainParams())
	if err != nil {
		return nil, err
	}

	// Ensure the outputs to be created adhere to the network's consensus
	// rules.
	if err := txrules.CheckOutput(redeemOutput, feeSatPerKb); err != nil {
		return nil, err
	}

	// Open a database read/write transaction and executes the function f
	err = walletdb.Update(w.db, func(dbtx walletdb.ReadWriteTx) error {
		addrmgrNs := dbtx.ReadWriteBucket(waddrmgrNamespaceKey)
		_ = addrmgrNs

		// Get current block's height and hash.
		bs, err := chainClient.BlockStamp()
		if err != nil {
			return err
		}

		eligible, err := w.findEligibleOutputs(dbtx, account, minconf, bs)
		if err != nil {
			return err
		}
		inputSource := makeInputSource(eligible)
		changeSource := func() ([]byte, error) {
			// Derive the change output script.  As a hack to allow
			// spending from the imported account, change addresses
			// are created from account 0.
			var changeAddr btcutil.Address
			var err error
			if account == waddrmgr.ImportedAddrAccount {
				changeAddr, err = w.newChangeAddress(addrmgrNs, 0)
			} else {
				changeAddr, err = w.newChangeAddress(addrmgrNs, account)
			}
			if err != nil {
				return nil, err
			}
			return txscript.PayToAddrScript(changeAddr)
		}

		outputs := []*wire.TxOut{redeemOutput}
		tx, err = txauthor.NewUnsignedTransactionFromInput(txToBoTransferred, outputs, feeSatPerKb,
			inputSource, changeSource)
		if err != nil {
			return err
		}

		// Randomize change position, if change exists, before signing.
		// This doesn't affect the serialize size, so the change amount
		// will still be valid.
		if tx.ChangeIndex >= 0 {
			tx.RandomizeChangePosition()
		}

		return tx.AddAllInputScripts(secretSource{w.Manager, addrmgrNs})
	})
	if err != nil {
		return nil, err
	}

	err = validateMsgTx(tx.Tx, tx.PrevScripts, tx.PrevInputValues)
	if err != nil {
		return nil, err
	}

	if tx.ChangeIndex >= 0 && account == waddrmgr.ImportedAddrAccount {
		changeAmount := btcutil.Amount(tx.Tx.TxOut[tx.ChangeIndex].Value)
		log.Warnf("Spend from imported account produced change: moving"+
			" %v from imported account into default account.", changeAmount)
	}

	return tx, nil
}

// makeOutput creates a transaction output from a pair of address
// strings to amounts.  This is used to create the outputs to include in newly
// created transactions from a JSON object describing the output destinations
// and amounts.
func makeOutput(addrStr string, amt btcutil.Amount, chainParams *chaincfg.Params) (*wire.TxOut, error) {
	addr, err := btcutil.DecodeAddress(addrStr, chainParams)
	if err != nil {
		return nil, fmt.Errorf("cannot decode address: %s", err)
	}

	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, fmt.Errorf("cannot create txout script: %s", err)
	}

	output := wire.NewTxOut(int64(amt), pkScript)
	return output, nil
}

// findTransaction is modified from 'findEligibleOutputs' in 'createtx.go'
func (w *Wallet) findTheTransaction(dbtx walletdb.ReadTx, txHash chainhash.Hash,
	account uint32, minconf int32, bs *waddrmgr.BlockStamp) (*wtxmgr.Credit, error) {

	addrmgrNs := dbtx.ReadBucket(waddrmgrNamespaceKey)
	txmgrNs := dbtx.ReadBucket(wtxmgrNamespaceKey)

	// TODO : Eventually get from post-dated transactions (POST-DATED feature)
	unspent, err := w.TxStore.UnspentOutputs(txmgrNs)
	if err != nil {
		return nil, err
	}

	// TODO: Eventually all of these filters (except perhaps output locking)
	// should be handled by the call to UnspentOutputs (or similar).
	// Because one of these filters requires matching the output script to
	// the desired account, this change depends on making wtxmgr a waddrmgr
	// dependancy and requesting unspent outputs for a single account.
	var eligible *wtxmgr.Credit
	for i := range unspent {
		output := &unspent[i]

		// For post-dated cheques, we want to transfer only tx with txHash
		if output.Hash != txHash {
			continue
		}

		// Only include this output if it meets the required number of
		// confirmations.  Coinbase transactions must have have reached
		// maturity before their outputs may be spent.
		if !confirmed(minconf, output.Height, bs.Height) {
			return nil, txNotMatureError{}
		}
		if output.FromCoinBase {
			target := int32(w.chainParams.CoinbaseMaturity)
			if !confirmed(target, output.Height, bs.Height) {
				return nil, txCoinbaseNotMatureError{}
			}
		}

		// Locked unspent outputs are skipped.
		if w.LockedOutpoint(output.OutPoint) {
			return nil, txIsLockedError{}
		}

		// Only include the output if it is associated with the passed
		// account.
		//
		// TODO: Handle multisig outputs by determining if enough of the
		// addresses are controlled.
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(
			output.PkScript, w.chainParams)
		if err != nil || len(addrs) != 1 {
			continue
		}
		_, addrAcct, err := w.Manager.AddrAccount(addrmgrNs, addrs[0])
		if err != nil || addrAcct != account {
			return nil, txIsNotOwnedError{}
		}
		eligible = output
	}

	if eligible == nil {
		return nil, txNotFoundError{}
	}

	return eligible, nil
}
