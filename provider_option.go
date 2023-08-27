package mdi

// ProviderOption represents provider options
type ProviderOption func(p *provider)

// WithMultiInstance provider's option multi instance (without cache)
func WithMultiInstance() ProviderOption {
	return func(p *provider) {
		p.canCache = false
	}
}

// WithRoundRobin provider's option for round-robin
func WithRoundRobin() ProviderOption {
	return func(p *provider) {
		p.canRoundRobin = true
	}
}
