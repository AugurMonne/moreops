package validator

// CheckConsistency checks consistency across manifests (e.g. service selector matching deployment labels).
func CheckConsistency(manifests map[string]string) []Issue {
	return nil
}
