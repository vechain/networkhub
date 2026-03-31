package genesis

type Config struct {
	BlockInterval              uint64  `json:"blockInterval"`
	EpochLength                uint32  `json:"epochLength"`
	SeederInterval             uint32  `json:"seederInterval"`
	ValidatorEvictionThreshold uint32  `json:"validatorEvictionThreshold"`
	EvictionCheckInterval      uint32  `json:"evictionCheckInterval"`
	LowStakingPeriod           uint32  `json:"lowStakingPeriod"`
	MediumStakingPeriod        uint32  `json:"mediumStakingPeriod"`
	HighStakingPeriod          uint32  `json:"highStakingPeriod"`
	CooldownPeriod             uint32  `json:"cooldownPeriod"`
	HayabusaTP                 *uint32 `json:"hayabusaTP"`
}
