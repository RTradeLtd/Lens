package server

import (
	"net"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/grpc/lensv2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RunV2 spins up the V2 Lens gRPC server
func RunV2(stop <-chan bool, l *zap.SugaredLogger, srv lensv2.LensV2Server, cfg config.Lens) error {
	// instantiate server settings
	serverOpts, err := options(
		cfg.TLS.CertPath,
		cfg.TLS.KeyFile,
		cfg.AuthKey,
		l)
	if err != nil {
		return err
	}

	// create connection we will listen on
	lis, err := net.Listen("tcp", cfg.URL)
	if err != nil {
		return err
	}
	defer lis.Close()

	// set up server
	gServer := grpc.NewServer(serverOpts...)
	lensv2.RegisterLensV2Server(gServer, srv)

	// interrupt server gracefully if context is cancelled
	go func() {
		for {
			select {
			case <-stop:
				l.Info("shutting down server")
				gServer.GracefulStop()
				return
			}
		}
	}()

	// spin up server
	l.Infow("spinning up server", "address", cfg.URL)
	if err = gServer.Serve(lis); err != nil {
		l.Warn("shutting down server", "error", err)
		return err
	}

	return nil
}
