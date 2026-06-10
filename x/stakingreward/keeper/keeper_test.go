package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/stakingreward/keeper"
	"github.com/cosmos/cosmos-sdk/x/stakingreward/types"
)

const testDenom = "stake"

// fakeAccountKeeper is a minimal AccountKeeper for tests.
type fakeAccountKeeper struct{}

func (fakeAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return authtypes.NewModuleAddress(name)
}

func (fakeAccountKeeper) GetModuleAccount(context.Context, string) sdk.ModuleAccountI {
	return nil
}

// fakeBankKeeper is an in-memory BankKeeper tracking balances per address.
type fakeBankKeeper struct {
	balances map[string]sdk.Coins
}

func newFakeBankKeeper() *fakeBankKeeper {
	return &fakeBankKeeper{balances: map[string]sdk.Coins{}}
}

func (b *fakeBankKeeper) setBalance(addr sdk.AccAddress, coins sdk.Coins) {
	b.balances[addr.String()] = coins
}

func (b *fakeBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, b.balances[addr.String()].AmountOf(denom))
}

func (b *fakeBankKeeper) SendCoinsFromModuleToModule(_ context.Context, sender, recipient string, amt sdk.Coins) error {
	return b.move(authtypes.NewModuleAddress(sender), authtypes.NewModuleAddress(recipient), amt)
}

func (b *fakeBankKeeper) SendCoinsFromAccountToModule(_ context.Context, sender sdk.AccAddress, recipient string, amt sdk.Coins) error {
	return b.move(sender, authtypes.NewModuleAddress(recipient), amt)
}

func (b *fakeBankKeeper) move(from, to sdk.AccAddress, amt sdk.Coins) error {
	fromBal := b.balances[from.String()]
	newFrom, neg := fromBal.SafeSub(amt...)
	if neg {
		return fmt.Errorf("insufficient funds: %s < %s", fromBal, amt)
	}
	b.balances[from.String()] = newFrom
	b.balances[to.String()] = b.balances[to.String()].Add(amt...)
	return nil
}

type testFixture struct {
	ctx       sdk.Context
	keeper    keeper.Keeper
	msgServer types.MsgServer
	bank      *fakeBankKeeper
	poolAddr  sdk.AccAddress
	feeAddr   sdk.AccAddress
	authority string
}

func setupKeeper(t *testing.T, blocksPerYear uint64, enabled bool, poolBalance math.Int) testFixture {
	t.Helper()

	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	bank := newFakeBankKeeper()
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	k := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		fakeAccountKeeper{},
		bank,
		authtypes.FeeCollectorName,
		authority,
	)

	poolAddr := authtypes.NewModuleAddress(types.ModuleName)
	feeAddr := authtypes.NewModuleAddress(authtypes.FeeCollectorName)
	bank.setBalance(poolAddr, sdk.NewCoins(sdk.NewCoin(testDenom, poolBalance)))

	require.NoError(t, k.SetParams(testCtx.Ctx, types.NewParams(testDenom, blocksPerYear, enabled)))

	return testFixture{
		ctx:       testCtx.Ctx,
		keeper:    k,
		msgServer: keeper.NewMsgServerImpl(k),
		bank:      bank,
		poolAddr:  poolAddr,
		feeAddr:   feeAddr,
		authority: authority,
	}
}

func TestRecomputePerSlotReward(t *testing.T) {
	// 1000 / 100 = 10 (floor)
	f := setupKeeper(t, 100, true, math.NewInt(1000))
	perSlot, err := f.keeper.RecomputePerSlotReward(f.ctx)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(10), perSlot)

	// floor: 1050 / 100 = 10
	f.bank.setBalance(f.poolAddr, sdk.NewCoins(sdk.NewCoin(testDenom, math.NewInt(1050))))
	perSlot, err = f.keeper.RecomputePerSlotReward(f.ctx)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(10), perSlot)

	// zero balance → zero
	f.bank.setBalance(f.poolAddr, sdk.NewCoins())
	perSlot, err = f.keeper.RecomputePerSlotReward(f.ctx)
	require.NoError(t, err)
	require.True(t, perSlot.IsZero())
}

func TestBeginBlockerReleasesToFeeCollector(t *testing.T) {
	f := setupKeeper(t, 100, true, math.NewInt(1000))
	_, err := f.keeper.RecomputePerSlotReward(f.ctx) // perSlot = 10
	require.NoError(t, err)

	require.NoError(t, f.keeper.BeginBlocker(f.ctx))

	require.Equal(t, math.NewInt(990), f.bank.GetBalance(f.ctx, f.poolAddr, testDenom).Amount)
	require.Equal(t, math.NewInt(10), f.bank.GetBalance(f.ctx, f.feeAddr, testDenom).Amount)

	// release event emitted
	require.True(t, hasEvent(f.ctx, types.EventTypeRelease))
}

