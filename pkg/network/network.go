package network

type Network int

const (
	Bitcoin Network = iota
	BitcoinTestNet
	Ethereum
	Ethereum_Sepolia
)

type NetworkMetadata struct {
	ID      int32
	Name    string
	ChainID *int64
	RpcURL  string
	// 필요한 다른 메타데이터 필드를 여기에 추가할 수 있습니다.
}

var networkMetadata = map[Network]NetworkMetadata{
	Bitcoin: {
		ID:      1,
		Name:    "Bitcoin",
		ChainID: nil,
		RpcURL:  "", // Bitcoin doesn't use RPC for this purpose
	},
	BitcoinTestNet: {
		ID:      2,
		Name:    "Bitcoin Testnet",
		ChainID: nil,
		RpcURL:  "", // Bitcoin testnet doesn't use RPC for this purpose
	},
	Ethereum: {
		ID:      3,
		Name:    "Ethereum",
		ChainID: intPtr(1),
		RpcURL:  "https://eth.llamarpc.com",
	},
	Ethereum_Sepolia: {
		ID:      4,
		Name:    "Ethereum Sepolia",
		ChainID: intPtr(11155111),
		RpcURL:  "https://gateway.tenderly.co/public/sepolia",
	},
}
var Networks = []Network{
	Bitcoin, BitcoinTestNet, Ethereum, Ethereum_Sepolia,
}

func (n Network) String() string {
	return networkMetadata[n].Name
}

func (n Network) ChainID() *int64 {
	return networkMetadata[n].ChainID
}
func (n Network) ID() int32 {
	return networkMetadata[n].ID
}

func (n Network) RPC() string {
	return networkMetadata[n].RpcURL
}

func GetNetworkByChainID(chainID int64) (Network, bool) {
	for network, metadata := range networkMetadata {
		if metadata.ChainID != nil && *metadata.ChainID == chainID {
			return network, true
		}
	}
	return 0, false
}

func intPtr(i int64) *int64 {
	return &i
}
