package types

// distribution module event types
const (
	EventTypeSetWithdrawAddress    = "set_withdraw_address"
	EventTypeRewards               = "rewards"
	EventTypeCommission            = "commission"
	EventTypeWithdrawRewards       = "withdraw_rewards"
	EventTypeWithdrawCommission    = "withdraw_commission"
	EventTypeProposerReward        = "proposer_reward"
	EventTypeTransferCollectedFees = "transfer_collected_fees"

	AttributeKeyWithdrawAddress = "withdraw_address"
	AttributeKeySenderModule    = "sender_module"
	AttributeKeyRecipientModule = "recipient_module"
	AttributeKeyValidator       = "validator"
	AttributeKeyDelegator       = "delegator"
)
