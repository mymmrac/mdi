package mdi

import (
	"errors"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
)

var errTest = errors.New("test_err")

func TestDI_ProvideAndInvoke(t *testing.T) {
	tests := map[string]struct {
		provide         any
		provideErr      error
		providerOptions []ProviderOption
		invoke          any
		invokeErr       error
	}{
		"success_value_int": {
			provide: 1,
			invoke: func(i int) {
				if i != 1 {
					t.Fatalf("unexpected: %d", i)
				}
			},
		},
		"success_func_int": {
			provide: func() int { return 1 },
			invoke: func(i int) {
				if i != 1 {
					t.Fatalf("unexpected: %d", i)
				}
			},
		},
		"success_invoke_reflect": {
			provide: 1,
			invoke: reflect.ValueOf(func(i int) {
				if i != 1 {
					t.Fatalf("unexpected: %d", i)
				}
			}),
		},
		"success_value_int_round_robin": {
			provide:         []int{1, 2},
			providerOptions: []ProviderOption{WithRoundRobin()},
			invoke: func(i1, i2, i3 int) {
				if i1 != 1 {
					t.Fatalf("unexpected: %d", i1)
				}
				if i2 != 2 {
					t.Fatalf("unexpected: %d", i2)
				}
				if i3 != 1 {
					t.Fatalf("unexpected: %d", i3)
				}
			},
		},
		"success_func_int_round_robin": {
			provide:         func() []int { return []int{1, 2} },
			providerOptions: []ProviderOption{WithRoundRobin()},
			invoke: func(i1, i2, i3 int) {
				if i1 != 1 {
					t.Fatalf("unexpected: %d", i1)
				}
				if i2 != 2 {
					t.Fatalf("unexpected: %d", i2)
				}
				if i3 != 1 {
					t.Fatalf("unexpected: %d", i3)
				}
			},
		},
		"success_value_eager_loading": {
			provide:         1,
			providerOptions: []ProviderOption{WithEagerLoading()},
			invoke:          func() {},
		},
		"success_func_eager_loading": {
			provide:         func() int { return 1 },
			providerOptions: []ProviderOption{WithEagerLoading()},
			invoke:          func() {},
		},
		"success_func_single_instance": {
			provide: func() int { return rand.Int() },
			invoke: func(i1, i2 int) {
				if i1 != i2 {
					t.Fatalf("expected euality: %d %d", i1, i2)
				}
			},
		},
		"success_func_multi_instance": {
			provide:         func() int { return rand.Int() },
			providerOptions: []ProviderOption{WithMultiInstance()},
			invoke: func(i1, i2 int) {
				if i1 == i2 {
					t.Fatalf("expected diff: %d %d", i1, i2)
				}
			},
		},
		"success_value_provide_interface": {
			provide: io.Reader(os.Stdin),
			invoke:  func(r *os.File) {},
		},
		"success_func_provide_interface": {
			provide: func() io.Reader { return os.Stdin },
			invoke:  func(r io.Reader) {},
		},
		"error_value_not_found": {
			provide:   1,
			invoke:    func(s string) {},
			invokeErr: errors.New("not found"),
		},
		"error_func_not_found": {
			provide:   func() int { return 1 },
			invoke:    func(s string) {},
			invokeErr: errors.New("not found"),
		},
		"error_already_exists": {
			provide:    func() (int, int) { return 1, 2 },
			provideErr: errors.New("already exists"),
		},
		"error_no_return_values": {
			provide:    func() {},
			provideErr: errors.New("without return values"),
		},
		"error_non_func_invoke": {
			provide:   1,
			invoke:    "test",
			invokeErr: errors.New("non-function"),
		},
		"error_value_cant_round_robin": {
			provide:         1,
			provideErr:      errors.New("can't round-robin"),
			providerOptions: []ProviderOption{WithRoundRobin()},
		},
		"error_func_cant_round_robin": {
			provide:         func() int { return 1 },
			provideErr:      errors.New("can't round-robin"),
			providerOptions: []ProviderOption{WithRoundRobin()},
		},
		"error_func_provide": {
			provide:   func() (int, error) { return 0, errTest },
			invoke:    func(i int) { t.Fatalf("should not be called") },
			invokeErr: errTest,
		},
		"error_func_eager_loading": {
			provide:         func() (int, error) { return 1, errTest },
			provideErr:      errTest,
			providerOptions: []ProviderOption{WithEagerLoading()},
		},
		"error_provide_interface": {
			provide:   io.Reader(os.Stdin),
			invoke:    func(r io.Reader) {},
			invokeErr: errors.New("not found"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			di := New()

			err := di.Provide(tc.provide, tc.providerOptions...)
			if err != nil {
				t.Logf("provide error: %q", err)
				if tc.provideErr == nil {
					t.Fatalf("unexpecte error: %q", err)
				} else {
					if !errors.Is(err, tc.provideErr) && !strings.Contains(err.Error(), tc.provideErr.Error()) {
						t.Fatalf("expected error: %q, but got: %q", tc.provideErr, err)
					}
				}
				return
			} else {
				if tc.provideErr != nil {
					t.Fatalf("expected error, but got nil")
				}
			}

			err = di.Invoke(tc.invoke)
			if err != nil {
				t.Logf("invoke error: %q", err)
				if tc.invokeErr == nil {
					t.Fatalf("unexpecte error: %q", err)
				} else {
					if !errors.Is(err, tc.invokeErr) && !strings.Contains(err.Error(), tc.invokeErr.Error()) {
						t.Fatalf("expected error: %q, but got: %q", tc.invokeErr, err)
					}
				}
			} else {
				if tc.invokeErr != nil {
					t.Fatalf("expected error, but got nil")
				}
			}
		})
	}
}