func TestBeginBlockerLinearReleaseConstant(t *testing.T) {
	f := setupKeeper(t, 100, true, math.NewInt(1000))
	_, err := f.keeper.RecomputePerSlotReward(f.ctx) // perSlot = 10, stays constant
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		require.NoError(t, f.keeper.BeginBlocker(f.ctx))
	}
	// 5 blocks * 10 = 50 released; perSlot never recomputed between blocks
	require.Equal(t, math.NewInt(950), f.bank.GetBalance(f.ctx, f.poolAddr, testDenom).Amount)
	require.Equal(t, math.NewInt(50), f.bank.GetBalance(f.ctx, f.feeAddr, testDenom).Amount)

	perSlot, err := f.keeper.GetPerSlotReward(f.ctx)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(10), perSlot)
}

func TestBeginBlockerDisabled(t *testing.T) {
	f := setupKeeper(t, 100, false, math.NewInt(1000))
	_, err := f.keeper.RecomputePerSlotReward(f.ctx)
	require.NoError(t, err)

	require.NoError(t, f.keeper.BeginBlocker(f.ctx))
	require.Equal(t, math.NewInt(1000), f.bank.GetBalance(f.ctx, f.poolAddr, testDenom).Amount)
}

func TestBeginBlockerStopsWhenInsufficient(t *testing.T) {
	// balance 5, perSlot computed from 5/1 = 5; then drain so balance < perSlot
	f := setupKeeper(t, 1, true, math.NewInt(15))
	_, err := f.keeper.RecomputePerSlotReward(f.ctx) // perSlot = 15
	require.NoError(t, err)

	// first block releases 15, balance -> 0
	require.NoError(t, f.keeper.BeginBlocker(f.ctx))
	require.Equal(t, math.ZeroInt(), f.bank.GetBalance(f.ctx, f.poolAddr, testDenom).Amount)

	// second block: balance 0 < perSlot 15 -> stopped, no transfer
	f.ctx = f.ctx.WithEventManager(sdk.NewEventManager())
	require.NoError(t, f.keeper.BeginBlocker(f.ctx))
	require.Equal(t, math.NewInt(15), f.bank.GetBalance(f.ctx, f.feeAddr, testDenom).Amount)
	require.True(t, hasEvent(f.ctx, types.EventTypeReleaseStopped))
}

func TestFundRewardPoolRecomputes(t *testing.T) {
	f := setupKeeper(t, 100, true, math.NewInt(0))
	_, err := f.keeper.RecomputePerSlotReward(f.ctx) // perSlot = 0
	require.NoError(t, err)

	depositor := sdk.AccAddress([]byte("depositor___________"))
	f.bank.setBalance(depositor, sdk.NewCoins(sdk.NewCoin(testDenom, math.NewInt(2000))))

	_, err = f.msgServer.FundRewardPool(f.ctx, &types.MsgFundRewardPool{
		Depositor: depositor.String(),
		Amount:    sdk.NewCoins(sdk.NewCoin(testDenom, math.NewInt(2000))),
	})
	require.NoError(t, err)

	require.Equal(t, math.NewInt(2000), f.bank.GetBalance(f.ctx, f.poolAddr, testDenom).Amount)
	perSlot, err := f.keeper.GetPerSlotReward(f.ctx)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(20), perSlot) // 2000 / 100
	require.True(t, hasEvent(f.ctx, types.EventTypeFundRewardPool))
}

func TestFundRewardPoolRejectsWrongDenom(t *testing.T) {
	f := setupKeeper(t, 100, true, math.NewInt(0))
	depositor := sdk.AccAddress([]byte("depositor___________"))
	f.bank.setBalance(depositor, sdk.NewCoins(sdk.NewCoin("other", math.NewInt(2000))))

	_, err := f.msgServer.FundRewardPool(f.ctx, &types.MsgFundRewardPool{
		Depositor: depositor.String(),
		Amount:    sdk.NewCoins(sdk.NewCoin("other", math.NewInt(2000))),
	})
	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestUpdateParamsAuthorityGate(t *testing.T) {
	f := setupKeeper(t, 100, true, math.NewInt(1000))

	// wrong authority rejected
	_, err := f.msgServer.UpdateParams(f.ctx, &types.MsgUpdateParams{
		Authority: "cosmos1wrongauthority",
		Params:    types.NewParams(testDenom, 200, true),
	})
	require.Error(t, err)

	// correct authority updates params and recomputes perSlot (1000/200 = 5)
	_, err = f.msgServer.UpdateParams(f.ctx, &types.MsgUpdateParams{
		Authority: f.authority,
		Params:    types.NewParams(testDenom, 200, true),
	})
	require.NoError(t, err)

	perSlot, err := f.keeper.GetPerSlotReward(f.ctx)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(5), perSlot)
}

func TestQueryPool(t *testing.T) {
	f := setupKeeper(t, 100, true, math.NewInt(1000))
	_, err := f.keeper.RecomputePerSlotReward(f.ctx) // perSlot = 10
	require.NoError(t, err)

	qs := keeper.NewQueryServerImpl(f.keeper)
	resp, err := qs.Pool(f.ctx, &types.QueryPoolRequest{})
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1000), resp.Balance.Amount)
	require.Equal(t, math.NewInt(10), resp.PerSlotReward)
	require.Equal(t, uint64(100), resp.EstimatedBlocksRemaining) // 1000 / 10
}

func hasEvent(ctx sdk.Context, eventType string) bool {
	for _, e := range ctx.EventManager().Events() {
		if e.Type == eventType {
			return true
		}
	}
	return false
}
