#!/bin/bash

# 설정
RPC_USER="myuser"
RPC_PASSWORD="SomeDecentp4ssw0rd"
RPC_PORT="18443"
BITCOIN_DATA_DIR="$(pwd)/bitcoin_data"

# RPC 명령 실행 함수
execute_rpc() {
    local method=$1
    local params=$2
    curl -s -u "$RPC_USER:$RPC_PASSWORD" \
        -d "{\"jsonrpc\":\"1.0\",\"id\":\"curltest\",\"method\":\"$method\",\"params\":$params}" \
        -H 'content-type: application/json;' \
        http://127.0.0.1:$RPC_PORT/
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
    mkdir -p "$BITCOIN_DATA_DIR"
    echo "rpcuser=$RPC_USER" > "$BITCOIN_DATA_DIR/bitcoin.conf"
    echo "rpcpassword=$RPC_PASSWORD" >> "$BITCOIN_DATA_DIR/bitcoin.conf"
    echo "rpcallowip=0.0.0.0/0" >> "$BITCOIN_DATA_DIR/bitcoin.conf"
    echo "server=1" >> "$BITCOIN_DATA_DIR/bitcoin.conf"
    docker run --name bitcoind -d \
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

# Explorer 실행 함수
open_explorer() {
    echo "Starting btc-rpc-explorer..."
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
        -e BTCEXP_BASIC_AUTH_PASSWORD=mypassword \
        btc-rpc-explorer
    echo "btc-rpc-explorer is running. You can access it at http://localhost:3002"
    echo "Use any username and 'mypassword' as the password to log in."
}

# Explorer 중지 함수
stop_explorer() {
    echo "Stopping btc-rpc-explorer..."
    if docker ps -q -f name=btc-explorer | grep -q .; then
        docker stop btc-explorer
        docker rm btc-explorer
        echo "btc-rpc-explorer container has been stopped and removed."
    else
        echo "btc-rpc-explorer container is not running."
    fi
    if docker images -q btc-rpc-explorer | grep -q .; then
        docker rmi btc-rpc-explorer
        echo "btc-rpc-explorer image has been removed."
    else
        echo "btc-rpc-explorer image not found."
    fi
    docker network rm bitcoin-network 2>/dev/null
    echo "btc-rpc-explorer has been fully removed."
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
# 주소 가져오기 함수
import_address() {
    local address=$1
    echo "Importing address $address to the wallet..."
    local result=$(execute_rpc "importaddress" "[\"$address\", \"\", false]")
    local error=$(echo "$result" | jq -r '.error')
    
    if [ "$error" != "null" ]; then
        echo "Error: Failed to import address. Response: $result"
        return 1
    else
        echo "Address $address has been imported successfully."
        
        # Verify the import by checking if the address is in the wallet
        local verify_result=$(execute_rpc "getaddressinfo" "[\"$address\"]")
        local is_mine=$(echo "$verify_result" | jq -r '.result.ismine')
        
        if [ "$is_mine" = "true" ]; then
            echo "Verified: The address is now in the wallet."
        else
            echo "Warning: The address import might have failed. Please check manually."
        fi
    fi
}

# 잔액 확인 함수
get_balance() {
    local address=$1
    if [ -z "$address" ]; then
        echo "Checking total wallet balance..."
        local result=$(execute_rpc "getbalance" "[]")
        local balance=$(echo "$result" | jq -r '.result')
    else
        echo "Checking balance for address $address..."
        local result=$(execute_rpc "listunspent" "[0, 9999999, [\"$address\"]]")
        if [[ $result == *"\"result\":[]"* ]]; then
            balance="0"
        else
            balance=$(echo "$result" | jq -r '.result | map(.amount) | add')
        fi
    fi
    if [ -z "$balance" ]; then
        echo "Error: Failed to get balance. Response: $result"
        return 1
    fi
    echo "Spendable balance: $balance BTC"
}

# 비트코인 전송 함수
send_bitcoin() {
    local to_address=$1
    local amount=$2
    echo "Sending $amount BTC to $to_address..."
    execute_rpc "sendtoaddress" "[\"$to_address\",$amount]" | jq '.'
}

# 블록 생성 함수
generate_to_address() {
    local address=$1
    local blocks=$2
    echo "Generating $blocks blocks to address $address..."
    execute_rpc "generatetoaddress" "[$blocks,\"$address\"]" | jq '.'
}


# 트랜잭션 조회 함수
get_transaction() {
    local txid=$1
    echo "Fetching transaction details for TXID: $txid"
    execute_rpc "gettransaction" "[\"$txid\"]" | jq '.'
}

# 사용법 출력 함수
print_usage() {
    echo "Usage:"
    echo "  $0 start                     - Start Bitcoin regtest node"
    echo "  $0 stop                      - Stop Bitcoin regtest node"
    echo "  $0 create_address <type>     - Create a new Bitcoin address"
    echo "  $0 send <address> <amount>   - Send Bitcoin to an address"
    echo "  $0 generate <address> <blocks> - Generate blocks with rewards going to the specified address"
    echo "  $0 balance <address>         - Check balance of an address"
    echo "  $0 tx <txid>                 - Get transaction details"
    echo "  $0 explorer start            - Start btc-rpc-explorer"
    echo "  $0 explorer stop             - Stop btc-rpc-explorer"
}

# 메인 로직
case "$1" in
    start) start_bitcoin_node ;;
    stop) stop_bitcoin_node ;;
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
    send)
        if [ $# -eq 3 ]; then
            send_bitcoin "$2" "$3"
        else
            echo "Error: 'send' command requires an address and an amount."
            print_usage
            exit 1
        fi
        ;;
    generate)
        if [ $# -eq 3 ]; then
            generate_to_address "$2" "$3"
        else
            echo "Error: 'generate' command requires an address and number of blocks."
            print_usage
            exit 1
        fi
        ;;
    balance)
        if [ $# -eq 2 ]; then
            get_balance "$2"
        else
            echo "Error: 'balance' command requires an address."
            print_usage
            exit 1
        fi
        ;;
    tx)
        if [ $# -eq 2 ]; then
            get_transaction "$2"
        else
            echo "Error: 'tx' command requires a transaction ID."
            print_usage
            exit 1
        fi
        ;;
    import_address)
        if [ $# -eq 2 ]; then
            import_address "$2"
        else
            echo "Error: 'import_address' command requires an address."
            echo "Usage: $0 import_address <address>"
            exit 1
        fi
        ;;
    explorer)
        case "$2" in
            start) open_explorer ;;
            stop) stop_explorer ;;
            *)
                echo "Error: Unknown explorer command '$2'"
                echo "Usage: $0 explorer (start|stop)"
                exit 1
                ;;
        esac
        ;;
    *)
        echo "Error: Unknown command '$1'"
        print_usage
        exit 1
        ;;
esac