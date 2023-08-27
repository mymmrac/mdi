# :paperclips: mDI • Go Dependency Injection

[![Go Reference](https://pkg.go.dev/badge/github.com/mymmrac/mdi#section-readme.svg)](https://pkg.go.dev/github.com/mymmrac/mdi)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mymmrac/mdi?logo=go)](go.mod)
[![CI Status](https://github.com/mymmrac/mdi/actions/workflows/ci.yml/badge.svg)](https://github.com/mymmrac/mdi/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mymmrac/mdi)](https://goreportcard.com/report/github.com/mymmrac/mdi)

mDI is a zero dependency reflection-based dependency injection library for Go

## :zap: Usage

How to get the library:

```shell
go get -u github.com/mymmrac/mdi
```

How to use:

```go
// Define your dependencies

type Service struct { ... }

func NewService() (*Service, error) { ... }

// Define DI and provide dependencies

di := mdi.New()
di.MustProvide(NewService)

// Invoke your functions with dependencies

err := di.Invoke(func(service *Service) { ... })
```

# :lock: License

mDI is distributed under [MIT license](LICENSE)

## :beer: Credits

The library is based on [rathil/rdi](https://gitlab.com/rathil/rdi), but with few modifications
