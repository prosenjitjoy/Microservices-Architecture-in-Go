package utils

import (
	"context"
	"main/discovery"
	"math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceConnection attemps to select a random service instance and returns gRPC connection to it.
func ServiceConnection(ctx context.Context, serviceName string, registry discovery.Registry) (*grpc.ClientConn, error) {
	addrs, err := registry.ServiceAddresses(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	targetAddress := addrs[rand.Intn(len(addrs))]
	// fmt.Printf("%s: %s\n", strings.ToUpper(serviceName), targetAddress)

	return grpc.Dial(targetAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
