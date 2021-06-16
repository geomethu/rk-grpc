// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgrpcbasic

import (
	"context"
	"github.com/rookie-ninja/rk-grpc/interceptor/context"
	"google.golang.org/grpc"
	"net"
)

func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	set := newOptionSet(RpcTypeUnaryClient, opts...)

	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Extract outgoing metadata from context
		outgoingMD := rkgrpcctx.GetOutgoingMD(ctx)
		incomingMD := rkgrpcctx.GetIncomingMD(ctx)

		remoteIp, remotePort, _ := net.SplitHostPort(cc.Target())
		grpcService, grpcMethod := getGrpcInfo(method)
		gwMethod, gwPath, gwScheme, gwUserAgent := getGwInfo(incomingMD)

		ctx = rkgrpcctx.ToRkContext(ctx,
			rkgrpcctx.WithEntryName(set.EntryName),
			rkgrpcctx.WithIncomingMD(incomingMD),
			rkgrpcctx.WithOutgoingMD(outgoingMD),
			rkgrpcctx.WithRpcInfo(&rkgrpcctx.RpcInfo{
				GrpcService: grpcService,
				GrpcMethod:  grpcMethod,
				GwMethod:    gwMethod,
				GwPath:      gwPath,
				GwScheme:    gwScheme,
				GwUserAgent: gwUserAgent,
				Type:        RpcTypeUnaryClient,
				RemoteIp:    remoteIp,
				RemotePort:  remotePort,
			}))

		// Set headers for internal usage
		md := rkgrpcctx.GetIncomingMD(ctx)
		opts = append(opts, grpc.Header(&md))

		// Invoking
		err := invoker(ctx, method, req, resp, cc, opts...)

		rpcInfo := rkgrpcctx.GetRpcInfo(ctx)
		if rpcInfo != nil {
			rpcInfo.Err = err
		}

		return err
	}
}

func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	set := newOptionSet(RpcTypeStreamClient, opts...)

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// Extract outgoing metadata from context
		outgoingMD := rkgrpcctx.GetOutgoingMD(ctx)
		incomingMD := rkgrpcctx.GetIncomingMD(ctx)

		remoteIp, remotePort, _ := net.SplitHostPort(cc.Target())
		grpcService, grpcMethod := getGrpcInfo(method)
		gwMethod, gwPath, gwScheme, gwUserAgent := getGwInfo(incomingMD)

		ctx = rkgrpcctx.ToRkContext(ctx,
			rkgrpcctx.WithEntryName(set.EntryName),
			rkgrpcctx.WithIncomingMD(incomingMD),
			rkgrpcctx.WithOutgoingMD(outgoingMD),
			rkgrpcctx.WithRpcInfo(&rkgrpcctx.RpcInfo{
				GrpcService: grpcService,
				GrpcMethod:  grpcMethod,
				GwMethod:    gwMethod,
				GwPath:      gwPath,
				GwScheme:    gwScheme,
				GwUserAgent: gwUserAgent,
				Type:        RpcTypeStreamClient,
				RemoteIp:    remoteIp,
				RemotePort:  remotePort,
			}))

		// Set headers for internal usage
		md := rkgrpcctx.GetIncomingMD(ctx)
		opts = append(opts, grpc.Header(&md))

		// Invoking
		clientStream, err := streamer(ctx, desc, cc, method, opts...)

		rpcInfo := rkgrpcctx.GetRpcInfo(ctx)
		if rpcInfo != nil {
			rpcInfo.Err = err
		}

		return clientStream, err
	}
}
