package miner

import (
	"github.com/PlatONnetwork/PlatON-Go/metrics"
)

var (
	//readyBlockMeter = metrics.NewRegisteredCounter("miner/block/ready", nil)
	NewMinedBlockMeter         = metrics.NewRegisteredMeter("miner/block/mined", nil)
	MinedBlockTxCountHistogram = metrics.NewRegisteredHistogram("miner/block/txCountHistogram", nil)
)
