


% docker exec bitcoind cat /root/.bitcoin/bitcoin.conf
regtest=1
listen=0
server=1
txindex=1

[regtest]
rpcuser=myuser
rpcpassword=SomeDecentp4ssw0rd
rpcclienttimeout=30
rpcallowip=::/0
rpcport=18443

printtoconsole=1
dbcache=512%     



스크립트 수정:
확인한 RPC 사용자 이름과 비밀번호로 스크립트를 수정합니다. execute_rpc 함수에서 다음 줄을 변경하세요:
bash
response=$(curl -v -s --user "실제_rpcuser:실제_rpcpassword" \



macOS: brew install jq
Ubuntu/Debian: sudo apt-get install jq


eun@boui-MacBookPro regtest % ./bitcoin-regtest.sh create_address p2pkh
Generating new p2pkh address...
{
  "address": "mx33bvmnAX3pASBEp2B4zmVeLGuYBFwCsq",
  "privatekey": "cTuxjJisknvTN1oqa9UjWQ214ET919ApZAqXtEKoeoqofMaKeBfm",
  "type": "p2pkh"
}
eun@boui-MacBookPro regtest % ./bitcoin-regtest.sh generate mx33bvmnAX3pASBEp2B4zmVeLGuYBFwCsq 500
Generating 500 blocks to address mx33bvmnAX3pASBEp2B4zmVeLGuYBFwCsq...
Successfully generated 500 blocks.
First and last block hashes:
"46e352c36c7dc05602054d46d2e9c40d2062207e9e612b25c0c787b8b0efabfe"
"0654c18df017daac578bc7d18d37b7c47b2f5432bb368178f4d7ce6f2cb56b3d"
eun@boui-MacBookPro regtest % ./bitcoin-regtest.sh balance mx33bvmnAX3pASBEp2B4zmVeLGuYBFwCsq
Checking balance for address mx33bvmnAX3pASBEp2B4zmVeLGuYBFwCsq...
Spendable balance: 8.254e-05 BTC



// regtest explorer
cd explorer
npm i && npm start
