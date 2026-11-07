package opentofu

// OpenTofuIntegration manages OpenTofu CLI executions.
type OpenTofuIntegration struct {
	BinaryPath string
}

func NewOpenTofuIntegration() *OpenTofuIntegration {
	return &OpenTofuIntegration{
		BinaryPath: "tofu",
	}
}
