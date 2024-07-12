package config

import "sync"

var (
	once     sync.Once
	instance *Config
)

type Config struct {
	DBHost           string
	DBUser           string
	DBPassword       string
	DBName           string
	ServerPort       string
	BobGRPCAddress   string
	AliceGRPCAddress string
}

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
	})
	return instance
}

func SetAddresses(bobAddress, aliceAddress string) {
	config := GetConfig()
	config.BobGRPCAddress = bobAddress
	config.AliceGRPCAddress = aliceAddress
}
