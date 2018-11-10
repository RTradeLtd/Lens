package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/RTradeLtd/Lens/lens"
	"github.com/RTradeLtd/config"
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

// NewAPIServer is used to create our API server
func NewAPIServer(listenAddr, protocol string, opts *lens.ConfigOpts, cfg *config.TemporalConfig) error {
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

	// create our lens service
	serice, err := lens.NewService(opts, cfg)
	if err != nil {
		return err
	}
	aServer := &APIServer{
		LS: serice,
	}
	// register our gRPC API server, and our service
	pb.RegisterIndexerAPIServer(gServer, aServer)
	// serve the connection
	if err = gServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

// SubmitIndexRequest is used to submit a request for something to be indexed by lens
func (as *APIServer) SubmitIndexRequest(ctx context.Context, req *pbreq.IndexRequest) (*pbresp.IndexResponse, error) {
	fmt.Println("new index request received")
	switch req.GetDataType() {
	case "ipld":
		break
	default:
		return nil, errors.New("invalid data type")
	}
	objectID := req.GetObjectIdentifier()
	_, metaData, err := as.LS.Magnify(objectID)
	if err != nil {
		return nil, err
	}
	indexResponse, err := as.LS.Store(metaData, objectID)
	if err != nil {
		return nil, err
	}
	fmt.Println(metaData.Summary)
	resp := &pbresp.IndexResponse{
		LensIdentifier: indexResponse.LensID.String(),
		Keywords:       metaData.Summary,
	}
	fmt.Println("finished processing index")
	return resp, nil
}

// SubmitSimpleSearchRequest is used to submit a simple search request against the lens index
func (as *APIServer) SubmitSimpleSearchRequest(ctx context.Context, req *pbreq.SearchRequest) (*pbresp.SimpleSearchResponse, error) {
	fmt.Println("receiving search request")
	objects, err := as.LS.SS.KeywordSearch(req.Keywords)
	if err != nil {
		return nil, err
	}
	hashes := []string{}
	for _, v := range *objects {
		hashes = append(hashes, v.Name)
	}
	resp := &pbresp.SimpleSearchResponse{
		Names:      hashes,
		ObjectType: "ipld",
	}
	fmt.Println("finished processing search request")
	return resp, nil
}

// SubmitAdvancedSearchRequest is used to submit an advanced search request against the lens index
func (as *APIServer) SubmitAdvancedSearchRequest(ctx context.Context, in *pbreq.SearchRequest) (*pbresp.AdvancedSearchResponse, error) {
	return nil, nil
}
