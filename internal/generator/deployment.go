package generator

// DeploymentOptions holds parameters for generating a Deployment manifest.
type DeploymentOptions struct {
	Name      string
	Namespace string
	Image     string
	Port      int
	Replicas  int
}
