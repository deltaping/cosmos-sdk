package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/stakingreward/types"
)

// InitGenesis initializes the x/stakingreward module state from genesis.
//
// NOTE: this must run after x/bank InitGenesis so the pool balance is available
// when recomputing the per-slot reward.
func (k Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}

	// ensure the module account exists
	k.authKeeper.GetModuleAccount(ctx, types.ModuleName)

	// If genesis carries an explicit per-slot reward, honor it; otherwise derive
	// it from the funded pool balance.
	if !data.PerSlotReward.IsNil() && data.PerSlotReward.IsPositive() {
		return k.SetPerSlotReward(ctx, data.PerSlotReward)
	}

	_, err := k.RecomputePerSlotReward(ctx)
	return err
}

// ExportGenesis returns the x/stakingreward module's genesis state.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	perSlot, err := k.GetPerSlotReward(ctx)
	if err != nil {
		return nil, err
	}

	return types.NewGenesisState(params, perSlot), nil
}
