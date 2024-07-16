package network

type Network int

const (
	Bitcoin Network = iota
	BitcoinTestNet
	BitcoinRegTest
	Ethereum
	Ethereum_Sepolia
	Avalanche_C_CHAIN
	Avalanche_C_CHAIN_Fuji
)

type NetworkMetadataInfo struct {
	ID      int32
	Name    string
	ChainID *int64
	RpcURL  string
}

var NetworkMetadata = map[Network]NetworkMetadataInfo{
	Bitcoin: {
		ID:      1,
		Name:    "Bitcoin",
		ChainID: nil,
		RpcURL:  "",
	},
	BitcoinTestNet: {
		ID:      2,
		Name:    "Bitcoin Testnet",
		ChainID: nil,
		RpcURL:  "",
	},
	BitcoinRegTest: {
		ID:      3,
		Name:    "Bitcoin RegTest",
		ChainID: nil,
		RpcURL:  "",
	},
	Ethereum: {
		ID:      4,
		Name:    "Ethereum",
		ChainID: intPtr(1),
		RpcURL:  "https://eth.llamarpc.com",
	},
	Ethereum_Sepolia: {
		ID:      5,
		Name:    "Ethereum Sepolia",
		ChainID: intPtr(11155111),
		RpcURL:  "https://gateway.tenderly.co/public/sepolia",
	},
	Avalanche_C_CHAIN: {
		ID:      6,
		Name:    "Avalanche C-Chain",
		ChainID: intPtr(43114),
		RpcURL:  "https://avalanche-c-chain-rpc.publicnode.com	",
	},
	Avalanche_C_CHAIN_Fuji: {
		ID:      7,
		Name:    "Ethereum Sepolia",
		ChainID: intPtr(43113),
		RpcURL:  "https://ava-testnet.public.blastapi.io/ext/bc/C/rpc",
	},
}
var Networks = []Network{
	Bitcoin, BitcoinTestNet, BitcoinRegTest, Ethereum, Ethereum_Sepolia,
}

func (n Network) String() string {
	return NetworkMetadata[n].Name
}

func (n Network) ChainID() *int64 {
	return NetworkMetadata[n].ChainID
}
func (n Network) ID() int32 {
	return NetworkMetadata[n].ID
}

func (n Network) RPC() string {
	return NetworkMetadata[n].RpcURL
}

func GetNetworkByChainID(chainID int64) (Network, bool) {
	for network, metadata := range NetworkMetadata {
		if metadata.ChainID != nil && *metadata.ChainID == chainID {
			return network, true
		}
	}
	return 0, false
}

func intPtr(i int64) *int64 {
	return &i
}
