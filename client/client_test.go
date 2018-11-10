package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/RTradeLtd/Lens/client"
	"github.com/RTradeLtd/Lens/lens"
	"github.com/RTradeLtd/Lens/server"
	"github.com/RTradeLtd/config"
	pb "github.com/RTradeLtd/grpc/lens/request"
)

const (
	testHash      = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
	defaultConfig = "../test/config.json"
)

func TestClient(t *testing.T) {
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
	server.NewAPIServer("0.0.0.0:9999", "tcp", &opts, cfg)
	opts.API.IP = "127.0.0.1"
	opts.API.Port = "9999"
	c, err := client.NewClient(&opts, true)
	if err != nil {
		t.Fatal(err)
	}
	req := &pb.IndexRequest{
		DataType:         "ipld",
		ObjectIdentifier: testHash,
	}
	resp, err := c.SubmitIndexRequest(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", resp)
}
