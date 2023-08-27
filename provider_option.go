package mdi

// ProviderOption represents provider options
type ProviderOption func(p *provider)

// WithEagerLoading provider's option to eager load dependency even if not used
func WithEagerLoading() ProviderOption {
	return func(p *provider) {
		p.eagerLoading = true
	}
}

// WithMultiInstance provider's option to use multiple instances (without caching)
func WithMultiInstance() ProviderOption {
	return func(p *provider) {
		p.disableCache = true
	}
}

// WithRoundRobin provider's option for round-robin dependency
func WithRoundRobin() ProviderOption {
	return func(p *provider) {
		p.useRoundRobin = true
	}
}
