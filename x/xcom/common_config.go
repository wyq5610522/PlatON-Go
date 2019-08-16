package xcom

import (
	"math/big"
	"sync"
)

var SecondsPerYear = uint64(365 * 24 * 3600)

// plugin rule key
const (
	DefualtRule = iota
	StakingRule
	SlashingRule
	RestrictingRule
	RewardRule
	GovernanceRule
)

type commonConfig struct {
	ExpectedMinutes uint64 // expected minutes every epoch
	Interval        uint64 // each block interval (uint: seconds)
	PerRoundBlocks  uint64 // blocks each validator will create per consensus epoch
	ValidatorCount  uint64 // The consensus validators count
}

type stakingConfig struct {
	StakeThreshold               *big.Int // The Staking minimum threshold allowed
	MinimumThreshold             *big.Int // The (incr, decr) delegate or incr staking minimum threshold allowed
	EpochValidatorNum            uint64   // The epoch (billing cycle) validators count
	ShiftValidatorNum            uint64   // The number of elections and replacements for each of the consensus rounds
	HesitateRatio                uint64   // Each hesitation period is a multiple of the epoch
	EffectiveRatio               uint64   // Each effective period is a multiple of the epoch
	ElectionDistance             uint64   // The interval of the last block of the high-distance consensus round of the election block for each consensus round
	UnStakeFreezeRatio           uint64   // The freeze period of the withdrew Staking (unit is  epochs)
	PassiveUnDelegateFreezeRatio uint64   // The freeze period of the delegate was invalidated due to the withdrawal of the Stake (unit is  epochs)
	ActiveUnDelegateFreezeRatio  uint64   // The freeze period of the delegate was invalidated due to active withdrew delegate (unit is  epochs)
}

type slashingConfig struct {
	PackAmountAbnormal        uint32 // The number of blocks packed per round, reaching this value is abnormal
	PackAmountHighAbnormal    uint32 // The number of blocks packed per round, reaching this value is a high degree of abnormality
	PackAmountLowSlashRate    uint32 // Proportion of deducted quality deposit (when the number of packing blocks is abnormal); 10% -> 10
	PackAmountHighSlashRate   uint32 // Proportion of quality deposits deducted (when the number of packing blocks is high degree of abnormality); 20% -> 20
	DuplicateSignNum          uint32 // Number of multiple signatures
	DuplicateSignLowSlashing  uint32 // Deduction ratio when the number of multi-signs is lower than DuplicateSignNum; 10% -> 10
	DuplicateSignHighSlashing uint32 // Deduction ratio when the number of multi-signs is higher than DuplicateSignNum; 20% -> 20
}

type governanceConfig struct {
	SupportRateThreshold float64
}

// total
type EconomicModel struct {
	Common   commonConfig
	Staking  stakingConfig
	Slashing slashingConfig
	Gov      governanceConfig
}

var (
	modelOnce sync.Once
	ec        *EconomicModel
)

// Getting the global EconomicModel single instance
func GetEc(netId int8) *EconomicModel {
	modelOnce.Do(func() {
		ec = getDefaultEMConfig(netId)
	})
	return ec
}

const (
	DefaultMainNet      = iota // PlatON default main net flag
	DefaultAlphaTestNet        // PlatON default Alpha test net flag
	DefaultBetaTestNet         // PlatON default Beta test net flag
	DefaultInnerTestNet        // PlatON default inner test net flag
	DefaultInnerDevNet         // PlatON default inner development net flag
	DefaultDeveloperNet        // PlatON default developer net flag
)

