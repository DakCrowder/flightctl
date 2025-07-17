package periodic

import (
	"context"
	"fmt"
	"sync"
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
)

// WorkerPool manages a fixed number of workers processing tasks
type WorkerPool struct {
	numWorkers      int
	taskQueue       *TaskQueue
	workers         []*Worker
	wg              sync.WaitGroup
	log             logrus.FieldLogger
	taskExecutors   map[TaskType]TaskExecutor
	serviceProvider ServiceProvider
}

// TaskExecutor executes a specific type of task
type TaskExecutor interface {
	Execute(ctx context.Context, orgID uuid.UUID) error
}

// ServiceProvider provides services needed by task executors
type ServiceProvider interface {
	GetServiceHandler(ctx context.Context, orgID uuid.UUID) (service.Service, error)
	GetCallbackManager() tasks_client.CallbackManager
	GetConfig() interface{}
}

// Worker represents a single worker in the pool
type Worker struct {
	id              int
	pool            *WorkerPool
	log             logrus.FieldLogger
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int, taskQueue *TaskQueue, serviceProvider ServiceProvider, log logrus.FieldLogger) *WorkerPool {
	return &WorkerPool{
		numWorkers:      numWorkers,
		taskQueue:       taskQueue,
		log:             log,
		taskExecutors:   make(map[TaskType]TaskExecutor),
		serviceProvider: serviceProvider,
	}
}

// RegisterExecutor registers a task executor for a specific task type
func (wp *WorkerPool) RegisterExecutor(taskType TaskType, executor TaskExecutor) {
	wp.taskExecutors[taskType] = executor
}

// Start starts all workers in the pool
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.log.Infof("Starting worker pool with %d workers", wp.numWorkers)
	
	// Register all task executors
	wp.registerDefaultExecutors()
	
	// Create and start workers
	for i := 0; i < wp.numWorkers; i++ {
		worker := &Worker{
			id:   i,
			pool: wp,
			log:  wp.log.WithField("worker_id", i),
		}
		wp.workers = append(wp.workers, worker)
		
		wp.wg.Add(1)
		go worker.run(ctx)
	}
}

// Stop stops all workers and waits for them to finish
func (wp *WorkerPool) Stop() {
	wp.log.Info("Stopping worker pool")
	wp.taskQueue.Close()
	wp.wg.Wait()
	wp.log.Info("Worker pool stopped")
}

// run is the main loop for a worker
func (w *Worker) run(ctx context.Context) {
	defer w.pool.wg.Done()
	
	w.log.Info("Worker started")
	
	for {
		select {
		case <-ctx.Done():
			w.log.Info("Worker stopping due to context cancellation")
			return
		default:
			// Get next task from queue
			task, err := w.pool.taskQueue.Pop(ctx)
			if err != nil {
				if err == context.Canceled {
					w.log.Info("Worker stopping")
					return
				}
				w.log.Errorf("Error getting task from queue: %v", err)
				continue
			}
			
			// Execute the task
			w.executeTask(ctx, task)
		}
	}
}

// executeTask executes a single task
func (w *Worker) executeTask(ctx context.Context, task *Task) {
	w.log.WithFields(logrus.Fields{
		"task_id":   task.ID,
		"task_type": task.Type,
		"org_id":    task.OrgID,
	}).Debug("Executing task")
	
	// Create context with organization ID
	orgCtx := context.WithValue(ctx, consts.OrganizationIDCtxKey, task.OrgID)
	orgCtx = context.WithValue(orgCtx, consts.EventSourceComponentCtxKey, "flightctl-periodic")
	orgCtx = context.WithValue(orgCtx, consts.EventActorCtxKey, "service:flightctl-periodic")
	orgCtx = context.WithValue(orgCtx, consts.InternalRequestCtxKey, true)
	
	// Get executor for this task type
	executor, ok := w.pool.taskExecutors[task.Type]
	if !ok {
		w.log.Errorf("No executor registered for task type: %s", task.Type)
		return
	}
	
	// Execute the task
	startTime := time.Now()
	err := executor.Execute(orgCtx, task.OrgID)
	duration := time.Since(startTime)
	
	if err != nil {
		w.log.WithError(err).WithFields(logrus.Fields{
			"task_id":   task.ID,
			"task_type": task.Type,
			"org_id":    task.OrgID,
			"duration":  duration,
		}).Error("Task execution failed")
	} else {
		w.log.WithFields(logrus.Fields{
			"task_id":   task.ID,
			"task_type": task.Type,
			"org_id":    task.OrgID,
			"duration":  duration,
		}).Debug("Task execution completed")
	}
	
	// Reschedule the task
	task.NextRunTime = time.Now().Add(task.Interval)
	w.pool.taskQueue.Push(task)
}

