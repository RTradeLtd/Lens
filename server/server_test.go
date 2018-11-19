package server_test

import (
	"context"
	"testing"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/server"
	"github.com/RTradeLtd/config"
	pb "github.com/RTradeLtd/grpc/lens/request"
)

const (
	testHash      = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
	defaultConfig = "../test/config.json"
)

func TestServer(t *testing.T) {
	// skip this as we test during the client test
	t.Skip()
	// start the server
	opts := lens.ConfigOpts{
		UseChainAlgorithm: true,
		DataStorePath:     "/tmp/badgerds-lens",
	}
	cfg, err := config.LoadConfig(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	server.NewAPIServer("127.0.0.1:9999", "Tcp", &opts, cfg)
	serv := server.APIServer{}
	lensService, err := lens.NewService(&opts, cfg)
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
	shouldBeEmptyResp, shouldBeEmptyErr := serv.SubmitSimpleSearchRequest(context.Background(), nil)
	if shouldBeEmptyErr != nil {
		t.Fatal(shouldBeEmptyErr)
	}
	if shouldBeEmptyResp != nil {
		t.Fatal("response should be empty")
	}
}
