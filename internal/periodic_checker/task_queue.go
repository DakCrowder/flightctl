package periodic

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskType string

const (
	TaskTypeRepoTest               TaskType = "repo_test"
	TaskTypeResourceSync           TaskType = "resource_sync"
	TaskTypeDeviceDisconnected     TaskType = "device_disconnected"
	TaskTypeRolloutDeviceSelection TaskType = "rollout_device_selection"
	TaskTypeDisruptionBudget       TaskType = "disruption_budget"
	TaskTypeEventCleanup           TaskType = "event_cleanup"
)

// Task represents a periodic task to be executed
type Task struct {
	ID          string
	OrgID       uuid.UUID
	Type        TaskType
	NextRunTime time.Time
	Interval    time.Duration
	index       int // Used by priority queue
}

// TaskQueue is a thread-safe priority queue for tasks
type TaskQueue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	heap   taskHeap
	tasks  map[string]*Task
	closed bool
}

func NewTaskQueue() *TaskQueue {
	tq := &TaskQueue{
		tasks: make(map[string]*Task),
	}
	tq.cond = sync.NewCond(&tq.mu)
	heap.Init(&tq.heap)
	return tq
}

// Push adds a task to the queue
func (tq *TaskQueue) Push(task *Task) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if tq.closed {
		return
	}

	// Remove existing task with same ID if present
	if existing, ok := tq.tasks[task.ID]; ok {
		heap.Remove(&tq.heap, existing.index)
	}

	tq.tasks[task.ID] = task
	heap.Push(&tq.heap, task)
	tq.cond.Signal()
}

// Pop removes and returns the next task to run, blocking if necessary
func (tq *TaskQueue) Pop(ctx context.Context) (*Task, error) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	for {
		if tq.closed {
			return nil, context.Canceled
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		now := time.Now()

		// Check if we have a task ready to run
		if tq.heap.Len() > 0 {
			task := tq.heap[0]
			if task.NextRunTime.Before(now) || task.NextRunTime.Equal(now) {
				heap.Pop(&tq.heap)
				delete(tq.tasks, task.ID)
				return task, nil
			}

			// Wait until the next task is ready or context is cancelled
			waitTime := task.NextRunTime.Sub(now)
			go func() {
				timer := time.NewTimer(waitTime)
				defer timer.Stop()
				select {
				case <-timer.C:
					tq.cond.Signal()
				case <-ctx.Done():
					tq.cond.Signal()
				}
			}()
		}

		// Wait for a signal (new task added, timer expired, or context cancelled)
		tq.cond.Wait()
	}
}

// RemoveOrganization removes all tasks for a given organization
func (tq *TaskQueue) RemoveOrganization(orgID uuid.UUID) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	var toRemove []*Task
	for _, task := range tq.tasks {
		if task.OrgID == orgID {
			toRemove = append(toRemove, task)
		}
	}

	for _, task := range toRemove {
		heap.Remove(&tq.heap, task.index)
		delete(tq.tasks, task.ID)
	}
}

func (tq *TaskQueue) Close() {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	tq.closed = true
	tq.cond.Broadcast()
}

// taskHeap implements heap.Interface for priority queue
type taskHeap []*Task

func (h taskHeap) Len() int { return len(h) }

func (h taskHeap) Less(i, j int) bool {
	return h[i].NextRunTime.Before(h[j].NextRunTime)
}

func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *taskHeap) Push(x interface{}) {
	n := len(*h)
	task := x.(*Task)
	task.index = n
	*h = append(*h, task)
}

func (h *taskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	task := old[n-1]
	old[n-1] = nil
	task.index = -1
	*h = old[0 : n-1]
	return task
}
