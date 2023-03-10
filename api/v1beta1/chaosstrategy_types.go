package v1beta1

type ChaosStrategyType string

const (
	ChaosStrategyTypeRandom     ChaosStrategyType = "random"
	ChaosStrategyTypeRoundRobin ChaosStrategyType = "round-robin"
	ChaosStrategyTypeFixed      ChaosStrategyType = "targeted"
)

type ChaosStrategy struct {
	ChaosStrategyType ChaosStrategyType `json:"type"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	NumberOfPodsToKill int `json:"numberOfPodsToKill"`
	// +kubebuilder:default=30
	FrequencySeconds int64 `json:"frequencySec"`
}

func (c *ChaosStrategy) Default() {
	if c.NumberOfPodsToKill == 0 {
		c.NumberOfPodsToKill = 1
	}
	if c.FrequencySeconds == -1 {
		c.FrequencySeconds = 30
	}
}

type RoundRobinStrategy struct {
	// +kubebuilder:validation:Optional
	TargetPodsToKill []TargetKill `json:"targetPodsToKill,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	CurrentTargetId string `json:"currentTarget"`
}

type FixedStrategy struct {
	// +kubebuilder:validation:Optional
	TargetPodsToKill []TargetKill `json:"targetPodsToKill,omitempty"`
}

type TargetKill struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TargetId string `json:"id"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	Number int `json:"number"`
}
