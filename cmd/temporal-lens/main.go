package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/search"
	"github.com/RTradeLtd/Lens/server"
	"github.com/RTradeLtd/cmd"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/rtfs"
)

var (
	// Version denotes the tag of this build
	Version string
	dsPath  = flag.String("datastore", "/data/lens/badgerds-lens",
		"path to Badger datastore")
	logPath = flag.String("logpath", "",
		"path to write logs to - leave blank for stdout")
	devMode = flag.Bool("dev", false,
		"enable dev mode")
)

var commands = map[string]cmd.Cmd{
	"server": cmd.Cmd{
		Blurb:       "start Lens server",
		Description: "Start the Lens meta data extraction service, which includes the API",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			l, err := logs.NewLogger(*logPath, *devMode)
			if err != nil {
				log.Fatal("failed to instantiate logger:", err.Error())
			}
			defer l.Sync()

			l = l.With("version", Version)
			if *logPath != "" {
				println("logger initialized - output will be written to", *logPath)
			}

			// handle graceful shutdown
			ctx, cancel := context.WithCancel(context.Background())
			signals := make(chan os.Signal)
			signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
			go func() {
				<-signals
				cancel()
			}()

			// let's goooo
			if err := server.Run(
				ctx,
				cfg.Endpoints.Lens.URL,
				lens.ConfigOpts{
					UseChainAlgorithm: true,
					DataStorePath:     *dsPath,
				},
				cfg,
				l.Named("server"),
			); err != nil {
				log.Fatal(err)
			}
		},
	},
	"migrate": cmd.Cmd{
		Blurb:       "Used to migrate teh datastore",
		Description: "Performs a complete migration of the old datastore to new datastore",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			im, err := rtfs.NewManager(fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port),
				nil, 5*time.Minute)
			if err != nil {
				log.Fatal(err)
			}
			s, err := search.NewService(*dsPath)
			if err != nil {
				log.Fatal(err)
			}
			entriesToMigrate, err := s.GetEntries()
			if err != nil {
				log.Fatal(err)
			}
			if err = s.MigrateEntries(entriesToMigrate, im, true); err != nil {
				log.Fatal(err)
			}
		},
	},
}

func main() {
	// create app
	temporal := cmd.New(commands, cmd.Config{
		Name:     "Lens",
		ExecName: "temporal-lens",
		Version:  "mvp",
		Desc:     "Lens is a tool to aid content discovery fro the distributed web",
	})

	// run no-config commands, exit if command was run
	flag.Parse()
	if exit := temporal.PreRun(flag.Args()); exit == cmd.CodeOK {
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
