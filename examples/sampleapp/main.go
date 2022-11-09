package main

//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
//go:generate go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
//go:generate protoc --go_out=./generated --go_opt=paths=source_relative --go-grpc_out=./generated --go-grpc_opt=paths=source_relative ./sample.proto
//go:generate go run golang.org/x/tools/cmd/goimports@latest -w .

import (
	"context"
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3"
	"flamingo.me/grpc"
	"sampleapp/generated"
)

type (
	SampleModule          struct{}
	ServiceImplementation struct {
		generated.UnimplementedExampleServiceServer
	}
)

func main() {
	flamingo.App([]dingo.Module{
		new(grpc.ServerModule),
		new(SampleModule),
	})
}

// Configure is the default Method a Module needs to implement
func (m *SampleModule) Configure(injector *dingo.Injector) {
	injector.BindMulti(new(grpc.ServerRegister)).ToProvider(serverRegisterProvider)
}

func serverRegisterProvider(s *ServiceImplementation) grpc.ServerRegister {
	return func(server grpc.ServerRegistrar) {
		generated.RegisterExampleServiceServer(server, s)
	}
}

// implement the generated service interface
var _ generated.ExampleServiceServer = new(ServiceImplementation)

func (s *ServiceImplementation) GetById(context.Context, *generated.GetByIdRequest) (*generated.MyResponse, error) {
	return &generated.MyResponse{Name: "Test"}, nil
}