// registerDefaultExecutors registers the default task executors
func (wp *WorkerPool) registerDefaultExecutors() {
	// Repository tester
	wp.RegisterExecutor(TaskTypeRepoTest, &repoTestExecutor{
		serviceProvider: wp.serviceProvider,
		log:             wp.log,
	})
	
	// Resource sync
	wp.RegisterExecutor(TaskTypeResourceSync, &resourceSyncExecutor{
		serviceProvider: wp.serviceProvider,
		log:             wp.log,
	})
	
	// Device disconnected
	wp.RegisterExecutor(TaskTypeDeviceDisconnected, &deviceDisconnectedExecutor{
		serviceProvider: wp.serviceProvider,
		log:             wp.log,
	})
	
	// Rollout device selection
	wp.RegisterExecutor(TaskTypeRolloutDeviceSelection, &rolloutDeviceSelectionExecutor{
		serviceProvider: wp.serviceProvider,
		log:             wp.log,
	})
	
	// Disruption budget
	wp.RegisterExecutor(TaskTypeDisruptionBudget, &disruptionBudgetExecutor{
		serviceProvider: wp.serviceProvider,
		log:             wp.log,
	})
	
	// Event cleanup
	wp.RegisterExecutor(TaskTypeEventCleanup, &eventCleanupExecutor{
		serviceProvider: wp.serviceProvider,
		log:             wp.log,
	})
}

// Task executor implementations

type repoTestExecutor struct {
	serviceProvider ServiceProvider
	log             logrus.FieldLogger
}

func (e *repoTestExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	serviceHandler, err := e.serviceProvider.GetServiceHandler(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}
	
	repoTester := tasks.NewRepoTester(e.log.WithField("org", orgID), serviceHandler)
	repoTester.TestRepositories(ctx)
	return nil
}

type resourceSyncExecutor struct {
	serviceProvider ServiceProvider
	log             logrus.FieldLogger
}

func (e *resourceSyncExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	serviceHandler, err := e.serviceProvider.GetServiceHandler(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}
	
	resourceSync := tasks.NewResourceSync(e.serviceProvider.GetCallbackManager(), serviceHandler, e.log.WithField("org", orgID))
	resourceSync.Poll(ctx)
	return nil
}

type deviceDisconnectedExecutor struct {
	serviceProvider ServiceProvider
	log             logrus.FieldLogger
}

func (e *deviceDisconnectedExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	serviceHandler, err := e.serviceProvider.GetServiceHandler(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}
	
	deviceDisconnected := tasks.NewDeviceDisconnected(e.log.WithField("org", orgID), serviceHandler)
	deviceDisconnected.Poll(ctx)
	return nil
}

type rolloutDeviceSelectionExecutor struct {
	serviceProvider ServiceProvider
	log             logrus.FieldLogger
}

func (e *rolloutDeviceSelectionExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	serviceHandler, err := e.serviceProvider.GetServiceHandler(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}
	
	rolloutDeviceSelection := device_selection.NewReconciler(serviceHandler, e.serviceProvider.GetCallbackManager(), e.log.WithField("org", orgID))
	rolloutDeviceSelection.Reconcile(ctx)
	return nil
}

type disruptionBudgetExecutor struct {
	serviceProvider ServiceProvider
	log             logrus.FieldLogger
}

func (e *disruptionBudgetExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	serviceHandler, err := e.serviceProvider.GetServiceHandler(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}
	
	disruptionBudget := disruption_budget.NewReconciler(serviceHandler, e.serviceProvider.GetCallbackManager(), e.log.WithField("org", orgID))
	disruptionBudget.Reconcile(ctx)
	return nil
}

type eventCleanupExecutor struct {
	serviceProvider ServiceProvider
	log             logrus.FieldLogger
}

func (e *eventCleanupExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	serviceHandler, err := e.serviceProvider.GetServiceHandler(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get service handler: %w", err)
	}
	
	// Get config to access event retention period
	cfg := e.serviceProvider.GetConfig()
	eventRetention := cfg.(*configWrapper).eventRetentionPeriod
	eventCleanup := tasks.NewEventCleanup(e.log.WithField("org", orgID), serviceHandler, util.Duration(eventRetention))
	eventCleanup.Poll(ctx)
	return nil
}

// configWrapper wraps config for event cleanup
type configWrapper struct {
	eventRetentionPeriod time.Duration
}