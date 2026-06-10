package keeper

import (
	"context"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/stakingreward/types"
)

var _ types.MsgServer = msgServer{}

// msgServer is a wrapper of Keeper.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/stakingreward MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{Keeper: k}
}

// UpdateParams updates the module params. Since blocks_per_year or the reward
// denom may change, the cached per-slot reward is recomputed afterwards.
func (ms msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	if err := ms.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	if _, err := ms.RecomputePerSlotReward(ctx); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// FundRewardPool deposits native tokens into the reward pool and recomputes the
// per-slot reward from the new balance.
func (ms msgServer) FundRewardPool(ctx context.Context, msg *types.MsgFundRewardPool) (*types.MsgFundRewardPoolResponse, error) {
	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}

	if !msg.Amount.IsValid() || !msg.Amount.IsAllPositive() {
		return nil, errors.Wrapf(types.ErrInvalidAmount, "amount must be valid and positive: %s", msg.Amount)
	}

	params, err := ms.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Only the reward denom can be released from the pool; reject other denoms
	// so funds cannot get stranded in the module account.
	if len(msg.Amount) != 1 || msg.Amount[0].Denom != params.RewardDenom {
		return nil, errors.Wrapf(types.ErrInvalidAmount, "only the reward denom %q can be deposited, got %s", params.RewardDenom, msg.Amount)
	}

	if err := ms.bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleName, msg.Amount); err != nil {
		return nil, err
	}

	perSlot, err := ms.RecomputePerSlotReward(ctx)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFundRewardPool,
			sdk.NewAttribute(types.AttributeKeyDepositor, msg.Depositor),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyPerSlotReward, perSlot.String()),
		),
	)

	return &types.MsgFundRewardPoolResponse{}, nil
}
