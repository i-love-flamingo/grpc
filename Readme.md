# GRPC Module

This module can be used to run a grpc server in your flamingo application.
It uses the [google grpc](google.golang.org/grpc) implementation and makes it easy to use within a flamingo application.

Before you start using this module - it makes sense to be familiar with the [official grpc go tutorials](https://grpc.io/docs/languages/go/basics/)

The module provides:

* Starting of a grpc server with the registered grpc services
* A concept to bind grpc services (in our own modules)
* Providing authentication features, that can be used to secure grpc services. For example by using OAuth tokens in a bearer request header.

## How to use it

1) Generate: Define your grpc service in your `.proto` file. And generate the go client and server. 
2) Service Implementation: Implementing the service interface generated from our service definition: doing the actual “work” of our service.
3) Binding: Inside your module you need to register the service implementation. This is done like this:
```go
func (m *SampleModule) Configure(injector *dingo.Injector) {
	injector.BindMulti(new(grpc.ServerRegister)).ToProvider(serverRegisterProvider)
}

func serverRegisterProvider(s *ServiceImplementation) grpc.ServerRegister {
	return func(server grpc.ServerRegistrar) {
		generated.RegisterExampleServiceServer(server, s)
	}
}
```
Where `ServiceImplementation` is your implementation of the generated service interface. And `RegisterExampleServiceServer` is the generated helper method to register.
4) Add flamingo grpc server: Inside your `main.go` add the `grpc.ServerModule` to the flamingo bootstrap.
5) You may now configure the authenticators inside your configuration:
```yaml
grpc.identifier: test
```

## Example

See the [example application](examples/sampleapp/main.go) inside this module.
The grpc server start on the default port :11101

## Authenticators

To enable OAuth bearer authentication add this to your `config.cue`

```cue
grpc: identifier: [
    {"identifier": "management", "provider": "oauth2", "issuer": "http://localhost:8080/auth/realms/testreal", "clientID": "sampleapp"},
]

```