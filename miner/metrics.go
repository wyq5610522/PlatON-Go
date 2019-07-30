package miner

import (
	"github.com/PlatONnetwork/PlatON-Go/metrics"
)

var (
	//readyBlockMeter = metrics.NewRegisteredCounter("miner/block/ready", nil)
	minedBlockCount    = metrics.NewRegisteredCounter("miner/block/mined", nil)
	minedBlockTxsCount = metrics.NewRegisteredCounter("miner/block/transactions", nil)
)
