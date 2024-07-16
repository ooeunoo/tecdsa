


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



brew install jq