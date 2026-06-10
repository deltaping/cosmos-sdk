package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
)

// GetPerSlotReward returns the cached per-block release amount. When it has
// never been set it returns zero.
func (k Keeper) GetPerSlotReward(ctx context.Context) (math.Int, error) {
	amount, err := k.PerSlotReward.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return math.ZeroInt(), nil
		}
		return math.Int{}, err
	}
	return amount, nil
}

// SetPerSlotReward stores the per-block release amount.
func (k Keeper) SetPerSlotReward(ctx context.Context, amount math.Int) error {
	return k.PerSlotReward.Set(ctx, amount)
}

// RecomputePerSlotReward recomputes and stores the per-block release amount from
// the current pool balance:
//
//	per_slot_reward = floor(pool_balance / blocks_per_year)
//
// It must be called whenever the pool balance changes (funding/genesis) or the
// params change. It is intentionally NOT called every block, so the per-block
// release stays constant between fundings (linear release).
func (k Keeper) RecomputePerSlotReward(ctx context.Context) (math.Int, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return math.Int{}, err
	}

	balance := k.bankKeeper.GetBalance(ctx, k.GetPoolAddress(), params.RewardDenom)

	// blocks_per_year is guaranteed positive by params validation; integer
	// division truncates toward zero (floor for non-negative balances).
	perSlot := balance.Amount.Quo(math.NewIntFromUint64(params.BlocksPerYear))

	if err := k.SetPerSlotReward(ctx, perSlot); err != nil {
		return math.Int{}, err
	}
	return perSlot, nil
}
