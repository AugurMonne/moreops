package generator

// ConfigMapOptions holds parameters for generating a ConfigMap manifest.
type ConfigMapOptions struct {
	Name      string
	Namespace string
	Data      map[string]string
}
