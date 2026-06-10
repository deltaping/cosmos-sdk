package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(params Params, perSlotReward math.Int) *GenesisState {
	return &GenesisState{
		Params:        params,
		PerSlotReward: perSlotReward,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		PerSlotReward: math.ZeroInt(),
	}
}

// ValidateGenesis validates the provided genesis state.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	if data.PerSlotReward.IsNil() {
		return fmt.Errorf("per slot reward cannot be nil")
	}
	if data.PerSlotReward.IsNegative() {
		return fmt.Errorf("per slot reward cannot be negative: %s", data.PerSlotReward)
	}
	return nil
}
