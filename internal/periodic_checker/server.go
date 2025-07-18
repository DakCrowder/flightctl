package periodic

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flightctl/flightctl/internal/config"
	"github.com/flightctl/flightctl/internal/kvstore"
	"github.com/flightctl/flightctl/internal/service"
	"github.com/flightctl/flightctl/internal/store"
	"github.com/flightctl/flightctl/internal/tasks_client"
	"github.com/flightctl/flightctl/internal/tasks_runner"
	"github.com/flightctl/flightctl/pkg/queues"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	defaultWorkerPoolSize = 20
)

type Server struct {
	cfg   *config.Config
	log   logrus.FieldLogger
	store store.Store
}

// New returns a new instance of a flightctl server.
func New(
	cfg *config.Config,
	log logrus.FieldLogger,
	store store.Store,
) *Server {
	return &Server{
		cfg:   cfg,
		log:   log,
		store: store,
	}
}

// TODO: expose metrics
func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	queuesProvider, err := queues.NewRedisProvider(ctx, s.log, s.cfg.KV.Hostname, s.cfg.KV.Port, s.cfg.KV.Password)
	if err != nil {
		return err
	}
	defer queuesProvider.Stop()

	kvStore, err := kvstore.NewKVStore(ctx, s.log, s.cfg.KV.Hostname, s.cfg.KV.Port, s.cfg.KV.Password)
	if err != nil {
		return err
	}

	publisher, err := queuesProvider.NewPublisher("periodic-checker")
	if err != nil {
		return err
	}
	callbackManager := tasks_client.NewCallbackManager(publisher, s.log)
	serviceHandler := service.WrapWithTracing(service.NewServiceHandler(s.store, callbackManager, kvStore, nil, s.log, "", ""))

	// Create service handler function for executors
	serviceHandlerFunc := func(ctx context.Context, orgID uuid.UUID) (service.Service, error) {
		return service.WrapWithTracing(service.NewServiceHandler(s.store, callbackManager, kvStore, nil, s.log, "", "")), nil
	}

	// Get event retention period from config
	eventRetentionPeriod := time.Duration(s.cfg.Service.EventRetentionPeriod)

	// Create executors with encapsulated dependencies
	executors := CreateExecutors(serviceHandlerFunc, callbackManager, eventRetentionPeriod, s.log)

	// Get task metadata
	taskMetadata := GetTaskMetadata()

	// Create task queue
	taskQueue := tasks_runner.NewTaskQueue()

	// Create worker pool with configurable number of workers (default to 20)
	numWorkers := defaultWorkerPoolSize
	if s.cfg.Periodic != nil && s.cfg.Periodic.WorkerPoolSize > 0 {
		numWorkers = s.cfg.Periodic.WorkerPoolSize
	}
	workerPool := tasks_runner.NewWorkerPool(numWorkers, taskQueue, executors, s.log)

	// Create organization manager
	orgManager := tasks_runner.NewOrganizationManager(serviceHandler, s.log, taskQueue, workerPool, taskMetadata)
	go orgManager.Start(ctx)

	sigShutdown := make(chan os.Signal, 1)

	signal.Notify(sigShutdown, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigShutdown
	s.log.Println("Shutdown signal received")
	return nil
}
