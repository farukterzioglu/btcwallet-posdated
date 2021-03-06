```
btcd -C .\btcd.conf

btcwallet -C .\btcwallet.conf --create
-> Wallet seed (SimNet) 51b63e594eafd12d1393f49936bcee0b8a54c12e506309f21c0d52af6ad0ea47

btcwallet -C .\btcwallet.conf 

btcctl -C .\btcctl.conf --wallet walletpassphrase "[password]" 600

btcctl -C .\btcctl.conf --wallet getnewaddress
-> SYYFCKLX38x6aZRsxuLmdNXSBDZ3BaJczq

btcd -C .\btcd.conf --miningaddr=SYYFCKLX38x6aZRsxuLmdNXSBDZ3BaJczq

btcctl -C .\btcctl.conf --wallet createnewaccount account1

btcctl -C .\btcctl.conf --wallet getnewaddress account1
-> SZMEzGtDCgxmm1WKjbSUqBie3xcU7ovuyG

btcctl -C .\btcctl.conf --wallet getblockhash 0
-> 683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6

btcctl -C .\btcctl.conf --wallet getblock 683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6
-> 
{
  "hash": "683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6",
  "confirmations": 1,
  "strippedsize": 285,
  "size": 285,
  "weight": 1140,
  "height": 0,
  "version": 1,
  "versionHex": "00000001",
  "merkleroot": "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
  "tx": [
    "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
  ],
  "time": 1401292357,
  "nonce": 2,
  "bits": "207fffff",
  "difficulty": 1,
  "previousblockhash": "0000000000000000000000000000000000000000000000000000000000000000"
}

btcctl -C .\btcctl.conf generate 109
-> 
[
	"2a666bba893f07ce7e9ebfa2d0aa0379b0b54bc6487b3715ab5200b74fda9154",
	"1641c6a0d5916cfc2f70b7c89fcd27dfb85fe11886516dc6f1de36b8f4761212",
	....
]

btcctl -C .\btcctl.conf --wallet getblockhash 8
-> 20c703d4ff8b515c634073ebe99cc01ac9f7739dee0507f89d01dbd0fe76c853

btcctl -C .\btcctl.conf --wallet getblock 20c703d4ff8b515c634073ebe99cc01ac9f7739dee0507f89d01dbd0fe76c853
-> 
{
  "hash": "20c703d4ff8b515c634073ebe99cc01ac9f7739dee0507f89d01dbd0fe76c853",
  "confirmations": 100,
  "strippedsize": 188,
  "size": 188,
  "weight": 752,
  "height": 8,
  "version": 536870912,
  "versionHex": "20000000",
  "merkleroot": "de14e6d076a217558dd23191ee6ae3b652297d359f92c6973f3139fe8d757359",
  "tx": [
    "de14e6d076a217558dd23191ee6ae3b652297d359f92c6973f3139fe8d757359"
  ],
  "time": 1539433122,
  "nonce": 0,
  "bits": "207fffff",
  "difficulty": 1,
  "previousblockhash": "4d2d4f8d266c329a24576ab314edace636e376b0cb33d661b18e52c03085f06a",
  "nextblockhash": "3294318179979673976745ab8b83bf5fc33c0399aa42db0299b33dab81ded9ac"
}

btcctl -C .\btcctl.conf --wallet listunspent
-> 
[
  {
    "txid": "e8511bd42a9bbdac69655fd3bb0e5cb2f9c42cc9b51476c65682ba18a18e4c0e",
    "vout": 0,
    "address": "SYYFCKLX38x6aZRsxuLmdNXSBDZ3BaJczq",
    "account": "default",
    "scriptPubKey": "76a9147b5b07b286ef494549db8630aeb37301bf99663c88ac",
    "amount": 50,
    "confirmations": 100,
    "spendable": true
  },
  {
    "txid": "de14e6d076a217558dd23191ee6ae3b652297d359f92c6973f3139fe8d757359",
    "vout": 0,
    "address": "SYYFCKLX38x6aZRsxuLmdNXSBDZ3BaJczq",
    "account": "default",
    "scriptPubKey": "76a9147b5b07b286ef494549db8630aeb37301bf99663c88ac",
    "amount": 50,
    "confirmations": 101,
    "spendable": true
  }
]

btcctl -C .\btcctl.conf --wallet transfertransaction SZMEzGtDCgxmm1WKjbSUqBie3xcU7ovuyG de14e6d076a217558dd23191ee6ae3b652297d359f92c6973f3139fe8d757359
-> d7ff90d4a3a07148446cc98cae786efe3e394e9660614fac262ab1354575f4c4

btcctl -C .\btcctl.conf --wallet listunspent
-> [] (bug here)

btcctl -C .\btcctl.conf generate 1
-> 
[
  "47b7d70a2088854f80a39fa57f392715c2ad7abf8f04e2c7938f6fc683f618c8"
]

btcctl -C .\btcctl.conf --wallet listunspent
-> 
[
  {
    "txid": "d7ff90d4a3a07148446cc98cae786efe3e394e9660614fac262ab1354575f4c4",
    "vout": 1,
    "address": "SZMEzGtDCgxmm1WKjbSUqBie3xcU7ovuyG",
    "account": "account1",
    "scriptPubKey": "76a914843e682e2cad833c2b4056a473423cf4458f38f388ac",
    "amount": 50,
    "confirmations": 1,
    "spendable": true
  },
  {
    "txid": "d7ff90d4a3a07148446cc98cae786efe3e394e9660614fac262ab1354575f4c4",
    "vout": 0,
    "address": "sb1qt05j4gu54adsxrvqfwg9e2wfwn2yzazy39x8hm",
    "account": "default",
    "scriptPubKey": "00145be92aa394af5b030d804b905ca9c974d4417444",
    "amount": 49.99999627,
    "confirmations": 1,
    "spendable": true
  },
  {
    "txid": "33c956a67f0d9423d376fc8fcb27426da727baf162b63e501ed8bb4ff4679aee",
    "vout": 0,
    "address": "SYYFCKLX38x6aZRsxuLmdNXSBDZ3BaJczq",
    "account": "default",
    "scriptPubKey": "76a9147b5b07b286ef494549db8630aeb37301bf99663c88ac",
    "amount": 50,
    "confirmations": 100,
    "spendable": true
  }
]

btcctl -C .\btcctl.conf --wallet gettransaction d7ff90d4a3a07148446cc98cae786efe3e394e9660614fac262ab1354575f4c4
(This tx includes transferred transaction and the change)
-> 
{
  "amount": 50,
  "fee": 0.00000373,
  "confirmations": 1,
  "blockhash": "47b7d70a2088854f80a39fa57f392715c2ad7abf8f04e2c7938f6fc683f618c8",
  "blockindex": 0,
  "blocktime": 1539549532,
  "txid": "d7ff90d4a3a07148446cc98cae786efe3e394e9660614fac262ab1354575f4c4",
  "walletconflicts": [],
  "time": 1539549533,
  "timereceived": 1539549533,
  "details": [
    {
      "account": "",
      "amount": -100,
      "category": "send",
      "fee": 0.00000373,
      "vout": 0
    },
    {
      "account": "account1",
      "address": "SZMEzGtDCgxmm1WKjbSUqBie3xcU7ovuyG",
      "amount": 50,
      "category": "receive",
      "vout": 1
    }
  ],
  "hex": "01000000020e4c8ea118ba8256c67614b5c92cc4f9b25c0ebbd35f6569acbd9b2ad41b51e8000000006a4730440220101c9a344296bd8fb49e48878b0342ba578ab92c38269c72915d79b53763c0440220446b54721493bff502a39eba0b69b5faa4a349bc1f7a6ff7cbbaabb00f4c3f9d01210223ac6a75f94e60de9ad1642982170c958f7eba2d8640830b28fb67e1fe57561dffffffff5973758dfe39313f97c6929f357d2952b6e36aee9131d28d5517a276d0e614de000000006b483045022100856c05b2136a03362aa5f44892caabd90efac09780409deb19955c1bca36858d02205a7563e0f265ab55c4c75e5508fa2e987c8977873a30aa62788985cd5fccc05901210223ac6a75f94e60de9ad1642982170c958f7eba2d8640830b28fb67e1fe57561dffffffff028bf0052a010000001600145be92aa394af5b030d804b905ca9c974d441744400f2052a010000001976a914843e682e2cad833c2b4056a473423cf4458f38f388ac00000000"
}

btcctl -C .\btcctl.conf --wallet gettransaction 33c956a67f0d9423d376fc8fcb27426da727baf162b63e501ed8bb4ff4679aee
(This is the reward of 10. block)
-> 
{
  "amount": 50,
  "confirmations": 100,
  "blockhash": "5887b49fd44146eb24447ae3c8eeb93bb80c76f8b1e931e445ecc1066427cb59",
  "blockindex": 0,
  "blocktime": 1539433122,
  "txid": "33c956a67f0d9423d376fc8fcb27426da727baf162b63e501ed8bb4ff4679aee",
  "walletconflicts": [],
  "time": 1539433120,
  "timereceived": 1539433120,
  "details": [
    {
      "account": "default",
      "address": "SYYFCKLX38x6aZRsxuLmdNXSBDZ3BaJczq",
      "amount": 50,
      "category": "generate",
      "vout": 0
    }
  ],
  "hex": "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff165a08830f409bcb227e3c0b2f503253482f627463642fffffffff0100f2052a010000001976a9147b5b07b286ef494549db8630aeb37301bf99663c88ac00000000"
}

(Block height of 47b7d70a2088854f80a39fa57f392715c2ad7abf8f04e2c7938f6fc683f618c8 : 109)
(Block height of 5887b49fd44146eb24447ae3c8eeb93bb80c76f8b1e931e445ecc1066427cb59 : 10)

```