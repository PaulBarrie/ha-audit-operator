package v1beta1

type ChaosStrategyType string

const (
	ChaosStrategyTypeRandom ChaosStrategyType = "random"
	ChaosStrategyTypeFixed  ChaosStrategyType = "targeted"
)

type ChaosStrategy struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=random;targeted
	Type ChaosStrategyType `json:"type"`
	// +kubebuilder:default=1
	DurationSeconds     int                  `json:"durationSeconds"`
	RandomChaosStrategy *RandomChaosStrategy `json:"randomChaosStrategy,omitempty"`
}

type RandomChaosStrategy struct {
	// +kubebuilder:validation:Minimum=0
	NumberOfPodsToDelete int `json:"numberOfPodsToDelete"`
	// +kubebuilder:description="Cron expression to tell the frequency of pod deletion"
	// +kubebuilder:default="* * * * *"
	FrequencyCron string `json:"frequencyCron"`
}
