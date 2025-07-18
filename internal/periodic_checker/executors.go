package periodic

import (
	"context"
	"fmt"
	"time"

	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/internal/rollout/device_selection"
	"github.com/flightctl/flightctl/internal/rollout/disruption_budget"
	"github.com/flightctl/flightctl/internal/service"
	"github.com/flightctl/flightctl/internal/tasks"
	"github.com/flightctl/flightctl/internal/tasks_client"
	"github.com/flightctl/flightctl/internal/tasks_runner"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Task type constants
const (
	TaskTypeRepoTest               = "repo_test"
	TaskTypeResourceSync           = "resource_sync"
	TaskTypeDeviceDisconnected     = "device_disconnected"
	TaskTypeRolloutDeviceSelection = "rollout_device_selection"
	TaskTypeDisruptionBudget       = "disruption_budget"
	TaskTypeEventCleanup           = "event_cleanup"
)

// baseExecutor provides common functionality for all executors
type baseExecutor struct {
	serviceHandlerFunc func(ctx context.Context, orgID uuid.UUID) (service.Service, error)
	callbackManager    tasks_client.CallbackManager
	log                logrus.FieldLogger
}

// enrichContext adds periodic-specific context values
func (b *baseExecutor) enrichContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, consts.EventSourceComponentCtxKey, "flightctl-periodic")
	ctx = context.WithValue(ctx, consts.EventActorCtxKey, "service:flightctl-periodic")
	ctx = context.WithValue(ctx, consts.InternalRequestCtxKey, true)
	return ctx
}

// RepoTestExecutor executes repository test tasks
type RepoTestExecutor struct {
	baseExecutor
}

func NewRepoTestExecutor(serviceHandlerFunc func(ctx context.Context, orgID uuid.UUID) (service.Service, error), log logrus.FieldLogger) *RepoTestExecutor {
	return &RepoTestExecutor{
		baseExecutor: baseExecutor{
			serviceHandlerFunc: serviceHandlerFunc,
			log:                log,
		},
	}
}

func (e *RepoTestExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	ctx = e.enrichContext(ctx)

	serviceHandler, err := e.serviceHandlerFunc(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}

	repoTester := tasks.NewRepoTester(e.log.WithField("org", orgID), serviceHandler)
	repoTester.TestRepositories(ctx)
	return nil
}

// ResourceSyncExecutor executes resource sync tasks
type ResourceSyncExecutor struct {
	baseExecutor
}

func NewResourceSyncExecutor(serviceHandlerFunc func(ctx context.Context, orgID uuid.UUID) (service.Service, error), callbackManager tasks_client.CallbackManager, log logrus.FieldLogger) *ResourceSyncExecutor {
	return &ResourceSyncExecutor{
		baseExecutor: baseExecutor{
			serviceHandlerFunc: serviceHandlerFunc,
			callbackManager:    callbackManager,
			log:                log,
		},
	}
}

func (e *ResourceSyncExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	ctx = e.enrichContext(ctx)

	serviceHandler, err := e.serviceHandlerFunc(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}

	resourceSync := tasks.NewResourceSync(e.callbackManager, serviceHandler, e.log.WithField("org", orgID))
	resourceSync.Poll(ctx)
	return nil
}

// DeviceDisconnectedExecutor executes device disconnected tasks
type DeviceDisconnectedExecutor struct {
	baseExecutor
}

func NewDeviceDisconnectedExecutor(serviceHandlerFunc func(ctx context.Context, orgID uuid.UUID) (service.Service, error), log logrus.FieldLogger) *DeviceDisconnectedExecutor {
	return &DeviceDisconnectedExecutor{
		baseExecutor: baseExecutor{
			serviceHandlerFunc: serviceHandlerFunc,
			log:                log,
		},
	}
}

func (e *DeviceDisconnectedExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	ctx = e.enrichContext(ctx)

	serviceHandler, err := e.serviceHandlerFunc(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}

	deviceDisconnected := tasks.NewDeviceDisconnected(e.log.WithField("org", orgID), serviceHandler)
	deviceDisconnected.Poll(ctx)
	return nil
}

// RolloutDeviceSelectionExecutor executes rollout device selection tasks
type RolloutDeviceSelectionExecutor struct {
	baseExecutor
}

func NewRolloutDeviceSelectionExecutor(serviceHandlerFunc func(ctx context.Context, orgID uuid.UUID) (service.Service, error), callbackManager tasks_client.CallbackManager, log logrus.FieldLogger) *RolloutDeviceSelectionExecutor {
	return &RolloutDeviceSelectionExecutor{
		baseExecutor: baseExecutor{
			serviceHandlerFunc: serviceHandlerFunc,
			callbackManager:    callbackManager,
			log:                log,
		},
	}
}

func (e *RolloutDeviceSelectionExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	ctx = e.enrichContext(ctx)

	serviceHandler, err := e.serviceHandlerFunc(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}

	rolloutDeviceSelection := device_selection.NewReconciler(serviceHandler, e.callbackManager, e.log.WithField("org", orgID))
	rolloutDeviceSelection.Reconcile(ctx)
	return nil
}

// DisruptionBudgetExecutor executes disruption budget tasks
type DisruptionBudgetExecutor struct {
	baseExecutor
}

func NewDisruptionBudgetExecutor(serviceHandlerFunc func(ctx context.Context, orgID uuid.UUID) (service.Service, error), callbackManager tasks_client.CallbackManager, log logrus.FieldLogger) *DisruptionBudgetExecutor {
	return &DisruptionBudgetExecutor{
		baseExecutor: baseExecutor{
			serviceHandlerFunc: serviceHandlerFunc,
			callbackManager:    callbackManager,
			log:                log,
		},
	}
}

func (e *DisruptionBudgetExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	ctx = e.enrichContext(ctx)

	serviceHandler, err := e.serviceHandlerFunc(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}

	disruptionBudget := disruption_budget.NewReconciler(serviceHandler, e.callbackManager, e.log.WithField("org", orgID))
	disruptionBudget.Reconcile(ctx)
	return nil
}

// EventCleanupExecutor executes event cleanup tasks
type EventCleanupExecutor struct {
	baseExecutor
	eventRetentionPeriod time.Duration
}

func NewEventCleanupExecutor(serviceHandlerFunc func(ctx context.Context, orgID uuid.UUID) (service.Service, error), eventRetentionPeriod time.Duration, log logrus.FieldLogger) *EventCleanupExecutor {
	return &EventCleanupExecutor{
		baseExecutor: baseExecutor{
			serviceHandlerFunc: serviceHandlerFunc,
			log:                log,
		},
		eventRetentionPeriod: eventRetentionPeriod,
	}
}

func (e *EventCleanupExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	ctx = e.enrichContext(ctx)

	serviceHandler, err := e.serviceHandlerFunc(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}

	eventCleanup := tasks.NewEventCleanup(e.log.WithField("org", orgID), serviceHandler, util.Duration(e.eventRetentionPeriod))
	eventCleanup.Poll(ctx)
	return nil
}

// CreateExecutors creates all task executors with their dependencies
func CreateExecutors(serviceHandlerFunc func(ctx context.Context, orgID uuid.UUID) (service.Service, error), callbackManager tasks_client.CallbackManager, eventRetentionPeriod time.Duration, log logrus.FieldLogger) map[tasks_runner.TaskType]tasks_runner.TaskExecutor {
	return map[tasks_runner.TaskType]tasks_runner.TaskExecutor{
		TaskTypeRepoTest:               NewRepoTestExecutor(serviceHandlerFunc, log),
		TaskTypeResourceSync:           NewResourceSyncExecutor(serviceHandlerFunc, callbackManager, log),
		TaskTypeDeviceDisconnected:     NewDeviceDisconnectedExecutor(serviceHandlerFunc, log),
		TaskTypeRolloutDeviceSelection: NewRolloutDeviceSelectionExecutor(serviceHandlerFunc, callbackManager, log),
		TaskTypeDisruptionBudget:       NewDisruptionBudgetExecutor(serviceHandlerFunc, callbackManager, log),
		TaskTypeEventCleanup:           NewEventCleanupExecutor(serviceHandlerFunc, eventRetentionPeriod, log),
	}
}

// GetTaskMetadata returns the task metadata for all periodic tasks
func GetTaskMetadata() []tasks_runner.TaskMetadata {
	return []tasks_runner.TaskMetadata{
		{TaskType: TaskTypeRepoTest, Interval: 2 * time.Minute},
		{TaskType: TaskTypeResourceSync, Interval: 2 * time.Minute},
		{TaskType: TaskTypeDeviceDisconnected, Interval: 2 * time.Minute},
		{TaskType: TaskTypeRolloutDeviceSelection, Interval: 30 * time.Second},
		{TaskType: TaskTypeDisruptionBudget, Interval: 30 * time.Second},
		{TaskType: TaskTypeEventCleanup, Interval: 10 * time.Minute},
	}
}
