package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/RTradeLtd/grpc/middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func options(certpath, keypath, token string, logger *zap.SugaredLogger) ([]grpc.ServerOption, error) {
	if token == "" || len(token) < 5 {
		return nil, fmt.Errorf("token '%s' is too short for safe use", token)
	}
	if logger == nil {
		return nil, errors.New("no logger provided")
	}

	// set up logger
	grpcLogger := logger.Desugar().Named("grpc")
	grpc_zap.ReplaceGrpcLogger(grpcLogger)
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Duration("grpc.duration", duration)
		}),
	}

	// set up authentication interceptors
	unaryIntercept, streamInterceptor := middleware.NewServerInterceptors(token)

	// set up server options
	serverOpts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			unaryIntercept,
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(grpcLogger, zapOpts...)),
		grpc_middleware.WithStreamServerChain(
			streamInterceptor,
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.StreamServerInterceptor(grpcLogger, zapOpts...)),
	}

	// set up tls configuration
	if certpath != "" {
		logger.Infow("setting up TLS",
			"cert", certpath,
			"key", keypath)
		creds, err := credentials.NewServerTLSFromFile(
			certpath,
			keypath)
		if err != nil {
			return nil, err
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
	} else {
		logger.Warn("no TLS configuration found")
	}

	return serverOpts, nil
}
