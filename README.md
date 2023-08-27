# :paperclips: mDI • Go Dependency Injection

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
