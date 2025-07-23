package periodic

import (
	"context"
	"time"

	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/internal/rollout/device_selection"
	"github.com/flightctl/flightctl/internal/rollout/disruption_budget"
	"github.com/flightctl/flightctl/internal/service"
	"github.com/flightctl/flightctl/internal/tasks"
	"github.com/flightctl/flightctl/internal/tasks_client"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/flightctl/flightctl/pkg/reqid"
	"github.com/go-chi/chi/v5/middleware"
)

type PeriodicTaskType string

const (
	PeriodicTaskTypeRepositoryTester       PeriodicTaskType = "repository-tester"
	PeriodicTaskTypeResourceSync           PeriodicTaskType = "resource-sync"
	PeriodicTaskTypeDeviceDisconnected     PeriodicTaskType = "device-disconnected"
	PeriodicTaskTypeRolloutDeviceSelection PeriodicTaskType = "rollout-device-selection"
	PeriodicTaskTypeDisruptionBudget       PeriodicTaskType = "disruption-budget"
	PeriodicTaskTypeEventCleanup           PeriodicTaskType = "event-cleanup"
)

type PeriodicTaskMetadata struct {
	TaskType PeriodicTaskType
	Interval time.Duration
}

var periodicTasks = []PeriodicTaskMetadata{
	{TaskType: PeriodicTaskTypeRepositoryTester, Interval: 2 * time.Minute},
	{TaskType: PeriodicTaskTypeResourceSync, Interval: 2 * time.Minute},
	{TaskType: PeriodicTaskTypeDeviceDisconnected, Interval: tasks.DeviceDisconnectedPollingInterval},
	{TaskType: PeriodicTaskTypeRolloutDeviceSelection, Interval: device_selection.RolloutDeviceSelectionInterval},
	{TaskType: PeriodicTaskTypeDisruptionBudget, Interval: disruption_budget.DisruptionBudgetReconcilationInterval},
	{TaskType: PeriodicTaskTypeEventCleanup, Interval: tasks.EventCleanupPollingInterval},
}

type PeriodicTaskReference struct {
	Type  PeriodicTaskType
	OrgID uuid.UUID
}

type PeriodicTaskLastRun struct {
	LastRun time.Time `json:"last_run"`
}

type PeriodicTaskExecutor interface {
	Execute(ctx context.Context, log logrus.FieldLogger)
}

type RepositoryTesterExecutor struct {
	repoTester *tasks.RepoTester
}

// createTaskContext creates a task context with request ID and event actor
func createTaskContext(ctx context.Context, taskType PeriodicTaskType) context.Context {
	taskName := string(taskType)
	reqid.OverridePrefix(taskName)
	requestID := reqid.NextRequestID()
	ctx = context.WithValue(ctx, middleware.RequestIDKey, requestID)

	return context.WithValue(ctx, consts.EventActorCtxKey, taskName)
}

func (e *RepositoryTesterExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log = log.WithField("pkg", "repository-tester")
	taskCtx := createTaskContext(ctx, PeriodicTaskTypeRepositoryTester)
	e.repoTester.TestRepositories(taskCtx)
}

type ResourceSyncExecutor struct {
	resourceSync *tasks.ResourceSync
}

func (e *ResourceSyncExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log = log.WithField("pkg", "resource-sync")
	taskCtx := createTaskContext(ctx, PeriodicTaskTypeResourceSync)
	e.resourceSync.Poll(taskCtx)
}

type DeviceDisconnectedExecutor struct {
	deviceDisconnected *tasks.DeviceDisconnected
}

func (e *DeviceDisconnectedExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log = log.WithField("pkg", "device-disconnected")
	taskCtx := createTaskContext(ctx, PeriodicTaskTypeDeviceDisconnected)
	e.deviceDisconnected.Poll(taskCtx)
}

type RolloutDeviceSelectionExecutor struct {
	rolloutDeviceSelection device_selection.Reconciler
}

func (e *RolloutDeviceSelectionExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log = log.WithField("pkg", "rollout-device-selection")
	taskCtx := createTaskContext(ctx, PeriodicTaskTypeRolloutDeviceSelection)
	e.rolloutDeviceSelection.Reconcile(taskCtx)
}

type DisruptionBudgetExecutor struct {
	disruptionBudget disruption_budget.Reconciler
}

func (e *DisruptionBudgetExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log = log.WithField("pkg", "disruption-budget")
	taskCtx := createTaskContext(ctx, PeriodicTaskTypeDisruptionBudget)
	e.disruptionBudget.Reconcile(taskCtx)
}

type EventCleanupExecutor struct {
	eventCleanup *tasks.EventCleanup
}

func (e *EventCleanupExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log = log.WithField("pkg", "event-cleanup")
	taskCtx := createTaskContext(ctx, PeriodicTaskTypeEventCleanup)
	e.eventCleanup.Poll(taskCtx)
}

func InitializeTaskExecutors(log logrus.FieldLogger, serviceHandler service.Service, callbackManager tasks_client.CallbackManager, eventRetentionPeriod util.Duration) map[PeriodicTaskType]PeriodicTaskExecutor {
	return map[PeriodicTaskType]PeriodicTaskExecutor{
		PeriodicTaskTypeRepositoryTester: &RepositoryTesterExecutor{
			repoTester: tasks.NewRepoTester(log, serviceHandler),
		},
		PeriodicTaskTypeResourceSync: &ResourceSyncExecutor{
			resourceSync: tasks.NewResourceSync(callbackManager, serviceHandler, log),
		},
		PeriodicTaskTypeDeviceDisconnected: &DeviceDisconnectedExecutor{
			deviceDisconnected: tasks.NewDeviceDisconnected(log, serviceHandler),
		},
		PeriodicTaskTypeRolloutDeviceSelection: &RolloutDeviceSelectionExecutor{
			rolloutDeviceSelection: device_selection.NewReconciler(serviceHandler, callbackManager, log),
		},
		PeriodicTaskTypeDisruptionBudget: &DisruptionBudgetExecutor{
			disruptionBudget: disruption_budget.NewReconciler(serviceHandler, callbackManager, log),
		},
		PeriodicTaskTypeEventCleanup: &EventCleanupExecutor{
			eventCleanup: tasks.NewEventCleanup(log, serviceHandler, eventRetentionPeriod),
		},
	}
}
