package server

import (
	"fmt"
	"net"
	"time"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"

	"github.com/RTradeLtd/Lens"
	"github.com/RTradeLtd/Lens/analyzer/images"
	"github.com/RTradeLtd/Lens/search"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/rtfs"

	pb "github.com/RTradeLtd/grpc/lens"
	pbreq "github.com/RTradeLtd/grpc/lens/request"
	pbresp "github.com/RTradeLtd/grpc/lens/response"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// API is the Lens API server
type API struct {
	meta Metadata
	lens *lens.Service

	l *zap.SugaredLogger
}

// Metadata denotes metadata about the server
type Metadata struct {
	Version string
	Edition string
}

// Run is used to create our API server
func Run(

	ctx context.Context,

	// options
	addr string,
	meta Metadata,
	opts lens.ConfigOpts,
	cfg config.TemporalConfig,

	logger *zap.SugaredLogger,

) error {
	// instantiate ipfs connection
	ipfsAPI := fmt.Sprintf("%s:%s", cfg.IPFS.APIConnection.Host, cfg.IPFS.APIConnection.Port)
	logger.Infow("instantiating IPFS connection",
		"ipfs.api", ipfsAPI)
	manager, err := rtfs.NewManager(ipfsAPI, nil, 1*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to instantiate ipfs manager: %s", err.Error())
	}

	// instantiate tensorflow wrappers
	logger.Infow("instantiating tensorflow wrappers",
		"tensorflow.models", opts.ModelsPath)
	ia, err := images.NewAnalyzer(images.ConfigOpts{
		ModelLocation: opts.ModelsPath,
	}, logger.Named("analyzer").Named("images"))
	if err != nil {
		return fmt.Errorf("failed to instantiate image analyzer: %s", err.Error())
	}

	// instantiate search service
	logger.Infow("setting up search",
		"search.datastore", opts.DataStorePath)
	ss, err := search.NewService(opts.DataStorePath)
	if err != nil {
		return fmt.Errorf("failed to instantiate search service: %s", err.Error())
	}
	defer ss.Close()

	// instantiate Lens proper
	logger.Info("instantiating lens service")
	service, err := lens.NewService(opts, cfg, manager, ia, ss, logger)
	if err != nil {
		return err
	}

	// create connection we will listen on
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// instantiate server settings
	serverOpts, err := options(
		cfg.Endpoints.Lens.TLS.CertPath,
		cfg.Endpoints.Lens.TLS.KeyFile,
		cfg.Endpoints.Lens.AuthKey,
		logger)
	if err != nil {
		return err
	}

	// create a grpc server
	var s = &API{
		meta: meta,
		lens: service,
		l:    logger,
	}
	gServer := grpc.NewServer(serverOpts...)
	pb.RegisterIndexerAPIServer(gServer, s)

	// interrupt server gracefully if context is cancelled
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("shutting down server")
				gServer.GracefulStop()
				return
			}
		}
	}()

	// spin up server
	logger.Infow("spinning up server",
		"address", addr)
	if err = gServer.Serve(lis); err != nil {
		logger.Warn("shutting down server",
			"error", err)
		return err
	}
	return nil
}

// Index is used to submit a request for something to be indexed by lens
func (as *API) Index(ctx context.Context, req *pbreq.Index) (*pbresp.Index, error) {
	switch req.GetType() {
	case "ipld":
		break
	default:
		return nil, fmt.Errorf("invalid data type '%s'", req.GetType())
	}

	var name = req.GetIdentifier()
	var reindex = req.GetReindex()
	metaData, err := as.lens.Magnify(name, reindex)
	if err != nil {
		return nil, fmt.Errorf("failed to perform indexing for '%s': %s",
			name, err.Error())
	}

	var resp *lens.Object
	if !reindex {
		if resp, err = as.lens.Store(name, metaData); err != nil {
			return nil, err
		}
	} else {
		b, err := as.lens.Get(name)
		if err != nil {
			return nil, fmt.Errorf("failed to find ID for object '%s'", name)
		}
		id, err := uuid.FromBytes(b)
		if err != nil {
			return nil, fmt.Errorf("invalid uuid found for '%s' ('%s'): %s",
				name, string(b), err.Error())
		}
		if resp, err = as.lens.Update(id, name, metaData); err != nil {
			return nil, fmt.Errorf("failed to update object: %s", err.Error())
		}
	}

	return &pbresp.Index{
		Id:       resp.LensID.String(),
		Keywords: metaData.Summary,
	}, nil
}

// Search is used to submit a simple search request against the lens index
func (as *API) Search(ctx context.Context, req *pbreq.Search) (*pbresp.Results, error) {
	objects, err := as.lens.KeywordSearch(req.Keywords)
	if err != nil {
		return nil, err
	}

	var objs = make([]*pbresp.Object, len(objects))
	for _, v := range objects {
		objs = append(objs, &pbresp.Object{
			Name:     v.Name,
			MimeType: v.MetaData.MimeType,
			Category: v.MetaData.Category,
		})
	}

	return &pbresp.Results{
		Objects: objs,
	}, nil
}
