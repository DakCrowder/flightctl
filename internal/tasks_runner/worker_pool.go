package tasks_runner

import (
	"context"
	"sync"
	"time"

	"github.com/flightctl/flightctl/internal/consts"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// WorkerPool manages a fixed number of workers processing tasks
type WorkerPool struct {
	numWorkers    int
	taskQueue     *TaskQueue
	workers       []*Worker
	wg            sync.WaitGroup
	log           logrus.FieldLogger
	taskExecutors map[TaskType]TaskExecutor
}

// TaskExecutor executes a specific type of task
type TaskExecutor interface {
	Execute(ctx context.Context, orgID uuid.UUID) error
}

// Worker represents a single worker in the pool
type Worker struct {
	id   int
	pool *WorkerPool
	log  logrus.FieldLogger
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int, taskQueue *TaskQueue, log logrus.FieldLogger) *WorkerPool {
	return &WorkerPool{
		numWorkers:    numWorkers,
		taskQueue:     taskQueue,
		log:           log,
		taskExecutors: make(map[TaskType]TaskExecutor),
	}
}

// Start starts all workers in the pool
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.log.Infof("Starting worker pool with %d workers", wp.numWorkers)

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
