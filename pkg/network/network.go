package network

type Network int

const (
	Bitcoin Network = iota
	BitcoinTestNet
	Ethereum
	Ethereum_Sepolia
)

var Networks = []Network{
	Bitcoin, BitcoinTestNet, Ethereum, Ethereum_Sepolia,
}

func (n Network) String() string {
	return [...]string{
		"Bitcoin", "Bitcoin Testnet", "Ethereum", "Ethereum Sepolia",
	}[n]
}
