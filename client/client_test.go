package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/client"
	pb "github.com/RTradeLtd/Lens/models"
	"github.com/RTradeLtd/Lens/server"
)

const (
	testHash = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
)

func TestClient(t *testing.T) {
	// start the server
	cfg := lens.ConfigOpts{
		UseChainAlgorithm: true,
		DataStorePath:     "/tmp/badgerds-lens",
	}
	go server.NewAPIServer("0.0.0.0:9999", "tcp", &cfg)
	cfg.API.IP = "127.0.0.1"
	cfg.API.Port = "9999"
	c, err := client.NewClient(&cfg, true)
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
