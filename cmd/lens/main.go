package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/server"

	"github.com/RTradeLtd/cmd"
	"github.com/RTradeLtd/config"
)

var (
	// Version denotes the tag of this build
	Version string

	certFile = filepath.Join(os.Getenv("HOME"), "/certificates/api.pem")
	keyFile  = filepath.Join(os.Getenv("HOME"), "/certificates/api.key")
	tCfg     config.TemporalConfig
)

var commands = map[string]cmd.Cmd{
	"service": cmd.Cmd{
		Blurb:       "start Lens service",
		Description: "Start the Lens meta data extraction service, which includes the API",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			lensCfg := lens.ConfigOpts{UseChainAlgorithm: true, DataStorePath: "/tmp/badgerds-lens"}
			if err := server.NewAPIServer("0.0.0.0:9999", "tcp", &lensCfg); err != nil {
				log.Fatal(err)
			}
		},
	},
}

func main() {
	// create app
	temporal := cmd.New(commands, cmd.Config{
		Name:     "Temporal",
		ExecName: "temporal",
		Version:  Version,
		Desc:     "Temporal is an easy-to-use interface into distributed and decentralized storage technologies for personal and enterprise use cases.",
	})

	// run no-config commands, exit if command was run
	if exit := temporal.PreRun(os.Args[1:]); exit == cmd.CodeOK {
		os.Exit(0)
	}

	// load config
	configDag := os.Getenv("CONFIG_DAG")
	if configDag == "" {
		log.Fatal("CONFIG_DAG is not set")
	}
	tCfg, err := config.LoadConfig(configDag)
	if err != nil {
		log.Fatal(err)
	}

	// load arguments
	flags := map[string]string{
		"configDag":     configDag,
		"certFilePath":  tCfg.API.Connection.Certificates.CertPath,
		"keyFilePath":   tCfg.API.Connection.Certificates.KeyPath,
		"listenAddress": tCfg.API.Connection.ListenAddress,

		"dbPass": tCfg.Database.Password,
		"dbURL":  tCfg.Database.URL,
		"dbUser": tCfg.Database.Username,
	}

	// execute
	os.Exit(temporal.Run(*tCfg, flags, os.Args[1:]))
}
