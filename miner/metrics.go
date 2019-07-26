package miner

import (
	"github.com/PlatONnetwork/PlatON-Go/metrics"
)

var (
	readyBlockMeter = metrics.NewRegisteredCounter("miner/block/ready", nil)
	newBlockMeter   = metrics.NewRegisteredCounter("miner/block/new", nil)
)
