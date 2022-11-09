package grpcroute

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HandleProto func(context.Context, interface{}) (interface{}, error)

type Option interface {
	GetHandler(string) (HandleProto, bool) // FullMethod is the full RPC method string, i.e.,  /package.service/method
	SetHandler(string, HandleProto)
	RemoveHandler(string)
}

func GrpcRoute(option Option) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if handler, ok := option.GetHandler(info.FullMethod); ok {
			return handler(ctx, req)
		}

		if handler != nil {
			return handler(ctx, req)
		}

		return nil, status.Error(codes.NotFound, "")
	}
}
