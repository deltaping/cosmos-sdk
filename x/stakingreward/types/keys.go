package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the name of the staking reward module.
	ModuleName = "stakingreward"

	// StoreKey is the store key for the staking reward module.
	//
	// NOTE: it intentionally differs from ModuleName ("stakingreward"). The
	// multistore's assertNoCommonPrefix check rejects any store key that is a
	// prefix of another, and "staking" (x/staking's key) is a prefix of
	// "stakingreward". Using a distinct, non-prefixing key avoids that collision.
	StoreKey = "stakereward"

	// RouterKey is the message route for the staking reward module.
	RouterKey = ModuleName
)

var (
	// ParamsKey is the store key for the module params.
	ParamsKey = collections.NewPrefix(0)

	// PerSlotRewardKey is the store key for the cached per-block release amount.
	PerSlotRewardKey = collections.NewPrefix(1)
)
