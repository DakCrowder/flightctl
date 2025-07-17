package periodic

import (
	"context"
	"fmt"
	"time"

	"github.com/flightctl/flightctl/internal/config"
	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/internal/kvstore"
	"github.com/flightctl/flightctl/internal/service"
	"github.com/flightctl/flightctl/internal/store"
	"github.com/flightctl/flightctl/internal/tasks_client"
	"github.com/flightctl/flightctl/pkg/queues"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type DefaultServiceProvider struct {
	cfg             *config.Config
	log             logrus.FieldLogger
	store           store.Store
	queuesProvider  queues.Provider
	kvStore         kvstore.KVStore
	callbackManager tasks_client.CallbackManager
}

func NewDefaultServiceProvider(
	cfg *config.Config,
	log logrus.FieldLogger,
	store store.Store,
	queuesProvider queues.Provider,
	kvStore kvstore.KVStore,
) (*DefaultServiceProvider, error) {
	publisher, err := tasks_client.TaskQueuePublisher(queuesProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create task queue publisher: %w", err)
	}

	callbackManager := tasks_client.NewCallbackManager(publisher, log)

	return &DefaultServiceProvider{
		cfg:             cfg,
		log:             log,
		store:           store,
		queuesProvider:  queuesProvider,
		kvStore:         kvStore,
		callbackManager: callbackManager,
	}, nil
}

// GetServiceHandler creates a service handler for a specific organization
func (sp *DefaultServiceProvider) GetServiceHandler(ctx context.Context, orgID uuid.UUID) (service.Service, error) {
	if ctx.Value(consts.OrganizationIDCtxKey) == nil {
		ctx = context.WithValue(ctx, consts.OrganizationIDCtxKey, orgID)
	}

	return service.WrapWithTracing(service.NewServiceHandler(sp.store, sp.callbackManager, sp.kvStore, nil, sp.log, "", "")), nil
}

func (sp *DefaultServiceProvider) GetCallbackManager() tasks_client.CallbackManager {
	return sp.callbackManager
}

func (sp *DefaultServiceProvider) GetConfig() interface{} {
	return &configWrapper{
		eventRetentionPeriod: time.Duration(sp.cfg.Service.EventRetentionPeriod),
	}
}

// TaskIntervals defines the intervals for each task type
var TaskIntervals = map[TaskType]time.Duration{
	TaskTypeRepoTest:               2 * time.Minute,
	TaskTypeResourceSync:           2 * time.Minute,
	TaskTypeDeviceDisconnected:     2 * time.Minute,  // From tasks.DeviceDisconnectedPollingInterval
	TaskTypeRolloutDeviceSelection: 30 * time.Second, // From device_selection.RolloutDeviceSelectionInterval
	TaskTypeDisruptionBudget:       30 * time.Second, // From disruption_budget.DisruptionBudgetReconcilationInterval
	TaskTypeEventCleanup:           10 * time.Minute, // From tasks.EventCleanupPollingInterval
}
