package generator

// ServiceOptions holds parameters for generating a Service manifest.
type ServiceOptions struct {
	Name       string
	Namespace  string
	Port       int
	TargetPort int
	Type       string
}