func getDefaultEMConfig(netId int8) *EconomicModel {
	var (
		success               bool
		stakeThresholdCount   string
		minimumThresholdCount string
		stakeThreshold        *big.Int
		minimumThreshold      *big.Int
	)

	switch netId {
	case DefaultMainNet, DefaultDeveloperNet:
		stakeThresholdCount = "10000000000000000000000000" // 1000W von
		minimumThresholdCount = "10000000000000000000"     // 10 von
	case DefaultAlphaTestNet:
		stakeThresholdCount = "10000000000000000000000000"
		minimumThresholdCount = "10000000000000000000"
	case DefaultBetaTestNet:
		stakeThresholdCount = "10000000000000000000000000"
		minimumThresholdCount = "10000000000000000000"
	case DefaultInnerTestNet:
		stakeThresholdCount = "10000000000000000000000000"
		minimumThresholdCount = "10000000000000000000"
	case DefaultInnerDevNet:
		stakeThresholdCount = "10000000000000000000000000"
		minimumThresholdCount = "10000000000000000000"
	}

	if stakeThreshold, success = new(big.Int).SetString(stakeThresholdCount, 10); !success {
		return nil
	}
	if minimumThreshold, success = new(big.Int).SetString(minimumThresholdCount, 10); !success {
		return nil
	}

	switch netId {
	case DefaultMainNet:
		ec = &EconomicModel{
			Common: commonConfig{
				ExpectedMinutes: uint64(360), // 6 hours
				Interval:        uint64(1),
				PerRoundBlocks:  uint64(10),
				ValidatorCount:  uint64(25),
			},
			Staking: stakingConfig{
				StakeThreshold:               stakeThreshold,
				MinimumThreshold:             minimumThreshold,
				EpochValidatorNum:            uint64(101),
				ShiftValidatorNum:            uint64(8),
				HesitateRatio:                uint64(1),
				EffectiveRatio:               uint64(1),
				ElectionDistance:             uint64(20),
				UnStakeFreezeRatio:           uint64(1),
				PassiveUnDelegateFreezeRatio: uint64(0),
				ActiveUnDelegateFreezeRatio:  uint64(0),
			},
			Slashing: slashingConfig{
				PackAmountAbnormal:        uint32(8),
				PackAmountHighAbnormal:    uint32(5),
				PackAmountLowSlashRate:    uint32(10),
				PackAmountHighSlashRate:   uint32(20),
				DuplicateSignNum:          uint32(2),
				DuplicateSignLowSlashing:  uint32(10),
				DuplicateSignHighSlashing: uint32(20),
			},
			Gov: governanceConfig{
				SupportRateThreshold: float64(0.85),
			},
		}

	case DefaultAlphaTestNet:
		ec = &EconomicModel{
			Common: commonConfig{
				ExpectedMinutes: uint64(10), // 10 minutes
				Interval:        uint64(1),
				PerRoundBlocks:  uint64(15),
				ValidatorCount:  uint64(4),
			},
			Staking: stakingConfig{
				StakeThreshold:               stakeThreshold,
				MinimumThreshold:             minimumThreshold,
				EpochValidatorNum:            uint64(21),
				ShiftValidatorNum:            uint64(1),
				HesitateRatio:                uint64(1),
				EffectiveRatio:               uint64(1),
				ElectionDistance:             uint64(10),
				UnStakeFreezeRatio:           uint64(1),
				PassiveUnDelegateFreezeRatio: uint64(0),
				ActiveUnDelegateFreezeRatio:  uint64(0),
			},
			Slashing: slashingConfig{
				PackAmountAbnormal:        uint32(8),
				PackAmountHighAbnormal:    uint32(5),
				PackAmountLowSlashRate:    uint32(10),
				PackAmountHighSlashRate:   uint32(20),
				DuplicateSignNum:          uint32(2),
				DuplicateSignLowSlashing:  uint32(10),
				DuplicateSignHighSlashing: uint32(20),
			},
			Gov: governanceConfig{
				SupportRateThreshold: float64(0.85),
			},
		}

	case DefaultBetaTestNet:
		ec = &EconomicModel{
			Common: commonConfig{
				ExpectedMinutes: uint64(10), // 10 minutes
				Interval:        uint64(1),
				PerRoundBlocks:  uint64(15),
				ValidatorCount:  uint64(4),
			},
			Staking: stakingConfig{
				StakeThreshold:               stakeThreshold,
				MinimumThreshold:             minimumThreshold,
				EpochValidatorNum:            uint64(21),
				ShiftValidatorNum:            uint64(1),
				HesitateRatio:                uint64(1),
				EffectiveRatio:               uint64(1),
				ElectionDistance:             uint64(10),
				UnStakeFreezeRatio:           uint64(1),
				PassiveUnDelegateFreezeRatio: uint64(0),
				ActiveUnDelegateFreezeRatio:  uint64(0),
			},
			Slashing: slashingConfig{
				PackAmountAbnormal:        uint32(8),
				PackAmountHighAbnormal:    uint32(5),
				PackAmountLowSlashRate:    uint32(10),
				PackAmountHighSlashRate:   uint32(20),
				DuplicateSignNum:          uint32(2),
				DuplicateSignLowSlashing:  uint32(10),
				DuplicateSignHighSlashing: uint32(20),
			},
			Gov: governanceConfig{
				SupportRateThreshold: float64(0.85),
			},
		}

	case DefaultInnerTestNet:
		ec = &EconomicModel{
			Common: commonConfig{
				ExpectedMinutes: uint64(666), // 11 hours
				Interval:        uint64(1),
				PerRoundBlocks:  uint64(25),
				ValidatorCount:  uint64(10),
			},
			Staking: stakingConfig{
				StakeThreshold:               stakeThreshold,
				MinimumThreshold:             minimumThreshold,
				EpochValidatorNum:            uint64(51),
				ShiftValidatorNum:            uint64(3),
				HesitateRatio:                uint64(1),
				EffectiveRatio:               uint64(1),
				ElectionDistance:             uint64(20),
				UnStakeFreezeRatio:           uint64(1),
				PassiveUnDelegateFreezeRatio: uint64(0),
				ActiveUnDelegateFreezeRatio:  uint64(0),
			},
			Slashing: slashingConfig{
				PackAmountAbnormal:        uint32(8),
				PackAmountHighAbnormal:    uint32(5),
				PackAmountLowSlashRate:    uint32(10),
				PackAmountHighSlashRate:   uint32(20),
				DuplicateSignNum:          uint32(2),
				DuplicateSignLowSlashing:  uint32(10),
				DuplicateSignHighSlashing: uint32(20),
			},
			Gov: governanceConfig{
				SupportRateThreshold: float64(0.85),
			},
		}

	case DefaultInnerDevNet:
		ec = &EconomicModel{
			Common: commonConfig{
				ExpectedMinutes: uint64(10), // 10 minutes
				Interval:        uint64(1),
				PerRoundBlocks:  uint64(15),
				ValidatorCount:  uint64(4),
			},
			Staking: stakingConfig{
				StakeThreshold:               stakeThreshold,
				MinimumThreshold:             minimumThreshold,
				EpochValidatorNum:            uint64(21),
				ShiftValidatorNum:            uint64(1),
				HesitateRatio:                uint64(1),
				EffectiveRatio:               uint64(1),
				ElectionDistance:             uint64(10),
				UnStakeFreezeRatio:           uint64(1),
				PassiveUnDelegateFreezeRatio: uint64(0),
				ActiveUnDelegateFreezeRatio:  uint64(0),
			},
			Slashing: slashingConfig{
				PackAmountAbnormal:        uint32(8),
				PackAmountHighAbnormal:    uint32(5),
				PackAmountLowSlashRate:    uint32(10),
				PackAmountHighSlashRate:   uint32(20),
				DuplicateSignNum:          uint32(2),
				DuplicateSignLowSlashing:  uint32(10),
				DuplicateSignHighSlashing: uint32(20),
			},
			Gov: governanceConfig{
				SupportRateThreshold: float64(0.85),
			},
		}

	default:
		// Default is inner develop net config
		ec = &EconomicModel{
			Common: commonConfig{
				ExpectedMinutes: uint64(10), // 10 minutes
				Interval:        uint64(1),
				PerRoundBlocks:  uint64(15),
				ValidatorCount:  uint64(4),
			},
			Staking: stakingConfig{
				StakeThreshold:               stakeThreshold,
				MinimumThreshold:             minimumThreshold,
				EpochValidatorNum:            uint64(21),
				ShiftValidatorNum:            uint64(1),
				HesitateRatio:                uint64(1),
				EffectiveRatio:               uint64(1),
				ElectionDistance:             uint64(10),
				UnStakeFreezeRatio:           uint64(1),
				PassiveUnDelegateFreezeRatio: uint64(0),
				ActiveUnDelegateFreezeRatio:  uint64(0),
			},
			Slashing: slashingConfig{
				PackAmountAbnormal:        uint32(8),
				PackAmountHighAbnormal:    uint32(5),
				PackAmountLowSlashRate:    uint32(10),
				PackAmountHighSlashRate:   uint32(20),
				DuplicateSignNum:          uint32(2),
				DuplicateSignLowSlashing:  uint32(10),
				DuplicateSignHighSlashing: uint32(20),
			},
			Gov: governanceConfig{
				SupportRateThreshold: float64(0.85),
			},
		}
	}

	return ec
}

