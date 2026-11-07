package provider

type CloudProvider interface {
	Name() string
	Region() string
}

type GenericProvider struct {
	ProviderName string
	RegionName   string
}

func (p *GenericProvider) Name() string   { return p.ProviderName }
func (p *GenericProvider) Region() string { return p.RegionName }
