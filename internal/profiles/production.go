package profiles

// NewProductionProfile returns a default production-ready profile with hardened defaults.
func NewProductionProfile() *Profile {
	return &Profile{
		Name:        "production",
		Environment: "prod",
		Values:      make(map[string]any),
	}
}
