package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams returns a Params instance with the given values.
func NewParams(rewardDenom string, blocksPerYear uint64, enabled bool) Params {
	return Params{
		RewardDenom:   rewardDenom,
		BlocksPerYear: blocksPerYear,
		Enabled:       enabled,
	}
}

// DefaultParams returns the default x/stakingreward module parameters.
// The module is disabled by default; it must be explicitly enabled and funded.
func DefaultParams() Params {
	return Params{
		RewardDenom:   sdk.DefaultBondDenom,
		BlocksPerYear: uint64(60 * 60 * 8766 / 5), // assuming 5 second block times
		Enabled:       false,
	}
}

// Validate performs sanity checks on the params.
func (p Params) Validate() error {
	if err := validateRewardDenom(p.RewardDenom); err != nil {
		return err
	}
	return validateBlocksPerYear(p.BlocksPerYear)
}

func validateRewardDenom(denom string) error {
	if strings.TrimSpace(denom) == "" {
		return errors.New("reward denom cannot be blank")
	}
	return sdk.ValidateDenom(denom)
}

func validateBlocksPerYear(blocksPerYear uint64) error {
	if blocksPerYear == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", blocksPerYear)
	}
	return nil
}
