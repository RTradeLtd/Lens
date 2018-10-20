package server_test

import (
	"context"
	"testing"

	lens "github.com/RTradeLtd/Lens"
	pb "github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/server"
)

const (
	testHash = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
)

func TestServer(t *testing.T) {
	// start the server
	cfg := lens.ConfigOpts{
		UseChainAlgorithm: true,
		DataStorePath:     "/tmp/badgerds-lens",
	}
	server.NewAPIServer("127.0.0.1:9999", "Tcp", &cfg)
	serv := server.APIServer{}
	lensService, err := lens.NewService(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	serv.LS = lensService
	req := pb.IndexRequest{
		DataType:         "ipld",
		ObjectIdentifier: testHash,
	}
	resp, err := serv.SubmitIndexRequest(context.Background(), &req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.LensIdentifier == "" {
		t.Fatal("response should have uuid")
	}
	shouldBeEmptyResp, shouldBeEmptyErr := serv.SubmitSearchRequest(context.Background(), nil)
	if shouldBeEmptyErr != nil {
		t.Fatal(shouldBeEmptyErr)
	}
	if shouldBeEmptyResp != nil {
		t.Fatal("response should be empty")
	}
}
