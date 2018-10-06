// TODO : License related notes

package wallet

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
	"github.com/btcsuite/btcwallet/walletdb"
	"github.com/btcsuite/btcwallet/wtxmgr"
)

// TODO : Write summary
func (w *Wallet) txTransferToOutputs(address string, txHash chainhash.Hash, account uint32,
	minconf int32, feeSatPerKB btcutil.Amount) (tx *txauthor.AuthoredTx, err error) {
	//TODO : Implement this
	fmt.Printf("txTransferToOutputs")

	chainClient, err := w.requireChainClient()
	if err != nil {
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

		txToBoTransferred, err := w.findTheTransaction(dbtx, txHash, account, minconf, bs)
		if err != nil {
			return err
		}

		// Make outputs for tx to be transferred
		amount := txToBoTransferred.Amount
		_ = amount

		return nil
	})

	return nil, nil
}

// findTransaction is modified from 'findEligibleOutputs' in 'createtx.go'
func (w *Wallet) findTheTransaction(dbtx walletdb.ReadTx, txHash chainhash.Hash,
	account uint32, minconf int32, bs *waddrmgr.BlockStamp) (wtxmgr.Credit, error) {

	addrmgrNs := dbtx.ReadBucket(waddrmgrNamespaceKey)
	txmgrNs := dbtx.ReadBucket(wtxmgrNamespaceKey)

	// TODO : Eventually get from post-dated transactions (POST-DATED feature)
	unspent, err := w.TxStore.UnspentOutputs(txmgrNs)
	if err != nil {
		return wtxmgr.Credit{}, err
	}

	// TODO: Eventually all of these filters (except perhaps output locking)
	// should be handled by the call to UnspentOutputs (or similar).
	// Because one of these filters requires matching the output script to
	// the desired account, this change depends on making wtxmgr a waddrmgr
	// dependancy and requesting unspent outputs for a single account.
	var eligible wtxmgr.Credit
	for i := range unspent {
		output := &unspent[i]

		// TODO :  Check this
		// For post-dated cheques, we want to transfer only tx with txHash
		if output.Hash != txHash {
			continue
		}

		// Only include this output if it meets the required number of
		// confirmations.  Coinbase transactions must have have reached
		// maturity before their outputs may be spent.
		if !confirmed(minconf, output.Height, bs.Height) {
			// TODO : return a custom 'not mature' error
			continue
		}
		if output.FromCoinBase {
			target := int32(w.chainParams.CoinbaseMaturity)
			if !confirmed(target, output.Height, bs.Height) {
				// TODO : return a custom 'not mature coinbase' error
				continue
			}
		}

		// Locked unspent outputs are skipped.
		if w.LockedOutpoint(output.OutPoint) {
			// TODO : return a custom 'locked tx' error
			continue
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
			// TODO : return a custom 'not an owned tx' error
			continue
		}
		eligible = *output
	}
	return eligible, nil
}
