package aws

// AWSIntegration handles AWS EKS, S3, and IAM roles configuration.
type AWSIntegration struct {
	Region    string
	AccountID string
}

func NewAWSIntegration(region, accountID string) *AWSIntegration {
	if region == "" {
		region = "us-east-1"
	}
	return &AWSIntegration{
		Region:    region,
		AccountID: accountID,
	}
}
