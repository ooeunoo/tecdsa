#!/bin/bash

# 설정
RPC_USER="myuser"
RPC_PASSWORD="SomeDecentp4ssw0rd"
RPC_PORT="18443"
BITCOIN_DATA_DIR="$(pwd)/bitcoin_data"


start_bitcoin_node() {
    echo "Starting Bitcoin regtest node..."
    
    # Create the Docker network if it doesn't exist
    docker network create bitcoin-network 2>/dev/null
    
    # Start the Bitcoin node
    docker run --name bitcoind -d \
        --platform linux/amd64 \
        --network bitcoin-network \
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

build_explorer() {
    echo "Building btc-rpc-explorer image..."
    if [ ! -d "btc-rpc-explorer" ]; then
        git clone https://github.com/janoside/btc-rpc-explorer.git
    fi
    cd btc-rpc-explorer
    docker build -t btc-rpc-explorer .
    cd ..
}

open_explorer() {
    echo "Starting btc-rpc-explorer..."
    
    # Create a Docker network if it doesn't exist
    docker network create bitcoin-network 2>/dev/null

    # Run Bitcoin Core if it's not already running
    start_bitcoin_node
    
    # Build btc-rpc-explorer image if it doesn't exist
    if [[ "$(docker images -q btc-rpc-explorer 2> /dev/null)" == "" ]]; then
        build_explorer
    fi
    
    # Get Bitcoin node's IP address
    BITCOIN_CONTAINER_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' bitcoind)

    # Run btc-rpc-explorer
    docker run -d --name btc-explorer \
        --network bitcoin-network \
        -p 3002:3002 \
        -e BTCEXP_HOST=0.0.0.0 \
        -e BTCEXP_BITCOIND_HOST=bitcoind \
        -e BTCEXP_BITCOIND_PORT=$RPC_PORT \
        -e BTCEXP_BITCOIND_USER=$RPC_USER \
        -e BTCEXP_BITCOIND_PASS=$RPC_PASSWORD \
        -e BTCEXP_BITCOIND_RPC_TIMEOUT=10000 \
        -e BTCEXP_ADDRESS_API=none \
        -e BTCEXP_SLOW_DEVICE_MODE=false \
        -e BTCEXP_NO_RATES=true \
        -e BTCEXP_PRIVACY_MODE=true \
        -e BTCEXP_RPC_ALLOWALL=true \
        btc-rpc-explorer
    echo "btc-rpc-explorer is running. You can access it at http://localhost:3002"
}

stop_explorer() {
    echo "Stopping btc-rpc-explorer..."
    if docker ps -q -f name=btc-explorer | grep -q .; then
        docker stop btc-explorer
        docker rm btc-explorer
        echo "btc-rpc-explorer container has been stopped and removed."
    else
        echo "btc-rpc-explorer container is not running."
    fi

    # Remove the btc-rpc-explorer image
    if docker images -q btc-rpc-explorer | grep -q .; then
        docker rmi btc-rpc-explorer
        echo "btc-rpc-explorer image has been removed."
    else
        echo "btc-rpc-explorer image not found."
    fi

    # Remove the Docker network
    docker network rm bitcoin-network 2>/dev/null

    echo "btc-rpc-explorer has been fully removed."
}



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

# 새 주소 생성 함수 (기본 타입)
create_address() {
    create_address_with_type "legacy"
}
# 특정 타입의 새 주소 생성 함수
create_address_with_type() {
    local address_type=$1
    echo "Generating new $address_type address..."
    
    local result
    case $address_type in
        "p2pkh")
            result=$(execute_rpc "getnewaddress" '["", "legacy"]')
            ;;
        "p2sh")
            result=$(execute_rpc "getnewaddress" '["", "p2sh-segwit"]')
            ;;
        "p2wpkh")
            result=$(execute_rpc "getnewaddress" '["", "bech32"]')
            ;;
        "p2wsh")
            result=$(execute_rpc "addmultisigaddress" '[2, ["'$(execute_rpc "getnewaddress" '["", "bech32"]' | jq -r .result)'", "'$(execute_rpc "getnewaddress" '["", "bech32"]' | jq -r .result)'"], "", "bech32"]')
            ;;
        "p2tr")
            result=$(execute_rpc "getnewaddress" '["", "bech32m"]')
            ;;
        *)
            echo "Error: Invalid address type. Use 'p2pkh', 'p2sh', 'p2wpkh', 'p2wsh', or 'p2tr'."
            return 1
            ;;
    esac

    local address=$(echo "$result" | jq -r '.result')
    if [ -z "$address" ]; then
        echo "Error: Failed to generate new address. Response: $result"
        return 1
    fi

    local privkey_result
    if [ "$address_type" != "p2wsh" ]; then
        privkey_result=$(execute_rpc "dumpprivkey" "[\"$address\"]")
        local privkey=$(echo "$privkey_result" | jq -r '.result')
        if [ -z "$privkey" ]; then
            echo "Error: Failed to get private key. Response: $privkey_result"
            return 1
        fi
        echo "{"
        echo "  \"address\": \"$address\","
        echo "  \"privatekey\": \"$privkey\","
        echo "  \"type\": \"$address_type\""
        echo "}"
    else
        echo "{"
        echo "  \"address\": \"$address\","
        echo "  \"type\": \"$address_type\""
        echo "  \"note\": \"P2WSH is a multisig address, no single private key available\""
        echo "}"
    fi
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

