package profiles

// Profile represents environment configuration profile (e.g. dev, prod).
type Profile struct {
	Name        string
	Environment string
	Values      map[string]any
}
