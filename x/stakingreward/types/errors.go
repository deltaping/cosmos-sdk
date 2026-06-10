package types

import "cosmossdk.io/errors"

// x/stakingreward module sentinel errors.
var (
	ErrInvalidParams = errors.Register(ModuleName, 2, "invalid params")
	ErrInvalidAmount = errors.Register(ModuleName, 3, "invalid amount")
)
