package periodic

import (
	"context"
	"time"

	"github.com/flightctl/flightctl/internal/config"
	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/internal/kvstore"
	"github.com/flightctl/flightctl/internal/rollout/device_selection"
	"github.com/flightctl/flightctl/internal/rollout/disruption_budget"
	"github.com/flightctl/flightctl/internal/service"
	"github.com/flightctl/flightctl/internal/store"
	"github.com/flightctl/flightctl/internal/tasks"
	"github.com/flightctl/flightctl/internal/tasks_client"
	"github.com/flightctl/flightctl/pkg/queues"
	"github.com/flightctl/flightctl/pkg/thread"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type DefaultWorkerFactory struct {
	cfg            *config.Config
	log            logrus.FieldLogger
	store          store.Store
	queuesProvider queues.Provider
	kvStore        kvstore.KVStore
}

func NewDefaultWorkerFactory(
	cfg *config.Config,
	log logrus.FieldLogger,
	store store.Store,
	queuesProvider queues.Provider,
	kvStore kvstore.KVStore,
) *DefaultWorkerFactory {
	return &DefaultWorkerFactory{
		cfg:            cfg,
		log:            log,
		store:          store,
		queuesProvider: queuesProvider,
		kvStore:        kvStore,
	}
}

// CreateWorkers creates all worker threads for an organization
func (f *DefaultWorkerFactory) CreateWorkers(ctx context.Context, orgId uuid.UUID) (*OrganizationWorkers, error) {
	orgCtx := context.WithValue(ctx, consts.OrganizationIDCtxKey, orgId)
	orgCtx = context.WithValue(orgCtx, consts.EventSourceComponentCtxKey, "flightctl-periodic")
	orgCtx = context.WithValue(orgCtx, consts.EventActorCtxKey, "service:flightctl-periodic")
	orgCtx = context.WithValue(orgCtx, consts.InternalRequestCtxKey, true)

	publisher, err := tasks_client.TaskQueuePublisher(f.queuesProvider)
	if err != nil {
		return nil, err
	}

	callbackManager := tasks_client.NewCallbackManager(publisher, f.log)

	serviceHandler := service.WrapWithTracing(service.NewServiceHandler(f.store, callbackManager, f.kvStore, nil, f.log, "", ""))

	orgLog := f.log.WithField("org", orgId)

	workers := &OrganizationWorkers{
		orgId: orgId,
	}

	// Repository tester
	repoTester := tasks.NewRepoTester(orgLog, serviceHandler)
	workers.repoTesterThread = thread.New(orgCtx,
		orgLog.WithField("pkg", "repository-tester"), "Repository tester", 2*time.Minute,
		func(ctx context.Context) { repoTester.TestRepositories(ctx) })
	workers.repoTesterThread.Start()

	// Resource sync
	resourceSync := tasks.NewResourceSync(callbackManager, serviceHandler, orgLog)
	workers.resourceSyncThread = thread.New(orgCtx,
		orgLog.WithField("pkg", "resourcesync"), "ResourceSync", 2*time.Minute,
		func(ctx context.Context) { resourceSync.Poll(ctx) })
	workers.resourceSyncThread.Start()

	// Device disconnected
	deviceDisconnected := tasks.NewDeviceDisconnected(orgLog, serviceHandler)
	workers.deviceDisconnectedThread = thread.New(orgCtx,
		orgLog.WithField("pkg", "device-disconnected"), "Device disconnected", tasks.DeviceDisconnectedPollingInterval,
		func(ctx context.Context) { deviceDisconnected.Poll(ctx) })
	workers.deviceDisconnectedThread.Start()

	// Rollout device selection
	rolloutDeviceSelection := device_selection.NewReconciler(serviceHandler, callbackManager, orgLog)
	workers.rolloutDeviceSelectionThread = thread.New(orgCtx,
		orgLog.WithField("pkg", "rollout-device-selection"), "Rollout device selection", device_selection.RolloutDeviceSelectionInterval,
		func(ctx context.Context) { rolloutDeviceSelection.Reconcile(ctx) })
	workers.rolloutDeviceSelectionThread.Start()

	// Rollout disruption budget
	disruptionBudget := disruption_budget.NewReconciler(serviceHandler, callbackManager, orgLog)
	workers.disruptionBudgetThread = thread.New(orgCtx,
		orgLog.WithField("pkg", "disruption-budget"), "Disruption budget", disruption_budget.DisruptionBudgetReconcilationInterval,
		func(ctx context.Context) { disruptionBudget.Reconcile(ctx) })
	workers.disruptionBudgetThread.Start()

	// Event cleanup
	eventCleanup := tasks.NewEventCleanup(orgLog, serviceHandler, f.cfg.Service.EventRetentionPeriod)
	workers.eventCleanupThread = thread.New(orgCtx,
		orgLog.WithField("pkg", "event-cleanup"), "Event cleanup", tasks.EventCleanupPollingInterval,
		func(ctx context.Context) { eventCleanup.Poll(ctx) })
	workers.eventCleanupThread.Start()

	return workers, nil
}
