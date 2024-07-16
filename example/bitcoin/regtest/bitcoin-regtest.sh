#!/bin/bash

# 설정
RPC_USER="myuser"
RPC_PASSWORD="SomeDecentp4ssw0rd"
RPC_PORT="18443"
BITCOIN_DATA_DIR="$HOME/bitcoin_data"

# RPC 명령 실행 함수
execute_rpc() {
    local method=$1
    local params=$2
    local response=$(curl -s -u "$RPC_USER:$RPC_PASSWORD" \
        -d "{\"jsonrpc\":\"1.0\",\"id\":\"curltest\",\"method\":\"$method\",\"params\":$params}" \
        -H 'content-type: application/json;' \
        http://127.0.0.1:$RPC_PORT/)
    echo "$response"
}

# JSON 파싱 함수
parse_json() {
    local json="$1"
    local key="$2"
    echo "$json" | grep -o "\"$key\":\"[^\"]*\"" | sed "s/\"$key\":\"//;s/\"$//"
}

# Bitcoin regtest 노드 실행 함수
start_bitcoin_node() {
    echo "Starting Bitcoin regtest node..."
    docker run --name bitcoind -d \
        --volume $BITCOIN_DATA_DIR:/root/.bitcoin \
        -p 127.0.0.1:$RPC_PORT:$RPC_PORT \
        farukter/bitcoind:regtest
    echo "Bitcoin node started. Please wait a moment for it to fully initialize."
}

# Bitcoin regtest 노드 중지 함수
stop_bitcoin_node() {
    echo "Stopping Bitcoin regtest node..."
    docker stop bitcoind
    docker rm bitcoind
    echo "Bitcoin node stopped and container removed."
}

# 새 주소 생성 함수
create_address() {
    echo "Generating new address..."
    local result=$(execute_rpc "getnewaddress" "[]")
    local address=$(parse_json "$result" "result")
    if [ -z "$address" ]; then
        echo "Error: Failed to generate new address. Response: $result"
        return 1
    fi

    local privkey_result=$(execute_rpc "dumpprivkey" "[\"$address\"]")
    local privkey=$(parse_json "$privkey_result" "result")
    if [ -z "$privkey" ]; then
        echo "Error: Failed to get private key. Response: $privkey_result"
        return 1
    fi
    echo "{"
    echo "  \"address\": \"$address\","
    echo "  \"privatekey\": \"$privkey\""
    echo "}"
}

# 비트코인 전송 함수
send_bitcoin() {
    local to_address=$1
    local amount=$2
    echo "Sending $amount BTC to $to_address..."
    local result=$(execute_rpc "sendtoaddress" "[\"$to_address\",$amount]")
    local txid=$(parse_json "$result" "result")
    if [ -z "$txid" ]; then
        echo "Error: Failed to send Bitcoin. Response: $result"
        return 1
    fi
    echo "Transaction ID: $txid"
}

# 특정 주소로 비트코인 생성 함수
generate_to_address() {
    local address=$1
    local blocks=$2
    echo "Generating $blocks blocks to address $address..."
    local result=$(execute_rpc "generatetoaddress" "[$blocks,\"$address\"]")
    if [[ $result == *"\"error\""*"null"* ]]; then
        echo "Successfully generated $blocks blocks."
        echo "First and last block hashes:"
        echo "$result" | grep -o '\[.*\]' | sed 's/\[//;s/\]//;s/,/\n/g' | sed -n '1p;$p'
    else
        echo "Error: Failed to generate blocks. Response: $result"
        return 1
    fi
}

# 잔액 확인 함수
get_balance() {
    local address=$1
    if [ -z "$address" ]; then
        echo "Checking total wallet balance..."
        local result=$(execute_rpc "getbalance" "[]")
        local balance=$(parse_json "$result" "result")
    else
        echo "Checking balance for address $address..."
        local result=$(execute_rpc "listunspent" "[0, 9999999, [\"$address\"]]")
        if [[ $result == *"\"result\":[]"* ]]; then
            balance="0"
        else
            balance=$(echo "$result" | grep -o '"amount":[0-9.]*' | cut -d':' -f2 | awk '{sum += $1} END {print sum}')
        fi
        local immature_result=$(execute_rpc "listunspent" "[0, 99, [\"$address\"]]")
        local immature_balance=$(echo "$immature_result" | grep -o '"amount":[0-9.]*' | cut -d':' -f2 | awk '{sum += $1} END {print sum}')
        echo "Immature balance (not yet spendable): $immature_balance BTC"
    fi
    if [ -z "$balance" ]; then
        echo "Error: Failed to get balance. Response: $result"
        return 1
    fi
    echo "Spendable balance: $balance BTC"
}

# 사용법 출력 함수
print_usage() {
    echo "Usage:"
    echo "  $0 start                     - Start Bitcoin regtest node"
    echo "  $0 stop                      - Stop Bitcoin regtest node"
    echo "  $0 create_address            - Create a new Bitcoin address"
    echo "  $0 send <address> <amount>   - Send Bitcoin to an address"
    echo "  $0 generate <address> <blocks> - Generate blocks with rewards going to the specified address"
    echo "  $0 balance [address]         - Check balance of an address or total wallet balance"
}

# 메인 로직
case "$1" in
    start)
        start_bitcoin_node
        ;;
    stop)
        stop_bitcoin_node
        ;;
    create_address)
        create_address
        ;;
    send)
        if [ $# -ne 3 ]; then
            echo "Error: 'send' command requires an address and an amount."
            print_usage
            exit 1
        fi
        send_bitcoin "$2" "$3"
        ;;
    generate)
        if [ $# -ne 3 ]; then
            echo "Error: 'generate' command requires an address and number of blocks."
            print_usage
            exit 1
        fi
        generate_to_address "$2" "$3"
        ;;
    balance)
        get_balance "$2"
        ;;
    *)
        echo "Error: Unknown command '$1'"
        print_usage
        exit 1
        ;;
esac