package mdi

import (
	"fmt"
	"reflect"
	"sync"
)

// New creates [DI] container
func New() *DI {
	return NewFrom(nil)
}

// NewFrom creates new [DI] container with parent (base) container
func NewFrom(parent *DI) *DI {
	di := &DI{
		parent:       parent,
		provide:      provideMap{},
		provideMutex: sync.RWMutex{},
	}
	return di.MustProvide(di)
}

// DI represents dependency container
type DI struct {
	parent       *DI
	provide      provideMap
	provideMutex sync.RWMutex
}

// Provide adds provider to container or returns error if the value can't be represented as provider
func (d *DI) Provide(provide any, options ...ProviderOption) error {
	pValue := reflect.ValueOf(provide)
	if pValue.Kind() == reflect.Func {
		return d.provideFunction(provide, options)
	}
	return d.provideValue(pValue, options)
}

// MustProvide is like [DI.Provide], but panics if error occurs
func (d *DI) MustProvide(value any, options ...ProviderOption) *DI {
	if err := d.Provide(value, options...); err != nil {
		panic(err)
	}
	return d
}

// Invoke calls functions with dependencies provided from the container
func (d *DI) Invoke(functions ...any) error {
	for _, function := range functions {
		if _, err := d.invoke(function); err != nil {
			return err
		}
	}
	return nil
}

// MustInvoke is like [DI.Invoke], but panics if error occurs
func (d *DI) MustInvoke(functions ...any) *DI {
	if err := d.Invoke(functions...); err != nil {
		panic(err)
	}
	return d
}

// addProvider adds a provider by type to container
func (d *DI) addProvider(pType reflect.Type, p *provider) error {
	d.provideMutex.Lock()

	if _, ok := d.provide[pType]; ok {
		d.provideMutex.Unlock()
		return newErrorProviderAlreadyExists(pType)
	}

	d.provide[pType] = p
	d.provideMutex.Unlock()
	return nil
}

// getProvider returns provider by type from container
func (d *DI) getProvider(pType reflect.Type) (*provider, bool) {
	d.provideMutex.RLock()
	p, ok := d.provide[pType]
	d.provideMutex.RUnlock()
	return p, ok
}

// canAddProvider check if provider can be added
func (d *DI) canAddProvider(pType reflect.Type) (bool, error) {
	if isTypeErr(pType) {
		return false, nil
	}
	if _, ok := d.getProvider(pType); ok {
		return false, newErrorProviderAlreadyExists(pType)
	}
	return true, nil
}

// provideValue adds value provider to container
func (d *DI) provideValue(pValue reflect.Value, options []ProviderOption) error {
	pType := pValue.Type()
	if ok, err := d.canAddProvider(pType); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("can't provide value of type %q", pType.String())
	}

	var err error
	p := newProviderFromOptions(options)
	if p.useRoundRobin {
		if eType, ok := elementType(pType); ok {
			err = d.addProvider(eType, p.setStrategyByValueRoundRobin(pValue))
		} else {
			err = newErrorProviderCantRoundRobin(pType)
		}
	} else {
		err = d.addProvider(pType, p.setStrategyByValue(pValue))
	}

	return err
}

// provideFunction adds function provider to container
func (d *DI) provideFunction(function any, options []ProviderOption) error {
	vType := reflect.TypeOf(function)

	provided := false
	for i := 0; i < vType.NumOut(); i++ {
		if err := d.provideFunctionValue(function, vType.Out(i), i, options); err != nil {
			return err
		}
		provided = true
	}
	if !provided {
		return fmt.Errorf("can't declare func provider without return values")
	}

	return nil
}

// provideFunctionValue adds function value provider to container
func (d *DI) provideFunctionValue(function any, pType reflect.Type, index int, options []ProviderOption) error {
	if ok, err := d.canAddProvider(pType); err != nil {
		return err
	} else if !ok {
		return nil
	}

	var err error
	p := newProviderFromOptions(options)
	if p.useRoundRobin {
		if eType, ok := elementType(pType); ok {
			err = d.addProvider(eType, p.setStrategyByFunctionValueRoundRobin(function, index))
		} else {
			err = newErrorProviderCantRoundRobin(pType)
		}
	} else {
		err = d.addProvider(pType, p.setStrategyByFunctionValue(function, index))
	}
	if err != nil {
		return err
	}

	if p.eagerLoading {
		if _, err = p.provide(d); err != nil {
			return fmt.Errorf("failed to eagerly load value of type %q: %w", pType, err)
		}
		if p.useRoundRobin {
			p.roundRobinIndex--
		}
	}

	return nil
}

// invoke calls function (or [reflect.Value] of kind [reflect.Func]) with dependencies provided from the container
func (d *DI) invoke(function any) ([]reflect.Value, error) {
	var fType reflect.Type
	vType, ok := function.(reflect.Value)
	if ok && vType.IsValid() {
		fType = vType.Type()
	} else {
		fType = reflect.TypeOf(function)
		vType = reflect.ValueOf(function)
	}

	if fType == nil || fType.Kind() != reflect.Func {
		return nil, fmt.Errorf("can't invoke a non-function value")
	}

	paramValues := make([]reflect.Value, 0, fType.NumIn())
	for i := 0; i < fType.NumIn(); i++ {
		paramValue, err := d.invokeParam(fType.In(i), i)
		if err != nil {
			return nil, err
		}
		paramValues = append(paramValues, paramValue)
	}

	return functionCall(vType, paramValues)
}

// invokeParam get one dependency from container
func (d *DI) invokeParam(param reflect.Type, i int) (reflect.Value, error) {
	p, ok := d.getProvider(param)
	if !ok {
		if d.parent != nil {
			return d.parent.invokeParam(param, i)
		}
		return reflect.Value{}, fmt.Errorf("not found provider for %d parameter of type %q",
			i+1, param.String())
	}

	paramValue, err := p.provide(d)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to provide %d parameter of type %q: %w",
			i+1, param.String(), err)
	}

	return paramValue, nil
}

// newErrorProviderAlreadyExists returns an error indicating that the provider of this type already exists
func newErrorProviderAlreadyExists(pType reflect.Type) error {
	return fmt.Errorf("provider of type %q already exists", pType.String())
}

// newErrorProviderCantRoundRobin returns an error indicating that the provider of this type is not suitable for
// round-robin
func newErrorProviderCantRoundRobin(pType reflect.Type) error {
	return fmt.Errorf("can't round-robin value of type %q, must be a slice or an array", pType.String())
}

// elementType returns type of element if the type is (pointer to) slice or array
func elementType(vType reflect.Type) (reflect.Type, bool) {
	checkType := vType
	if checkType.Kind() == reflect.Ptr {
		checkType = checkType.Elem()
	}
	if checkType.Kind() == reflect.Slice || checkType.Kind() == reflect.Array {
		checkType = checkType.Elem()
		return checkType, true
	}
	return checkType, false
}

// functionCall call a user's function
func functionCall(fValue reflect.Value, params []reflect.Value) ([]reflect.Value, error) {
	results := fValue.Call(params)
	for _, result := range results {
		if isTypeErr(result.Type()) {
			if err, ok := result.Interface().(error); ok {
				return nil, err
			}
		}
	}
	return results, nil
}

// isTypeErr checks if the type is built-in error
func isTypeErr(vType reflect.Type) bool {
	return vType.String() == "error"
}
