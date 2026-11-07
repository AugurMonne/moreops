package profiles

// NewDevelopmentProfile returns a default development profile.
func NewDevelopmentProfile() *Profile {
	return &Profile{
		Name:        "development",
		Environment: "dev",
		Values:      make(map[string]any),
	}
}
