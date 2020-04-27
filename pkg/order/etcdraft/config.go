package etcdraft

import (
	"path/filepath"
	"time"

	"github.com/coreos/etcd/raft"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type RAFTConfig struct {
	RAFT RAFT
}

type TxPoolConfig struct {
	PackSize  int           `mapstructure:"pack_size"`
	BlockTick time.Duration `mapstructure:"block_tick"`
	PoolSize  int           `mapstructure:"pool_size"`
}

type RAFT struct {
	ElectionTick              int          `mapstructure:"election_tick"`
	HeartbeatTick             int          `mapstructure:"heartbeat_tick"`
	MaxSizePerMsg             uint64       `mapstructure:"max_size_per_msg"`
	MaxInflightMsgs           int          `mapstructure:"max_inflight_msgs"`
	CheckQuorum               bool         `mapstructure:"check_quorum"`
	PreVote                   bool         `mapstructure:"pre_vote"`
	DisableProposalForwarding bool         `mapstructure:"disable_proposal_forwarding"`
	TxPoolConfig              TxPoolConfig `mapstructure:"tx_pool"`
}

func defaultRaftConfig() raft.Config {
	return raft.Config{
		ElectionTick:              10,          //ElectionTick is the number of Node.Tick invocations that must pass between elections.(s)
		HeartbeatTick:             1,           //HeartbeatTick is the number of Node.Tick invocations that must pass between heartbeats.(s)
		MaxSizePerMsg:             1024 * 1024, //1024*1024, MaxSizePerMsg limits the max size of each append message.
		MaxInflightMsgs:           500,         //MaxInflightMsgs limits the max number of in-flight append messages during optimistic replication phase.
		PreVote:                   true,        // PreVote prevents reconnected node from disturbing network.
		CheckQuorum:               true,        // Leader steps down when quorum is not active for an electionTimeout.
		DisableProposalForwarding: true,        // This prevents blocks from being accidentally proposed by followers
	}
}

func defaultTxPoolConfig() TxPoolConfig {
	return TxPoolConfig{
		PackSize:  500,                    // How many transactions should the primary pack.
		BlockTick: 500 * time.Millisecond, //Block packaging time period.
		PoolSize:  50000,                  //How many transactions could the txPool stores in total.
	}
}

func generateRaftConfig(id uint64, repoRoot string, logger logrus.FieldLogger, ram MemoryStorage) (*raft.Config, error) {
	readConfig, err := readConfig(repoRoot)
	if err != nil {
		return &raft.Config{}, nil
	}
	defaultConfig := defaultRaftConfig()
	defaultConfig.ID = id
	defaultConfig.Storage = ram
	defaultConfig.Logger = logger
	if readConfig.RAFT.ElectionTick > 0 {
		defaultConfig.ElectionTick = readConfig.RAFT.ElectionTick
	}
	if readConfig.RAFT.HeartbeatTick > 0 {
		defaultConfig.HeartbeatTick = readConfig.RAFT.HeartbeatTick
	}
	if readConfig.RAFT.MaxSizePerMsg > 0 {
		defaultConfig.MaxSizePerMsg = readConfig.RAFT.MaxSizePerMsg
	}
	if readConfig.RAFT.MaxInflightMsgs > 0 {
		defaultConfig.MaxInflightMsgs = readConfig.RAFT.MaxInflightMsgs
	}
	defaultConfig.PreVote = readConfig.RAFT.PreVote
	defaultConfig.CheckQuorum = readConfig.RAFT.CheckQuorum
	defaultConfig.DisableProposalForwarding = readConfig.RAFT.DisableProposalForwarding
	return &defaultConfig, nil
}

func generateTxPoolConfig(repoRoot string) (*TxPoolConfig, error) {
	readConfig, err := readConfig(repoRoot)
	if err != nil {
		return &TxPoolConfig{}, nil
	}
	defaultTxPoolConfig := defaultTxPoolConfig()
	if readConfig.RAFT.TxPoolConfig.BlockTick > 0 {
		defaultTxPoolConfig.BlockTick = readConfig.RAFT.TxPoolConfig.BlockTick
	}
	if readConfig.RAFT.TxPoolConfig.PackSize > 0 {
		defaultTxPoolConfig.PackSize = readConfig.RAFT.TxPoolConfig.PackSize
	}
	if readConfig.RAFT.TxPoolConfig.PoolSize > 0 {
		defaultTxPoolConfig.PoolSize = readConfig.RAFT.TxPoolConfig.PoolSize
	}
	return &defaultTxPoolConfig, nil
}

func readConfig(repoRoot string) (*RAFTConfig, error) {
	v := viper.New()
	v.SetConfigFile(filepath.Join(repoRoot, "order.toml"))
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &RAFTConfig{}

	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
