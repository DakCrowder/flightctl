package periodic

import (
	"context"
	"time"

	"github.com/flightctl/flightctl/internal/rollout/device_selection"
	"github.com/flightctl/flightctl/internal/rollout/disruption_budget"
	"github.com/flightctl/flightctl/internal/service"
	"github.com/flightctl/flightctl/internal/tasks"
	"github.com/flightctl/flightctl/internal/tasks_client"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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

func (e *RepositoryTesterExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log.WithField("pkg", "repository-tester")
	e.repoTester.TestRepositories(ctx)
}

type ResourceSyncExecutor struct {
	resourceSync *tasks.ResourceSync
}

func (e *ResourceSyncExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log.WithField("pkg", "resource-sync")
	e.resourceSync.Poll(ctx)
}

type DeviceDisconnectedExecutor struct {
	deviceDisconnected *tasks.DeviceDisconnected
}

func (e *DeviceDisconnectedExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log.WithField("pkg", "device-disconnected")
	e.deviceDisconnected.Poll(ctx)
}

type RolloutDeviceSelectionExecutor struct {
	rolloutDeviceSelection device_selection.Reconciler
}

func (e *RolloutDeviceSelectionExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log.WithField("pkg", "rollout-device-selection")
	e.rolloutDeviceSelection.Reconcile(ctx)
}

type DisruptionBudgetExecutor struct {
	disruptionBudget disruption_budget.Reconciler
}

func (e *DisruptionBudgetExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log.WithField("pkg", "disruption-budget")
	e.disruptionBudget.Reconcile(ctx)
}

type EventCleanupExecutor struct {
	eventCleanup *tasks.EventCleanup
}

func (e *EventCleanupExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	log.WithField("pkg", "event-cleanup")
	e.eventCleanup.Poll(ctx)
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
