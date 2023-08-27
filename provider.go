package mdi

import (
	"reflect"
	"sync"
)

// provideMap represents a map from type to it's provider
type provideMap map[reflect.Type]*provider

// invoker represents function needed to get (invoke) dependency
type invoker func(*provider, *DI) (reflect.Value, error)

// newProviderFromOptions creates a new provider applying all options
func newProviderFromOptions(options []ProviderOption) *provider {
	p := &provider{}
	for _, option := range options {
		option(p)
	}
	return p
}

// provider represents one dependency provider
type provider struct {
	eagerLoading       bool
	disableCache       bool
	useRoundRobin      bool
	roundRobinIndex    int
	cache              reflect.Value
	invoker            invoker
	function           any
	functionParamIndex int
	mutex              sync.RWMutex
}

// setStrategyByValue sets by value strategy
func (p *provider) setStrategyByValue(pValue reflect.Value) *provider {
	p.cache = pValue
	p.invoker = func(iP *provider, di *DI) (reflect.Value, error) {
		return iP.cache, nil
	}
	return p
}

// setStrategyByValueRoundRobin sets by value strategy with round-robin
func (p *provider) setStrategyByValueRoundRobin(pValue reflect.Value) *provider {
	p.roundRobinIndex = -1
	p.cache = pValue
	p.invoker = func(iP *provider, di *DI) (reflect.Value, error) {
		iP.mutex.Lock()
		iP.roundRobinIndex++
		if iP.roundRobinIndex >= iP.cache.Len() {
			iP.roundRobinIndex = 0
		}
		iP.mutex.Unlock()
		return iP.cache.Index(iP.roundRobinIndex), nil
	}
	return p
}

// setStrategyByFunctionValue sets by function value strategy
func (p *provider) setStrategyByFunctionValue(function any, index int) *provider {
	p.function = function
	p.functionParamIndex = index
	p.invoker = func(iP *provider, di *DI) (reflect.Value, error) {
		result, iFunc := iP.getCacheOrFunction()
		if !result.IsValid() {
			results, err := di.invoke(iFunc)
			if err != nil {
				return result, err
			}
			result = results[iP.functionParamIndex]
			iP.setCache(result)
		}
		return result, nil
	}
	return p
}

// setStrategyByFunctionValueRoundRobin sets by function value strategy with round-robin
func (p *provider) setStrategyByFunctionValueRoundRobin(function any, index int) *provider {
	p.function = function
	p.functionParamIndex = index
	p.roundRobinIndex = -1
	p.invoker = func(iP *provider, di *DI) (reflect.Value, error) {
		result, iFunc := iP.getCacheOrFunction()
		if !result.IsValid() {
			results, err := di.invoke(iFunc)
			if err != nil {
				return result, err
			}
			result = results[iP.functionParamIndex]
			iP.setCache(result)
		}
		iP.mutex.Lock()
		iP.roundRobinIndex++
		if iP.roundRobinIndex >= result.Len() {
			iP.roundRobinIndex = 0
		}
		iP.mutex.Unlock()
		return result.Index(iP.roundRobinIndex), nil
	}
	return p
}

// getCacheOrFunction returns data from cache or function to invoke
func (p *provider) getCacheOrFunction() (reflect.Value, any) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	if p.disableCache {
		return reflect.Value{}, p.function
	}
	return p.cache, p.function
}

// setCache sets data into cache
func (p *provider) setCache(data reflect.Value) {
	if p.disableCache {
		return
	}
	p.mutex.Lock()
	p.cache = data
	p.function = nil
	p.mutex.Unlock()
}

// provide data using invoker
func (p *provider) provide(di *DI) (reflect.Value, error) {
	return p.invoker(p, di)
}
