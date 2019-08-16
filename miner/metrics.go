package miner

import (
	"github.com/PlatONnetwork/PlatON-Go/metrics"
)

var (
	minedBlockCount    = metrics.NewRegisteredCounter("miner/block/mined", nil)
	minedBlockTxsCount = metrics.NewRegisteredCounter("miner/block/transactions", nil)
)
