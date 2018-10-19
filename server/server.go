package server

import (
	"net"

	"github.com/RTradeLtd/Lens"
	pb "github.com/RTradeLtd/Lens/models"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// APIServer is the Lens API server
type APIServer struct {
	LS *lens.Service
}

// NewAPIServer is used to create our API server
func NewAPIServer(listenAddr, protocol string) error {
	lis, err := net.Listen(listenAddr, protocol)
	if err != nil {
		return err
	}
	defer lis.Close()
	gServer := grpc.NewServer()
	aServer := &APIServer{}
	pb.RegisterIndexerAPIServer(gServer, aServer)
	if err = gServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

// SubmitIndexRequest is used to submit a request for something to be indexed by lens
func (as *APIServer) SubmitIndexRequest(ctx context.Context, req *pb.IndexRequest) (*pb.IndexResponse, error) {
	return nil, nil
}

// SubmitSearchRequest is used to submit a request ot search through lens
func (as *APIServer) SubmitSearchRequest(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	return nil, nil
}
