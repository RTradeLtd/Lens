package server

import (
	"github.com/RTradeLtd/grpc/middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func options(certpath, keypath, token string, logger *zap.SugaredLogger) ([]grpc.ServerOption, error) {
	// set up authentication interceptors
	unaryIntercept, streamInterceptor := middleware.NewServerInterceptors(token)

	// set up server options
	serverOpts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			unaryIntercept,
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor))),
		grpc_middleware.WithStreamServerChain(
			streamInterceptor,
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor))),
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
