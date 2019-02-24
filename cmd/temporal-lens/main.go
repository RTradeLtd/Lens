package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	lens "github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/Lens/engine"
	"github.com/RTradeLtd/Lens/engine/queue"
	"github.com/RTradeLtd/Lens/logs"
	"github.com/RTradeLtd/Lens/server"
	"github.com/RTradeLtd/cmd"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/rtfs"
)

var (
	// Version denotes the tag of this build
	Version string

	// Edition indicates the this build's type
	Edition string

	// flag configuration
	cfgPath = flag.String("cfg", os.Getenv("CONFIG_DAG"),
		"path to Temporal configuration")
	modelPath = flag.String("models", "/tmp",
		"path to TensorFlow models")
	logPath = flag.String("logpath", "",
		"path to write logs to - leave blank for stdout")
	devMode = flag.Bool("dev", false,
		"enable dev mode")
)

var commands = map[string]cmd.Cmd{
	"v2": cmd.Cmd{
		Blurb: "start the Lens V2 server",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			// set up logger
			l, err := logs.NewLogger(*logPath, *devMode)
			if err != nil {
				log.Fatal("failed to instantiate logger:", err.Error())
			}
			defer l.Sync()

			// instantiate ipfs connection
			var ipfsURL = fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port)
			l.Infow("instantiating IPFS connection", "ipfs.url", ipfsURL)
			manager, err := rtfs.NewManager(ipfsURL, "", 1*time.Minute)
			if err != nil {
				l.Fatalw("failed to instantiate ipfs manager", "error", err)
			}

			// instantiate tensorflow wrapper
			l.Infow("instantiating tensorflow wrappers", "tensorflow.models", *modelPath)
			tf, err := images.NewAnalyzer(images.ConfigOpts{
				ModelLocation: *modelPath,
			}, l.Named("analyzer").Named("images"))
			if err != nil {
				l.Fatalw("failed to instantiate image analyzer", "error", err)
			}

			// create lens v2 service
			l.Info("instantiating Lens V2")
			srv, err := lens.NewV2(lens.V2Options{
				Engine: engine.Opts{
					StorePath: cfg.Lens.Options.Engine.StorePath,
					Queue: queue.Options{
						Rate:      time.Duration(cfg.Lens.Options.Engine.Queue.Rate) * time.Second,
						BatchSize: cfg.Lens.Options.Engine.Queue.Batch,
					},
				},
			}, manager, tf, l)
			if err != nil {
				l.Fatalw("failed to instantiate Lens V2", "error", err)
			}

			// set up interrupts
			var stop = make(chan bool)
			var signals = make(chan os.Signal)
			signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
			go func() {
				<-signals
				stop <- true
			}()

			// go!
			l.Infow("spinning up server", "config", cfg.Services.Lens)
			if err := server.RunV2(stop, l, srv, cfg.Services.Lens); err != nil {
				l.Fatalw("error encountered on server run", "error", err)
			}
		},
	},
}

func main() {
	if Version == "" {
		Version = "unknown"
	}

	// create app
	tlens := cmd.New(commands, cmd.Config{
		Name:     "Lens",
		ExecName: "temporal-lens",
		Version:  fmt.Sprintf("%s (%s edition)", Version, Edition),
		Desc:     "Lens is a tool to aid content discovery fro the distributed web",
	})

	// run no-config commands, exit if command was run
	flag.Parse()
	if exit := tlens.PreRun(flag.Args()); exit == cmd.CodeOK {
		os.Exit(0)
	}

	// load config
	if cfgPath == nil || *cfgPath == "" {
		log.Fatal("no configuration file provided - set CONFIG_DAG or use the --cfg flag")
	}
	tCfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	// load arguments
	flags := map[string]string{
		"configDag":     *cfgPath,
		"certFilePath":  tCfg.API.Connection.Certificates.CertPath,
		"keyFilePath":   tCfg.API.Connection.Certificates.KeyPath,
		"listenAddress": tCfg.API.Connection.ListenAddress,

		"dbPass": tCfg.Database.Password,
		"dbURL":  tCfg.Database.URL,
		"dbUser": tCfg.Database.Username,
	}

	// execute
	os.Exit(tlens.Run(*tCfg, flags, flag.Args()))
}
