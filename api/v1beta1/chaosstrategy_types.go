package v1beta1

type ChaosStrategyType string

const (
	ChaosStrategyTypeRandom     ChaosStrategyType = "random"
	ChaosStrategyTypeRoundRobin ChaosStrategyType = "round-robin"
	ChaosStrategyTypeFixed      ChaosStrategyType = "targeted"
)

type ChaosStrategy struct {
	// +kubebuilder:default=3600
	DurationSeconds   int               `json:"durationSeconds"`
	ChaosStrategyType ChaosStrategyType `json:"chaosStrategyType"`
	// +kubebuilder:default=1
	NumberOfPodsToKill int `json:"numberOfPodsToKill"`
	// +optional
	RoundRobinStrategy RoundRobinStrategy `json:"roundRobinStrategy,omitempty"`
	// +optional
	FixedStrategy FixedStrategy `json:"fixedStrategy,omitempty"`
	// +kubebuilder:default=1
	FrequencyCron string `json:"frequencyCron"`
}

func (c *ChaosStrategy) Default() {
	if c.DurationSeconds == 0 {
		c.DurationSeconds = 3600
	}
	if c.NumberOfPodsToKill == 0 {
		c.NumberOfPodsToKill = 1
	}
	if c.FrequencyCron == "" {
		c.FrequencyCron = "1/* * * * *"
	}
}

type RoundRobinStrategy struct {
	// +optional
	TargetPodsToKill []TargetKill `json:"targetPodsToKill"`
	CurrentTargetId  string       `json:"currentTarget"`
}

func (r *RoundRobinStrategy) Default(targets []Target) {

}

type FixedStrategy struct {
	// +optional
	TargetPodsToKill []TargetKill `json:"targetPodsToKill"`
}

type TargetKill struct {
	TargetId string `json:"id"`
	// +kubebuilder:default=1
	Number int `json:"number"`
}
