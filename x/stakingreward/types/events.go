package types

// staking reward module event types
const (
	// EventTypeRelease is emitted on every per-block release to the fee collector.
	EventTypeRelease = "staking_reward_release"
	// EventTypeReleaseStopped is emitted when a block cannot pay the per-slot
	// reward (pool exhausted) so the release is skipped.
	EventTypeReleaseStopped = "staking_reward_release_stopped"
	// EventTypeFundRewardPool is emitted when the reward pool is funded.
	EventTypeFundRewardPool = "fund_reward_pool"

	// AttributeKeyRemaining is the pool balance remaining after a release.
	AttributeKeyRemaining = "remaining"
	// AttributeKeyPerSlotReward is the cached per-block release amount.
	AttributeKeyPerSlotReward = "per_slot_reward"
	// AttributeKeyDepositor is the address that funded the reward pool.
	AttributeKeyDepositor = "depositor"
)
