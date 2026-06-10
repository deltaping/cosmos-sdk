package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/stakingreward/types"
)

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the x/stakingreward QueryServer interface.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// Params returns the module params.
func (q queryServer) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

// Pool returns the current pool balance, the cached per-slot reward and the
// estimated number of blocks the pool can keep paying.
func (q queryServer) Pool(ctx context.Context, _ *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	balance, err := q.k.GetPoolBalance(ctx)
	if err != nil {
		return nil, err
	}

	perSlot, err := q.k.GetPerSlotReward(ctx)
	if err != nil {
		return nil, err
	}

	var estimated uint64
	if perSlot.IsPositive() {
		estimated = balance.Amount.Quo(perSlot).Uint64()
	}

	return &types.QueryPoolResponse{
		Balance:                  balance,
		PerSlotReward:            perSlot,
		EstimatedBlocksRemaining: estimated,
	}, nil
}