get_balance() {
    local address=$1
    echo "Checking balance for address $address..."
    local result=$(execute_rpc "getreceivedbyaddress" "[\"$address\", 0]")
    local balance=$(echo "$result" | grep -o '"result":[0-9.]*' | cut -d':' -f2)
    if [ -z "$balance" ]; then
        echo "Error: Failed to get balance. Response: $result"
        return 1
    fi
    echo "Total balance (including unconfirmed): $balance BTC"
}

# # 잔액 확인 함수
# get_balance() {
#     local address=$1
#     if [ -z "$address" ]; then
#         echo "Checking total wallet balance..."
#         local result=$(execute_rpc "getbalance" "[]")
#         local balance=$(parse_json "$result" "result")
#     else
#         echo "Checking balance for address $address..."
#         local result=$(execute_rpc "listunspent" "[0, 9999999, [\"$address\"]]")
#         if [[ $result == *"\"result\":[]"* ]]; then
#             balance="0"
#         else
#             balance=$(echo "$result" | grep -o '"amount":[0-9.]*' | cut -d':' -f2 | awk '{sum += $1} END {print sum}')
#         fi
#     fi
#     if [ -z "$balance" ]; then
#         echo "Error: Failed to get balance. Response: $result"
#         return 1
#     fi
#     echo "Spendable balance: $balance BTC"
# }

# 트랜잭션 조회 함수
get_transaction() {
    local txid=$1
    echo "Fetching transaction details for TXID: $txid"
    local result=$(execute_rpc "gettransaction" "[\"$txid\"]")
    if [[ $result == *"\"error\""*"null"* ]]; then
        echo "Transaction details:"
        echo "$result" | jq '.'
    else
        echo "Error: Failed to get transaction details. Response: $result"
        return 1
    fi
}



print_usage() {
    echo "Usage:"
    echo "  $0 start                     - Start Bitcoin regtest node"
    echo "  $0 stop                      - Stop Bitcoin regtest node"
    echo "  $0 create_address <type>     - Create a new Bitcoin address"
    echo "  $0 send <address> <amount>   - Send Bitcoin to an address"
    echo "  $0 generate <address> <blocks> - Generate blocks with rewards going to the specified address"
    echo "  $0 balance [address]         - Check balance of an address or total wallet balance"
    echo "  $0 tx <txid>                 - Get transaction details"
    echo "  $0 explorer start            - Start bitcoin-abe explorer"
    echo "  $0 explorer stop             - Stop bitcoin-abe explorer"

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
        if [ $# -eq 2 ]; then
            create_address_with_type "$2"
        else
            echo "Error: Address type is required."
            echo "Usage: $0 create_address <type>"
            echo "Available types: p2pkh, p2sh, p2wpkh, p2wsh, p2tr"
            exit 1
        fi
        ;;
    tx)
        if [ $# -ne 2 ]; then
            echo "Error: 'tx' command requires a transaction ID."
            print_usage
            exit 1
        fi
        get_transaction "$2"
        ;;
    explorer)
        case "$2" in
            start)
                open_explorer
                ;;
            stop)
                stop_explorer
                ;;
            *)
                echo "Error: Unknown explorer command '$2'"
                echo "Usage: $0 explorer (start|stop)"
                exit 1
                ;;
        esac
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