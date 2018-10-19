package server_test

import (
	"testing"

	lens "github.com/RTradeLtd/Lens"
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
	go server.NewAPIServer("127.0.0.1:9999", "Tcp", &cfg)
}
