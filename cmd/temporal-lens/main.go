package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/client"
	"github.com/RTradeLtd/Lens/server"
	"github.com/RTradeLtd/cmd"
	"github.com/RTradeLtd/config"
	pbreq "github.com/RTradeLtd/grpc/lens/request"
)

var (
	// Version denotes the tag of this build
	Version string

	certFile = filepath.Join(os.Getenv("HOME"), "/certificates/api.pem")
	keyFile  = filepath.Join(os.Getenv("HOME"), "/certificates/api.key")
	tCfg     config.TemporalConfig
)

var commands = map[string]cmd.Cmd{
	"server": cmd.Cmd{
		Blurb:       "start Lens server",
		Description: "Start the Lens meta data extraction service, which includes the API",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			dsPath := os.Getenv("DS_PATH")
			if dsPath == "" {
				dsPath = "/data/lens/badgerds-lens"
			}
			lensOpts := lens.ConfigOpts{UseChainAlgorithm: true, DataStorePath: dsPath}
			if err := server.NewAPIServer(cfg.Endpoints.LensGRPC, "tcp", &lensOpts, &cfg); err != nil {
				log.Fatal(err)
			}
		},
	},
	"client": cmd.Cmd{
		Blurb:       "run Lens client commands",
		Description: "Used to start query the lens server\n./temporal-lens client [index|search]\nindex is used to submit an index request\nsearch is used to submit a search request",
		Action: func(cfg config.TemporalConfig, args map[string]string) {
			if len(os.Args) < 3 {
				err := fmt.Errorf(
					"not enough arguments provided\n\n%s",
					"/temporal-lens client [index|search]\nindex is used to submit an index request\nsearch is used to submit a search request",
				)
				log.Fatal(err)
			}
			cmd := os.Args[2]
			switch cmd {
			case "index", "search":
				break
			default:
				log.Fatal("invalid 2nd arg, must be index or search")
			}
			lensCfg := lens.ConfigOpts{
				UseChainAlgorithm: true,
				DataStorePath:     "/tmp/badgerds-lens",
			}
			lensIP := os.Getenv("LENS_IP")
			if lensIP == "" {
				log.Fatal("LENS_IP env var is empty")
			}
			lensPort := os.Getenv("LENS_PORT")
			if lensPort == "" {
				log.Fatal("LENS_PORT env var is empty")
			}
			lensCfg.API.IP = "127.0.0.1"
			lensCfg.API.Port = "9998"
			client, err := client.NewClient(&lensCfg, true)
			if err != nil {
				log.Fatal(err)
			}
			if cmd == "index" {
				dataType := os.Getenv("DATA_TYPE")
				if dataType == "" {
					log.Fatal("DATA_TYPE env var is empty, must be the type of data you are submitting to be indexed")
				}
				objectID := os.Getenv("OBJECT_ID")
				if objectID == "" {
					log.Fatal("OBJECT_ID env var is empty. This is the name of an object, such as a content hash for IPFS")
				}
				indexReq := pbreq.IndexRequest{
					DataType:         dataType,
					ObjectIdentifier: objectID,
				}
				indexResp, err := client.SubmitIndexRequest(context.Background(), &indexReq)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("response from index request")
				fmt.Println("Lens Identifier:", indexResp.LensIdentifier)
				fmt.Println("Keywords for this object:", indexResp.Keywords)
			} else if cmd == "search" {
				scanner := bufio.NewScanner(os.Stdin)
				fmt.Println("how many keywords do you wish to search for?")

				numKeywordsString, err := hanldeScanner(scanner)
				if err != nil {
					log.Fatal(err)
				}
				numKeywords, err := strconv.ParseInt(numKeywordsString, 10, 64)
				if err != nil {
					log.Fatal(err)
				}
				keywords := []string{}
				for count := int64(0); count < numKeywords; count++ {
					fmt.Println("enter a keyword to search for")
					word, err := hanldeScanner(scanner)
					if err != nil {
						log.Fatal(err)
					}
					keywords = append(keywords, word) // grab a single line of input
				}
				searchReq := pbreq.SearchRequest{
					Keywords: keywords,
				}
				fmt.Println("sending search request")
				searchResp, err := client.SubmitSearchRequest(context.Background(), &searchReq)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("Response from search request")
				fmt.Printf("%+v\n", searchResp)
			}
		},
	},
}

func hanldeScanner(scanner *bufio.Scanner) (string, error) {
	// grab a single line of input
	scanned := scanner.Scan()
	// check to see if false
	if !scanned {
		// make sure that the false is due to finished reading, and not an error
		if scanner.Err() != nil {
			return "", scanner.Err()
		}
	}
	return scanner.Text(), nil
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
