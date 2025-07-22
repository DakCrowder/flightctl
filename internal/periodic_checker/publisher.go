package periodic

import (
	"context"
	"time"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	"github.com/flightctl/flightctl/pkg/queues"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type OrganizationService interface {
	ListOrganizations(ctx context.Context) (*api.OrganizationList, api.Status)
}

type PeriodicTaskPublisher struct {
	publisher     queues.Publisher
	log           logrus.FieldLogger
	tasksMetadata []PeriodicTaskMetadata
	orgService    OrganizationService
	organizations map[uuid.UUID]bool // Track which organizations are registered
}

func NewPeriodicTaskPublisher(publisher queues.Publisher, log logrus.FieldLogger, orgService OrganizationService, tasksMetadata []PeriodicTaskMetadata) *PeriodicTaskPublisher {
	return &PeriodicTaskPublisher{
		publisher:     publisher,
		log:           log,
		orgService:    orgService,
		tasksMetadata: tasksMetadata,
	}
}

func (p *PeriodicTaskPublisher) publishTasks(ctx context.Context) {
	// TODO: Implement
}

func (p *PeriodicTaskPublisher) publishTask(ctx context.Context, taskType PeriodicTaskType, orgID uuid.UUID) {
	// TODO: Implement
}

func (p *PeriodicTaskPublisher) Start(ctx context.Context) {
	p.syncOrganizations(ctx)

	taskTicker := time.NewTicker(1 * time.Minute)
	defer taskTicker.Stop()

	// Update organizations every 5 minutes
	orgTicker := time.NewTicker(5 * time.Minute)
	defer orgTicker.Stop()

	for {
		select {
		case <-taskTicker.C:
			p.publishTasks(ctx)
		case <-orgTicker.C:
			p.syncOrganizations(ctx)
		case <-ctx.Done():
			p.stopAll()
			return
		}
	}
}

func (p *PeriodicTaskPublisher) syncOrganizations(ctx context.Context) {
	orgList, status := p.orgService.ListOrganizations(ctx)
	if status.Code != 200 {
		p.log.Errorf("Failed to list organizations: %v", status)
		return
	}

	// Track which organizations we've seen
	organizations := make(map[uuid.UUID]bool)

	for _, org := range orgList.Items {
		orgId, err := uuid.Parse(*org.Metadata.Name)
		if err != nil {
			p.log.Errorf("Failed to parse organization ID %s: %v", *org.Metadata.Name, err)
			continue
		}

		if !p.organizations[orgId] {
			p.log.Infof("Registering organization %s", orgId)
		}

		organizations[orgId] = true
	}

	for orgID := range p.organizations {
		if !organizations[orgID] {
			p.log.Infof("Organization %s is no longer registered, removing from tracking", orgID)
			delete(p.organizations, orgID)
		}
	}

	p.organizations = organizations
}

func (p *PeriodicTaskPublisher) stopAll() {
	// Clear all organizations
	p.organizations = make(map[uuid.UUID]bool)
}
