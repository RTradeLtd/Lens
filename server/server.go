package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/config"
	pb "github.com/RTradeLtd/grpc/lens"
	pbreq "github.com/RTradeLtd/grpc/lens/request"
	pbresp "github.com/RTradeLtd/grpc/lens/response"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
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
	defer lis.Close()
	// create a grpc server
	gServer := grpc.NewServer()
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

// SubmitSearchRequest is used to submit a request ot search through lens
func (as *APIServer) SubmitSearchRequest(ctx context.Context, req *pbreq.SearchRequest) (*pbresp.SearchResponse, error) {
	fmt.Println("receiving search request")
	hashes, err := as.LS.SS.KeywordSearch(req.Keywords)
	if err != nil {
		return nil, err
	}
	resp := &pbresp.SearchResponse{
		Names:      hashes,
		ObjectType: "ipld",
	}
	fmt.Println("finished processing search request")
	return resp, nil
}
