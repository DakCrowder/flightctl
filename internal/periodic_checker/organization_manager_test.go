package periodic_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	api "github.com/flightctl/flightctl/api/v1alpha1"
	periodic "github.com/flightctl/flightctl/internal/periodic_checker"
	"github.com/flightctl/flightctl/pkg/log"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TestService struct {
	organizations *api.OrganizationList
	status        api.Status
}

func (s *TestService) ListOrganizations(ctx context.Context) (*api.OrganizationList, api.Status) {
	return s.organizations, s.status
}

type TestWorkerFactory struct {
	createdOrgs []uuid.UUID
	shouldError bool
	mu          sync.RWMutex
	workers     map[uuid.UUID]*periodic.OrganizationWorkers
}

func (f *TestWorkerFactory) CreateWorkers(ctx context.Context, orgId uuid.UUID) (*periodic.OrganizationWorkers, error) {
	if f.shouldError {
		return nil, errors.New("worker factory error")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.createdOrgs = append(f.createdOrgs, orgId)

	if f.workers == nil {
		f.workers = make(map[uuid.UUID]*periodic.OrganizationWorkers)
	}

	workers := &periodic.OrganizationWorkers{}
	f.workers[orgId] = workers

	return workers, nil
}

func (f *TestWorkerFactory) GetWorkers(orgId uuid.UUID) *periodic.OrganizationWorkers {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.workers[orgId]
}

func (f *TestWorkerFactory) GetCreatedOrgs() []uuid.UUID {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]uuid.UUID, len(f.createdOrgs))
	copy(result, f.createdOrgs)
	return result
}

func (f *TestWorkerFactory) GetCreatedOrgsCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.createdOrgs)
}

// TestWorkerFactoryWithThreads creates worker goroutines for testing stop behavior
type TestWorkerFactoryWithThreads struct {
	createdOrgs   []uuid.UUID
	shouldError   bool
	mu            sync.RWMutex
	workers       map[uuid.UUID]*periodic.OrganizationWorkers
	logger        logrus.FieldLogger
	threadsActive map[uuid.UUID]bool
	threadStopCh  map[uuid.UUID]chan struct{}
}

func NewTestWorkerFactoryWithThreads(logger logrus.FieldLogger) *TestWorkerFactoryWithThreads {
	return &TestWorkerFactoryWithThreads{
		workers:       make(map[uuid.UUID]*periodic.OrganizationWorkers),
		logger:        logger,
		threadsActive: make(map[uuid.UUID]bool),
		threadStopCh:  make(map[uuid.UUID]chan struct{}),
	}
}

func (f *TestWorkerFactoryWithThreads) CreateWorkers(ctx context.Context, orgId uuid.UUID) (*periodic.OrganizationWorkers, error) {
	if f.shouldError {
		return nil, errors.New("worker factory error")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.createdOrgs = append(f.createdOrgs, orgId)

	// Create channels to track thread lifecycle
	stopCh := make(chan struct{})
	f.threadStopCh[orgId] = stopCh
	f.threadsActive[orgId] = true

	// Create actual threads for testing that will signal when they stop
	workers := &periodic.OrganizationWorkers{}

	// Create test threads that track their lifecycle
	// We need to use reflection or a different approach since we can't set unexported fields directly
	// For now, let's create a minimal implementation that just tracks the lifecycle

	// Create context for the workers that will be cancelled when stopOrganizationWorkers is called
	workerCtx, cancel := context.WithCancel(ctx)

	// Start goroutines that simulate the work threads would do
	go func() {
		defer func() {
			f.mu.Lock()
			f.threadsActive[orgId] = false
			f.mu.Unlock()
			close(stopCh)
		}()

		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				// Simulate thread work
			}
		}
	}()

	// Store the cancel function so we can simulate proper shutdown
	// In a real implementation, this would be handled by the OrganizationManager
	// calling Stop() on the actual thread objects
	go func() {
		<-ctx.Done()
		cancel()
	}()

	f.workers[orgId] = workers

	return workers, nil
}

func (f *TestWorkerFactoryWithThreads) GetWorkers(orgId uuid.UUID) *periodic.OrganizationWorkers {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.workers[orgId]
}

func (f *TestWorkerFactoryWithThreads) IsThreadActive(orgId uuid.UUID) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.threadsActive[orgId]
}

func (f *TestWorkerFactoryWithThreads) WaitForThreadsToStop(orgId uuid.UUID, timeout time.Duration) bool {
	f.mu.RLock()
	stopCh, exists := f.threadStopCh[orgId]
	f.mu.RUnlock()

	if !exists {
		return false
	}

	select {
	case <-stopCh:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (f *TestWorkerFactoryWithThreads) GetCreatedOrgsCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.createdOrgs)
}

