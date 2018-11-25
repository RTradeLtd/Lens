package server

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/rtfs"

	pb "github.com/RTradeLtd/grpc/lens"
	pbreq "github.com/RTradeLtd/grpc/lens/request"
	pbresp "github.com/RTradeLtd/grpc/lens/response"

	"github.com/RTradeLtd/grpc/middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// APIServer is the Lens API server
type APIServer struct {
	LS *lens.Service
}

// Run is used to create our API server
func Run(listenAddr, protocol string, opts lens.ConfigOpts, cfg config.TemporalConfig) error {
	// instantiate ipfs connection
	ipfsAPI := fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port)
	manager, err := rtfs.NewManager(ipfsAPI, nil, 1*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to instantiate ipfs manager: %s", err.Error())
	}

	// instantiate tensorflow wrappers
	ia, err := images.NewAnalyzer(images.ConfigOpts{
		ModelLocation: opts.ModelsPath,
	})
	if err != nil {
		return fmt.Errorf("failed to instantiate image analyzer: %s", err.Error())
	}

	// create our lens service
	service, err := lens.NewService(opts, cfg, manager, ia)
	if err != nil {
		return err
	}
	var s = &APIServer{LS: service}

	// create connection we will listen on
	lis, err := net.Listen(protocol, listenAddr)
	if err != nil {
		return err
	}

	// setup authentication interceptor
	unaryIntercept, streamInterceptor := middleware.NewServerInterceptors(cfg.Endpoints.Lens.AuthKey)
	// setup server options
	serverOpts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			unaryIntercept,
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor))),
		grpc_middleware.WithStreamServerChain(
			streamInterceptor,
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor))),
	}

	// setup tls configuration
	if cfg.Lens.TLS.CertPath != "" {
		creds, err := credentials.NewServerTLSFromFile(
			cfg.Endpoints.Lens.TLS.CertPath,
			cfg.Endpoints.Lens.TLS.KeyFile)
		if err != nil {
			return err
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
	}

	// create a grpc server
	gServer := grpc.NewServer(serverOpts...)
	pb.RegisterIndexerAPIServer(gServer, s)
	if err = gServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

// Index is used to submit a request for something to be indexed by lens
func (as *APIServer) Index(ctx context.Context, req *pbreq.Index) (*pbresp.Index, error) {
	switch req.GetDataType() {
	case "ipld":
		break
	default:
		return nil, errors.New("invalid data type")
	}

	var objectID = req.GetObjectIdentifier()
	metaData, err := as.LS.Magnify(objectID)
	if err != nil {
		return nil, err
	}

	indexResponse, err := as.LS.Store(metaData, objectID)
	if err != nil {
		return nil, err
	}

	return &pbresp.Index{
		Id:       indexResponse.LensID.String(),
		Keywords: metaData.Summary,
	}, nil
}

// Search is used to submit a simple search request against the lens index
func (as *APIServer) Search(ctx context.Context, req *pbreq.Search) (*pbresp.Results, error) {
	objects, err := as.LS.KeywordSearch(req.Keywords)
	if err != nil {
		return nil, err
	}

	var objs = make([]*pbresp.Object, len(objects))
	for _, v := range objects {
		objs = append(objs, &pbresp.Object{
			Name:     v.Name,
			MimeType: v.MetaData.MimeType,
			Category: v.MetaData.Category,
		})
	}

	return &pbresp.Results{
		Objects: objs,
	}, nil
}