/******
 * Common configure
 ******/
func ExpectedMinutes() uint64 {
	return ec.Common.ExpectedMinutes
}
func Interval() uint64 {
	return ec.Common.Interval
}
func BlocksWillCreate() uint64 {
	return ec.Common.PerRoundBlocks
}
func ConsValidatorNum() uint64 {
	return ec.Common.ValidatorCount
}

/******
 * Staking configure
 ******/
func StakeThreshold() *big.Int {
	return ec.Staking.StakeThreshold
}

func MinimumThreshold() *big.Int {
	return ec.Staking.MinimumThreshold
}

func EpochValidatorNum() uint64 {
	return ec.Staking.EpochValidatorNum
}

func ShiftValidatorNum() uint64 {
	return ec.Staking.ShiftValidatorNum
}

func HesitateRatio() uint64 {
	return ec.Staking.HesitateRatio
}

func EffectiveRatio() uint64 {
	return ec.Staking.EffectiveRatio
}

func ElectionDistance() uint64 {
	return ec.Staking.ElectionDistance
}

func UnStakeFreezeRatio() uint64 {
	return ec.Staking.UnStakeFreezeRatio
}

func PassiveUnDelFreezeRatio() uint64 {
	return ec.Staking.PassiveUnDelegateFreezeRatio
}

func ActiveUnDelFreezeRatio() uint64 {
	return ec.Staking.ActiveUnDelegateFreezeRatio
}

/******
 * Slashing config
 ******/
func PackAmountAbnormal() uint32 {
	return ec.Slashing.PackAmountAbnormal
}

func PackAmountHighAbnormal() uint32 {
	return ec.Slashing.PackAmountHighAbnormal
}

func PackAmountLowSlashRate() uint32 {
	return ec.Slashing.PackAmountLowSlashRate
}

func PackAmountHighSlashRate() uint32 {
	return ec.Slashing.PackAmountHighSlashRate
}

func DuplicateSignNum() uint32 {
	return ec.Slashing.DuplicateSignNum
}

func DuplicateSignLowSlash() uint32 {
	return ec.Slashing.DuplicateSignLowSlashing
}

func DuplicateSignHighSlash() uint32 {
	return ec.Slashing.DuplicateSignHighSlashing
}

/******
 * Reward config
 ******/

/******
 * Governance config
 ******/
func SupportRateThreshold() float64 {
	return ec.Gov.SupportRateThreshold
}
