package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stakingreward/types"
)

// Keeper of the x/stakingreward store.
type Keeper struct {
	cdc              codec.BinaryCodec
	storeService     storetypes.KVStoreService
	authKeeper       types.AccountKeeper
	bankKeeper       types.BankKeeper
	feeCollectorName string

	// authority is the address capable of executing a MsgUpdateParams message.
	// Typically this is the x/gov module account.
	authority string

	Schema collections.Schema
	// Params holds the module parameters.
	Params collections.Item[types.Params]
	// PerSlotReward caches the per-block release amount. It is recomputed only
	// when the pool balance changes (funding/genesis) or params are updated, so
	// the per-block release stays constant between fundings (linear release).
	PerSlotReward collections.Item[math.Int]
}

// NewKeeper creates a new x/stakingreward Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	feeCollectorName string,
	authority string,
) Keeper {
	// ensure the module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("the x/%s module account has not been set", types.ModuleName))
	}

	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid x/%s authority address: %w", types.ModuleName, err))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:              cdc,
		storeService:     storeService,
		authKeeper:       ak,
		bankKeeper:       bk,
		feeCollectorName: feeCollectorName,
		authority:        authority,
		Params:           collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		PerSlotReward:    collections.NewItem(sb, types.PerSlotRewardKey, "per_slot_reward", sdk.IntValue),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAuthority returns the x/stakingreward module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// GetPoolAddress returns the reward pool (module account) address.
func (k Keeper) GetPoolAddress() sdk.AccAddress {
	return k.authKeeper.GetModuleAddress(types.ModuleName)
}

// GetPoolBalance returns the current reward pool balance in the reward denom.
func (k Keeper) GetPoolBalance(ctx context.Context) (sdk.Coin, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}
	return k.bankKeeper.GetBalance(ctx, k.GetPoolAddress(), params.RewardDenom), nil
}
