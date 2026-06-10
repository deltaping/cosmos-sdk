package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/stakingreward/types"
)

// GetParams returns the current x/stakingreward module parameters.
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	return k.Params.Get(ctx)
}

// SetParams sets the x/stakingreward module parameters.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	return k.Params.Set(ctx, params)
}
