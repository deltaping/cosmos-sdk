package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/x/stakingreward/types"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name    string
		params  types.Params
		expErr  bool
		errText string
	}{
		{
			name:   "default params are valid",
			params: types.DefaultParams(),
			expErr: false,
		},
		{
			name:   "valid custom params",
			params: types.NewParams("uatom", 6_311_520, true),
			expErr: false,
		},
		{
			name:    "blank denom",
			params:  types.NewParams("   ", 100, true),
			expErr:  true,
			errText: "reward denom cannot be blank",
		},
		{
			name:   "invalid denom",
			params: types.NewParams("1bad", 100, true),
			expErr: true,
		},
		{
			name:    "zero blocks per year",
			params:  types.NewParams("stake", 0, true),
			expErr:  true,
			errText: "blocks per year must be positive",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expErr {
				require.Error(t, err)
				if tc.errText != "" {
					require.Contains(t, err.Error(), tc.errText)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestGenesisValidate(t *testing.T) {
	require.NoError(t, types.ValidateGenesis(*types.DefaultGenesisState()))

	// negative per-slot reward is rejected
	bad := types.NewGenesisState(types.DefaultParams(), math.NewInt(-1))
	require.Error(t, types.ValidateGenesis(*bad))

	// nil per-slot reward is rejected
	nilGen := types.GenesisState{Params: types.DefaultParams(), PerSlotReward: math.Int{}}
	require.Error(t, types.ValidateGenesis(nilGen))

	// invalid params propagate
	badParams := types.NewGenesisState(types.NewParams("", 0, true), math.ZeroInt())
	require.Error(t, types.ValidateGenesis(*badParams))
}
