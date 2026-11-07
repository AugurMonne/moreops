package generator

// IngressOptions holds parameters for generating an Ingress manifest.
type IngressOptions struct {
	Name      string
	Namespace string
	Host      string
	Path      string
	Service   string
	Port      int
}
