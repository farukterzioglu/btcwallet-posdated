// TODO : Change this for transaction transfer rules  
(Taken from https://en.bitcoin.it/wiki/Protocol_rules)  

These messages hold a single transaction.  

Check syntactic correctness  
Make sure neither in or out lists are empty  
Size in bytes <= MAX_BLOCK_SIZE  
Each output value, as well as the total, must be in legal money range  
Make sure none of the inputs have hash=0, n=-1 (coinbase transactions)  
Check that nLockTime <= INT_MAX, size in bytes >= 100, and sig opcount <= 2  
Reject "nonstandard" transactions: scriptSig doing anything other than pushing numbers on the stack, or scriptPubkey not matching the two usual forms    
Reject if we already have matching tx in the pool, or in a block in the main branch  
For each input, if the referenced output exists in any other tx in the pool, reject this transaction.  
For each input, look in the main branch and the transaction pool to find the referenced output transaction. If the output transaction is missing for any input, this will be an orphan transaction. Add to the orphan transactions, if a   matching transaction is not in there already.  
For each input, if the referenced output transaction is coinbase (i.e. only 1 input, with hash=0, n=-1), it must have at least COINBASE_MATURITY (100) confirmations; else reject this transaction  
For each input, if the referenced output does not exist (e.g. never existed or has already been spent), reject this transaction  
Using the referenced output transactions to get input values, check that each input value, as well as the sum, are in legal money range  
Reject if the sum of input values < sum of output values  
Reject if transaction fee (defined as sum of input values minus sum of output values) would be too low to get into an empty block  
Verify the scriptPubKey accepts for each input; reject if any are bad  
Add to transaction pool  
"Add to wallet if mine"  
Relay transaction to peers  
For each orphan transaction that uses this one as one of its inputs, run all these steps (including this one) recursively on that orphan  