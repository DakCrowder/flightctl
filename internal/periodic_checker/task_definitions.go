package periodic

import (
	"time"

	"github.com/flightctl/flightctl/internal/rollout/device_selection"
	"github.com/flightctl/flightctl/internal/rollout/disruption_budget"
	"github.com/flightctl/flightctl/internal/tasks"
	"github.com/google/uuid"
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
