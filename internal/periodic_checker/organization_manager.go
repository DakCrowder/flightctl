package periodic

import (
	"context"
	"sync"
	"time"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/pkg/thread"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// OrganizationManager manages the lifecycle of organizations and their associated worker threads
type OrganizationManager struct {
	orgService    OrganizationService
	log           logrus.FieldLogger
	mu            sync.RWMutex
	organizations map[uuid.UUID]*OrganizationWorkers
	workerFactory WorkerFactory
}

// OrganizationWorkers holds all worker threads for a single organization
type OrganizationWorkers struct {
	orgId                        uuid.UUID
	repoTesterThread             *thread.Thread
	resourceSyncThread           *thread.Thread
	deviceDisconnectedThread     *thread.Thread
	rolloutDeviceSelectionThread *thread.Thread
	disruptionBudgetThread       *thread.Thread
	eventCleanupThread           *thread.Thread
	cancel                       context.CancelFunc
}

type WorkerFactory interface {
	CreateWorkers(ctx context.Context, orgId uuid.UUID) (*OrganizationWorkers, error)
}

// OrganizationService abstracts the organization listing functionality
type OrganizationService interface {
	ListOrganizations(ctx context.Context) (*api.OrganizationList, api.Status)
}

func NewOrganizationManager(orgService OrganizationService, log logrus.FieldLogger, workerFactory WorkerFactory) *OrganizationManager {
	return &OrganizationManager{
		orgService:    orgService,
		log:           log,
		organizations: make(map[uuid.UUID]*OrganizationWorkers),
		workerFactory: workerFactory,
	}
}

func (om *OrganizationManager) Start(ctx context.Context) {
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

	// Start workers for new organizations
	for _, org := range orgList.Items {
		orgId, err := uuid.Parse(*org.Metadata.Name)
		if err != nil {
			om.log.Errorf("Failed to parse organization ID %s: %v", *org.Metadata.Name, err)
			continue
		}
		seen[orgId] = true
		if _, exists := om.organizations[orgId]; !exists {
			om.log.Infof("Starting workers for organization %s", orgId)
			if err := om.startOrganizationWorkers(ctx, orgId); err != nil {
				om.log.Errorf("Failed to start workers for organization %s: %v", orgId, err)
			}
		}
	}

	// Stop workers for removed organizations
	for orgId, workers := range om.organizations {
		if !seen[orgId] {
			om.log.Infof("Stopping workers for organization %s", orgId)
			om.stopOrganizationWorkers(orgId, workers)
			delete(om.organizations, orgId)
		}
	}
}

func (om *OrganizationManager) startOrganizationWorkers(ctx context.Context, orgId uuid.UUID) error {
	orgCtx, cancel := context.WithCancel(ctx)

	workers, err := om.workerFactory.CreateWorkers(orgCtx, orgId)
	if err != nil {
		cancel()
		return err
	}

	workers.cancel = cancel
	om.organizations[orgId] = workers
	return nil
}

func (om *OrganizationManager) stopOrganizationWorkers(orgId uuid.UUID, workers *OrganizationWorkers) {
	om.log.Infof("Stopping workers for organization %s", orgId)

	// Cancel the context
	if workers.cancel != nil {
		workers.cancel()
	}

	// Stop all threads
	if workers.repoTesterThread != nil {
		workers.repoTesterThread.Stop()
	}
	if workers.resourceSyncThread != nil {
		workers.resourceSyncThread.Stop()
	}
	if workers.deviceDisconnectedThread != nil {
		workers.deviceDisconnectedThread.Stop()
	}
	if workers.rolloutDeviceSelectionThread != nil {
		workers.rolloutDeviceSelectionThread.Stop()
	}
	if workers.disruptionBudgetThread != nil {
		workers.disruptionBudgetThread.Stop()
	}
	if workers.eventCleanupThread != nil {
		workers.eventCleanupThread.Stop()
	}
}

func (om *OrganizationManager) stopAll() {
	om.mu.Lock()
	defer om.mu.Unlock()

	for orgId, workers := range om.organizations {
		om.stopOrganizationWorkers(orgId, workers)
	}
	om.organizations = make(map[uuid.UUID]*OrganizationWorkers)
}
