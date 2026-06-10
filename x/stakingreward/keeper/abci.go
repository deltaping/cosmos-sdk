package keeper

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stakingreward/types"
)

// BeginBlocker releases the cached per-slot reward from the pool to the fee
// collector, where it is merged with gas fees and distributed by x/distribution.
//
// It must run before x/distribution's BeginBlocker so the released reward is
// allocated within the same block.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	if !params.Enabled {
		return nil
	}

	perSlot, err := k.GetPerSlotReward(ctx)
	if err != nil {
		return err
	}
	// Nothing configured to release (e.g. pool never funded).
	if !perSlot.IsPositive() {
		return nil
	}

	poolAddr := k.GetPoolAddress()
	balance := k.bankKeeper.GetBalance(ctx, poolAddr, params.RewardDenom)

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Requirement: if the pool cannot cover a full per-slot reward, stop
	// distributing (no partial payouts) and emit an alert event.
	if balance.Amount.LT(perSlot) {
		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeReleaseStopped,
				sdk.NewAttribute(types.AttributeKeyPerSlotReward, perSlot.String()),
				sdk.NewAttribute(types.AttributeKeyRemaining, balance.Amount.String()),
			),
		)
		return nil
	}

	coins := sdk.NewCoins(sdk.NewCoin(params.RewardDenom, perSlot))
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, coins); err != nil {
		return err
	}

	remaining := balance.Amount.Sub(perSlot)

	if perSlot.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(perSlot.Int64()), "released_tokens")
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRelease,
			sdk.NewAttribute(sdk.AttributeKeyAmount, coins.String()),
			sdk.NewAttribute(types.AttributeKeyPerSlotReward, perSlot.String()),
			sdk.NewAttribute(types.AttributeKeyRemaining, remaining.String()),
		),
	)

	return nil
}
