// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgrpcbasicauth

import (
	"context"
	"encoding/base64"
	rkgrpcbasic "github.com/rookie-ninja/rk-grpc/interceptor/basic"
	"github.com/rookie-ninja/rk-grpc/interceptor/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	set := newOptionSet(rkgrpcbasic.RpcTypeUnaryServer, opts...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Before invoking
		if err := serverBefore(ctx, set); err != nil {
			if rpcInfo := rkgrpcctx.GetRpcInfo(ctx); rpcInfo != nil {
				rpcInfo.Err = err
			}
			return nil, err
		}

		// Invoking
		return handler(ctx, req)
	}
}

func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	set := newOptionSet(rkgrpcbasic.RpcTypeStreamServer, opts...)

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Before invoking
		wrappedStream := rkgrpcctx.WrapServerStream(stream)
		ctx := wrappedStream.WrappedContext

		// Before invoking
		if err := serverBefore(ctx, set); err != nil {
			if rpcInfo := rkgrpcctx.GetRpcInfo(ctx); rpcInfo != nil {
				rpcInfo.Err = err
			}
			return err
		}

		// Invoking
		err := handler(srv, wrappedStream)

		// After invoking
		if rpcInfo := rkgrpcctx.GetRpcInfo(ctx); rpcInfo != nil {
			rpcInfo.Err = err
		}

		return err
	}
}

func serverBefore(ctx context.Context, set *optionSet) error {
	val := rkgrpcctx.GetValueFromIncomingMD(ctx, "authorization")

	if len(val) < 1 {
		return status.Error(codes.Unauthenticated, `Missing auth header`)
	}

	credRaw := val[0]

	// missing prefix
	prefix := "Bearer "
	if !strings.HasPrefix(credRaw, prefix) {
		return status.Error(codes.Unauthenticated, `Missing "Basic " prefix in "Authorization" header`)
	}

	// invalid decoding
	credBytes, err := base64.StdEncoding.DecodeString(credRaw[len(prefix):])
	if err != nil {
		return status.Error(codes.Unauthenticated, `Invalid base64 in header`)
	}

	credStr := string(credBytes)
	index := strings.IndexByte(credStr, ':')
	if index < 0 {
		return status.Error(codes.Unauthenticated, `Invalid basic auth format`)
	}

	username, password := credStr[:index], credStr[index+1:]

	if !set.Authorized(username, password) {
		return status.Error(codes.Unauthenticated, "Invalid username or password")
	}

	return nil
}
