package config

import (
	c "gitlink.org.cn/cloudream/utils/config"
	log "gitlink.org.cn/cloudream/utils/logger"
)

type Config struct {
	ID                int        `json:"id"`
	GRPCListenAddress string     `json:"grpcListenAddress"`
	LocalIP           string     `json:"localIP"`
	IPFSPort          int        `json:"ipfsPort"`
	StorageBaseDir    string     `json:"storageBaseDir"`
	Logger            log.Config `json:"logger"`
}

var cfg Config

func Init() error {
	return c.DefaultLoad(&cfg)
}

func Cfg() *Config {
	return &cfg
}
