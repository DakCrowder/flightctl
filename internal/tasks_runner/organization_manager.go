package tasks_runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TaskQueueInterface interface {
	Push(task *Task)
	Pop(ctx context.Context) (*Task, error)
	RemoveOrganization(orgID uuid.UUID)
	Close()
}

type WorkerPoolInterface interface {
	Start(ctx context.Context)
	Stop()
}

type OrganizationManager struct {
	orgService    OrganizationService
	log           logrus.FieldLogger
	mu            sync.RWMutex
	organizations map[uuid.UUID]bool // Track which organizations are registered
	taskQueue     TaskQueueInterface
	workerPool    WorkerPoolInterface
	tasksMetadata []TaskMetadata
}

type OrganizationService interface {
	ListOrganizations(ctx context.Context) (*api.OrganizationList, api.Status)
}

type TaskMetadata struct {
	TaskType TaskType
	Interval time.Duration
}

func NewOrganizationManager(orgService OrganizationService, log logrus.FieldLogger, taskQueue TaskQueueInterface, workerPool WorkerPoolInterface, tasksMeta []TaskMetadata) *OrganizationManager {
	return &OrganizationManager{
		orgService:    orgService,
		log:           log,
		organizations: make(map[uuid.UUID]bool),
		taskQueue:     taskQueue,
		workerPool:    workerPool,
		tasksMetadata: tasksMeta,
	}
}

func (om *OrganizationManager) Start(ctx context.Context) {
	// Start the worker pool
	om.workerPool.Start(ctx)

	om.syncOrganizations(ctx)

	// Update organizations every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			om.syncOrganizations(ctx)
		case <-ctx.Done():
			om.stopAll()
			return
		}
	}
}

func (om *OrganizationManager) syncOrganizations(ctx context.Context) {
	orgList, status := om.orgService.ListOrganizations(ctx)
	if status.Code != 200 {
		om.log.Errorf("Failed to list organizations: %v", status)
		return
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	// Track which organizations we've seen
	seen := make(map[uuid.UUID]bool)

	// Register tasks for new organizations
	for _, org := range orgList.Items {
		orgId, err := uuid.Parse(*org.Metadata.Name)
		if err != nil {
			om.log.Errorf("Failed to parse organization ID %s: %v", *org.Metadata.Name, err)
			continue
		}
		seen[orgId] = true
		if !om.organizations[orgId] {
			om.log.Infof("Registering tasks for organization %s", orgId)
			if err := om.registerOrganizationTasks(ctx, orgId); err != nil {
				om.log.Errorf("Failed to register tasks for organization %s: %v", orgId, err)
			} else {
				om.organizations[orgId] = true
			}
		}
	}

	// Remove tasks for deleted organizations
	for orgId := range om.organizations {
		if !seen[orgId] {
			om.log.Infof("Removing tasks for organization %s", orgId)
			om.removeOrganizationTasks(orgId)
			delete(om.organizations, orgId)
		}
	}
}

func (om *OrganizationManager) registerOrganizationTasks(ctx context.Context, orgId uuid.UUID) error {
	for _, taskMeta := range om.tasksMetadata {
		task := &Task{
			ID:          fmt.Sprintf("%s-%s", orgId, taskMeta.TaskType),
			OrgID:       orgId,
			Type:        taskMeta.TaskType,
			NextRunTime: time.Now(), // Start immediately
			Interval:    taskMeta.Interval,
		}
		om.taskQueue.Push(task)
	}

	return nil
}

func (om *OrganizationManager) removeOrganizationTasks(orgId uuid.UUID) {
	om.taskQueue.RemoveOrganization(orgId)
}

func (om *OrganizationManager) stopAll() {
	om.mu.Lock()
	defer om.mu.Unlock()

	// Stop the worker pool
	om.workerPool.Stop()

	// Clear all organizations
	om.organizations = make(map[uuid.UUID]bool)
}