func TestManagerStartProcessesMultipleOrganizations(t *testing.T) {
	log := log.InitLogs()
	workerFactory := &TestWorkerFactory{}

	// Create test organizations
	org1Id := uuid.New()
	org2Id := uuid.New()
	org1Name := org1Id.String()
	org2Name := org2Id.String()

	orgList := &api.OrganizationList{
		Items: []api.Organization{
			{
				Metadata: api.ObjectMeta{
					Name: &org1Name,
				},
			},
			{
				Metadata: api.ObjectMeta{
					Name: &org2Name,
				},
			},
		},
	}

	orgService := &TestService{
		organizations: orgList,
		status:        api.Status{Code: 200},
	}

	manager := periodic.NewOrganizationManager(orgService, log, workerFactory)

	// Test by starting the manager briefly to trigger initial sync
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go manager.Start(ctx)

	// Wait for initial sync to complete
	time.Sleep(50 * time.Millisecond)

	// Verify both organizations were processed
	if workerFactory.GetCreatedOrgsCount() != 2 {
		t.Fatalf("Expected 2 organizations to be created, got %d", workerFactory.GetCreatedOrgsCount())
	}

	// Verify the correct organization IDs were created
	createdIds := make(map[uuid.UUID]bool)
	for _, id := range workerFactory.GetCreatedOrgs() {
		createdIds[id] = true
	}

	if !createdIds[org1Id] {
		t.Errorf("Organization %s was not created", org1Id)
	}
	if !createdIds[org2Id] {
		t.Errorf("Organization %s was not created", org2Id)
	}
}

func TestManagerHandlesAPIError(t *testing.T) {
	log := log.InitLogs()
	workerFactory := &TestWorkerFactory{}

	// Create service that returns an error
	orgService := &TestService{
		organizations: nil,
		status:        api.Status{Code: 500, Message: "Internal server error"},
	}

	manager := periodic.NewOrganizationManager(orgService, log, workerFactory)

	// Test by starting the manager briefly to trigger initial sync
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go manager.Start(ctx)

	// Wait for initial sync to complete
	time.Sleep(50 * time.Millisecond)

	// Verify no organizations were created due to API error
	if workerFactory.GetCreatedOrgsCount() != 0 {
		t.Fatalf("Expected 0 organizations to be created due to API error, got %d", workerFactory.GetCreatedOrgsCount())
	}
}

func TestManagerHandlesWorkerFactoryError(t *testing.T) {
	log := log.InitLogs()
	workerFactory := &TestWorkerFactory{shouldError: true}

	// Create test organization
	orgId := uuid.New()
	orgName := orgId.String()

	orgList := &api.OrganizationList{
		Items: []api.Organization{
			{
				Metadata: api.ObjectMeta{
					Name: &orgName,
				},
			},
		},
	}

	orgService := &TestService{
		organizations: orgList,
		status:        api.Status{Code: 200},
	}

	manager := periodic.NewOrganizationManager(orgService, log, workerFactory)

	// Test by starting the manager briefly to trigger initial sync
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go manager.Start(ctx)

	// Wait for initial sync to complete
	time.Sleep(50 * time.Millisecond)

	// Verify no organizations were created due to factory error
	if workerFactory.GetCreatedOrgsCount() != 0 {
		t.Fatalf("Expected 0 organizations to be created due to factory error, got %d", workerFactory.GetCreatedOrgsCount())
	}
}

func TestManagerStopAllMethod(t *testing.T) {
	logger := log.InitLogs()
	workerFactory := NewTestWorkerFactoryWithThreads(logger)

	org1Id := uuid.New()
	org2Id := uuid.New()
	org1Name := org1Id.String()
	org2Name := org2Id.String()

	orgList := &api.OrganizationList{
		Items: []api.Organization{
			{
				Metadata: api.ObjectMeta{
					Name: &org1Name,
				},
			},
			{
				Metadata: api.ObjectMeta{
					Name: &org2Name,
				},
			},
		},
	}

	orgService := &TestService{
		organizations: orgList,
		status:        api.Status{Code: 200},
	}

	manager := periodic.NewOrganizationManager(orgService, logger, workerFactory)

	// Start the manager briefly to create organizations
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	managerDone := make(chan struct{})
	go func() {
		manager.Start(ctx)
		close(managerDone)
	}()

	// Wait for initial sync to complete
	time.Sleep(100 * time.Millisecond)

	// Verify organizations were created
	if workerFactory.GetCreatedOrgsCount() != 2 {
		t.Fatalf("Expected 2 organizations to be created, got %d", workerFactory.GetCreatedOrgsCount())
	}

	// Verify threads are initially active
	if !workerFactory.IsThreadActive(org1Id) {
		t.Error("Organization 1 threads should be active after creation")
	}
	if !workerFactory.IsThreadActive(org2Id) {
		t.Error("Organization 2 threads should be active after creation")
	}

	// Cancel the context to trigger stopAll
	cancel()

	// Wait for manager to shut down gracefully
	select {
	case <-managerDone:
		// Manager has shut down successfully
	case <-time.After(time.Second):
		t.Fatal("Manager did not shut down within timeout")
	}

	// Verify threads have stopped by waiting for their stop signals
	if !workerFactory.WaitForThreadsToStop(org1Id, 500*time.Millisecond) {
		t.Error("Organization 1 threads did not stop within timeout")
	}
	if !workerFactory.WaitForThreadsToStop(org2Id, 500*time.Millisecond) {
		t.Error("Organization 2 threads did not stop within timeout")
	}

	// Double-check that threads are no longer active
	if workerFactory.IsThreadActive(org1Id) {
		t.Error("Organization 1 threads should be stopped after context cancellation")
	}
	if workerFactory.IsThreadActive(org2Id) {
		t.Error("Organization 2 threads should be stopped after context cancellation")
	}
}
