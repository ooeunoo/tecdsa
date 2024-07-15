#!/bin/bash

start() {
    docker-compose up -d
    echo "Bitcoin regtest node is starting..."
    docker-compose logs -f bitcoin
}

stop() {
    docker-compose down
    echo "Bitcoin regtest node is stopping..."
}

status() {
    docker-compose ps
}

cli() {
    docker-compose exec bitcoin bitcoin-cli -regtest -rpcuser=bitcoinrpc -rpcpassword=CkWFeKaVVNbF5yGgJ1Dhyg== "$@"
}


case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        start
        ;;
    status)
        status
        ;;
    cli)
        shift
        cli "$@"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|cli}"
        exit 1
esac

exit 0