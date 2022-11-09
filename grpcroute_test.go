package grpcroute

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

type AOption struct {
	Handlers map[string]HandleProto
}

func (option *AOption) GetHandler(key string) (HandleProto, bool) {
	handler, ok := option.Handlers[key]
	if ok {
		return handler, true
	}

	return nil, false
}

func (option *AOption) SetHandler(key string, h HandleProto) {
	if key == "" || h == nil {
		return
	}

	if option.Handlers == nil {
		option.Handlers = make(map[string]HandleProto)
	}

	option.Handlers[key] = h
}

func (option *AOption) RemoveHandler(key string) {
	delete(option.Handlers, key)
}

type A struct {
}

func (a *A) Ping(ctx context.Context, req interface{}) (interface{}, error) {
	if req.(string) == "ping" {
		return "pone", nil
	}

	return nil, status.Error(codes.InvalidArgument, "bad request")
}

func TestGrpcRoute(t *testing.T) {
	a := &A{}
	option := &AOption{}
	chain := grpc_middleware.ChainUnaryServer(GrpcRoute(option))

	key := "key"
	value := "val"
	ctx := context.WithValue(context.TODO(), key, value)
	option.SetHandler("/pkg.Service/Ping", a.Ping)

	tests := []struct {
		name       string
		req        interface{}
		fullMethod string
		handler    func(context.Context, interface{}) (interface{}, error)
		wantResp   interface{}
		wantCode   codes.Code
	}{
		{
			name:       "TestRightRequest",
			req:        "ping",
			fullMethod: "/pkg.Service/Ping",
			wantResp:   "pone",
			wantCode:   codes.OK,
		},
		{
			name:       "TestBadRequestForError",
			req:        "hello",
			fullMethod: "/pkg.Service/Ping",
			wantResp:   nil,
			wantCode:   codes.InvalidArgument,
		},
		{
			name:       "TestNoFound",
			req:        "hello",
			fullMethod: "/pkg.Service/Hello",
			wantResp:   nil,
			wantCode:   codes.NotFound,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			resp, err := chain(ctx, tt.req, &grpc.UnaryServerInfo{FullMethod: tt.fullMethod}, nil)
			assert.Equal(t, tt.wantCode, status.Convert(err).Code())
			assert.Equal(t, tt.wantResp, resp)
		})
	}

}
