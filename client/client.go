package client

import (
	"context"
	"fmt"

	"github.com/RTradeLtd/Lens/lens"
	pb "github.com/RTradeLtd/grpc/lens"
	pbreq "github.com/RTradeLtd/grpc/lens/request"
	pbresp "github.com/RTradeLtd/grpc/lens/response"
	"google.golang.org/grpc"
)

// Client is a lens client used to make requests to the Lens gRPC server
type Client struct {
	GC *grpc.ClientConn
	IC pb.IndexerAPIClient
}

// NewClient is used to generate our lens client
func NewClient(opts *lens.ConfigOpts, insecure bool) (*Client, error) {
	apiURL := fmt.Sprintf("%s:%s", opts.API.IP, opts.API.Port)
	var (
		gconn *grpc.ClientConn
		err   error
	)
	if insecure {
		gconn, err = grpc.Dial(apiURL, grpc.WithInsecure())
	}
	if err != nil {
		return nil, err
	}
	indexerConn := pb.NewIndexerAPIClient(gconn)
	return &Client{
		GC: gconn,
		IC: indexerConn,
	}, nil
}

// SubmitIndexRequest is used to submit an index request to lens
func (c *Client) SubmitIndexRequest(ctx context.Context, req *pbreq.IndexRequest) (*pbresp.IndexResponse, error) {
	return c.IC.SubmitIndexRequest(ctx, req)
}

// SubmitSimpleSearchRequest is used to submit a simple search request against lens
func (c *Client) SubmitSimpleSearchRequest(ctx context.Context, req *pbreq.SearchRequest) (*pbresp.SimpleSearchResponse, error) {
	return c.IC.SubmitSimpleSearchRequest(ctx, req)
}
