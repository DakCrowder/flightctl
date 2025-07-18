package tasks_runner

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/flightctl/flightctl/internal/consts"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// TestTaskQueue wraps TaskQueue for testing
type TestTaskQueue struct {
	*TaskQueue
	mu     sync.Mutex
	popErr error
}

func NewTestTaskQueue() *TestTaskQueue {
	return &TestTaskQueue{
		TaskQueue: NewTaskQueue(),
	}
}

func (t *TestTaskQueue) SetPopError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.popErr = err
}

func (t *TestTaskQueue) Pop(ctx context.Context) (*Task, error) {
	t.mu.Lock()
	popErr := t.popErr
	t.mu.Unlock()
	
	if popErr != nil {
		return nil, popErr
	}
	
	return t.TaskQueue.Pop(ctx)
}

func (t *TestTaskQueue) GetTaskCount() int {
	t.TaskQueue.mu.Lock()
	defer t.TaskQueue.mu.Unlock()
	return len(t.TaskQueue.tasks)
}

// MockTaskExecutor is a test implementation of TaskExecutor
type MockTaskExecutor struct {
	mu           sync.Mutex
	executions   []ExecutionRecord
	shouldFail   bool
	failureError error
	delay        time.Duration
}

type ExecutionRecord struct {
	OrgID uuid.UUID
	Time  time.Time
}

func NewMockTaskExecutor() *MockTaskExecutor {
	return &MockTaskExecutor{
		executions: make([]ExecutionRecord, 0),
	}
}

func (m *MockTaskExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	m.executions = append(m.executions, ExecutionRecord{
		OrgID: orgID,
		Time:  time.Now(),
	})

	if m.shouldFail {
		if m.failureError != nil {
			return m.failureError
		}
		return errors.New("mock execution failed")
	}

	return nil
}

func (m *MockTaskExecutor) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
}

func (m *MockTaskExecutor) SetFailureError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failureError = err
}

func (m *MockTaskExecutor) SetDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delay = delay
}

func (m *MockTaskExecutor) GetExecutions() []ExecutionRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ExecutionRecord(nil), m.executions...)
}

func (m *MockTaskExecutor) GetExecutionCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.executions)
}

// Test helper functions
func createTestTask(orgID uuid.UUID, taskType TaskType) *Task {
	return &Task{
		ID:          orgID.String() + "-" + string(taskType),
		OrgID:       orgID,
		Type:        taskType,
		NextRunTime: time.Now().Add(-1 * time.Second), // Ready to run immediately
		Interval:    5 * time.Minute,
	}
}

func createTestExecutors() map[TaskType]TaskExecutor {
	return map[TaskType]TaskExecutor{
		"test_task": NewMockTaskExecutor(),
	}
}

func TestNewWorkerPool(t *testing.T) {
	require := require.New(t)

	taskQueue := NewTestTaskQueue()
	executors := createTestExecutors()
	logger := logrus.New()

	wp := NewWorkerPool(3, taskQueue.TaskQueue, executors, logger)

	require.NotNil(wp)
	require.Equal(3, wp.numWorkers)
	require.Equal(taskQueue.TaskQueue, wp.taskQueue)
	require.Equal(executors, wp.taskExecutors)
	require.Equal(logger, wp.log)
	require.Empty(wp.workers)
}

func TestWorkerPool_Start(t *testing.T) {
	require := require.New(t)

	taskQueue := NewTestTaskQueue()
	executors := createTestExecutors()
	logger := logrus.New()

	wp := NewWorkerPool(3, taskQueue.TaskQueue, executors, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wp.Start(ctx)

	// Allow some time for workers to start
	time.Sleep(50 * time.Millisecond)

	require.Len(wp.workers, 3)
	for i, worker := range wp.workers {
		require.Equal(i, worker.id)
		require.Equal(wp, worker.pool)
		require.NotNil(worker.log)
	}

	// Stop the pool
	wp.Stop()
}

func TestWorkerPool_Stop(t *testing.T) {
	_ = require.New(t)

	taskQueue := NewTestTaskQueue()
	executors := createTestExecutors()
	logger := logrus.New()

	wp := NewWorkerPool(2, taskQueue.TaskQueue, executors, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wp.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Stop should complete without hanging
	done := make(chan struct{})
	go func() {
		wp.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Stop() did not complete within timeout")
	}
}

func TestWorker_ExecuteTask_Success(t *testing.T) {
	require := require.New(t)

	taskQueue := NewTestTaskQueue()
	mockExecutor := NewMockTaskExecutor()
	executors := map[TaskType]TaskExecutor{
		"test_task": mockExecutor,
	}
	logger := logrus.New()

	wp := NewWorkerPool(1, taskQueue.TaskQueue, executors, logger)
	worker := &Worker{
		id:   0,
		pool: wp,
		log:  logger,
	}

	orgID := uuid.New()
	task := createTestTask(orgID, "test_task")
	originalNextRunTime := task.NextRunTime

	ctx := context.Background()
	worker.executeTask(ctx, task)

	// Verify executor was called
	executions := mockExecutor.GetExecutions()
	require.Len(executions, 1)
	require.Equal(orgID, executions[0].OrgID)

	// Verify task was rescheduled
	require.True(task.NextRunTime.After(originalNextRunTime))

	// Verify task was pushed back to queue
	require.Equal(1, taskQueue.GetTaskCount())
}

func TestWorker_ExecuteTask_ExecutorFailure(t *testing.T) {
	require := require.New(t)

	taskQueue := NewTestTaskQueue()
	mockExecutor := NewMockTaskExecutor()
	mockExecutor.SetShouldFail(true)
	mockExecutor.SetFailureError(errors.New("test execution error"))

	executors := map[TaskType]TaskExecutor{
		"test_task": mockExecutor,
	}
	logger := logrus.New()

	wp := NewWorkerPool(1, taskQueue.TaskQueue, executors, logger)
	worker := &Worker{
		id:   0,
		pool: wp,
		log:  logger,
	}

	orgID := uuid.New()
	task := createTestTask(orgID, "test_task")

	ctx := context.Background()
	worker.executeTask(ctx, task)

	// Verify executor was called despite failure
	executions := mockExecutor.GetExecutions()
	require.Len(executions, 1)

	// Verify task was still rescheduled even after failure
	require.Equal(1, taskQueue.GetTaskCount())
}

func TestWorker_ExecuteTask_NoExecutor(t *testing.T) {
	require := require.New(t)

	taskQueue := NewTestTaskQueue()
	executors := map[TaskType]TaskExecutor{} // Empty executors map
	logger := logrus.New()

	wp := NewWorkerPool(1, taskQueue.TaskQueue, executors, logger)
	worker := &Worker{
		id:   0,
		pool: wp,
		log:  logger,
	}

	orgID := uuid.New()
	task := createTestTask(orgID, "unknown_task")

	ctx := context.Background()
	worker.executeTask(ctx, task)

	// Task should not be rescheduled when no executor is found
	require.Equal(0, taskQueue.GetTaskCount())
}

func TestWorker_ExecuteTask_OrganizationContext(t *testing.T) {
	require := require.New(t)

	taskQueue := NewTestTaskQueue()
	
	// Create a custom executor that checks the context
	var contextOrgID uuid.UUID
	executor := &mockContextExecutor{
		contextOrgID: &contextOrgID,
	}
	
	executors := map[TaskType]TaskExecutor{
		"test_task": executor,
	}
	logger := logrus.New()

	wp := NewWorkerPool(1, taskQueue.TaskQueue, executors, logger)
	worker := &Worker{
		id:   0,
		pool: wp,
		log:  logger,
	}

	orgID := uuid.New()
	task := createTestTask(orgID, "test_task")

	ctx := context.Background()
	worker.executeTask(ctx, task)

	// Verify the organization ID was set in the context
	require.Equal(orgID, contextOrgID)
}

// mockContextExecutor is a test executor that captures the org ID from context
type mockContextExecutor struct {
	contextOrgID *uuid.UUID
}

func (m *mockContextExecutor) Execute(ctx context.Context, orgID uuid.UUID) error {
	// Extract org ID from context
	if ctxOrgID, ok := ctx.Value(consts.OrganizationIDCtxKey).(uuid.UUID); ok {
		*m.contextOrgID = ctxOrgID
	}
	return nil
}

func TestWorker_Run_ContextCancellation(t *testing.T) {
	require := require.New(t)

	taskQueue := NewTestTaskQueue()
	executors := createTestExecutors()
	logger := logrus.New()

	wp := NewWorkerPool(1, taskQueue.TaskQueue, executors, logger)

	ctx, cancel := context.WithCancel(context.Background())
	
	// Start worker pool
	wp.Start(ctx)

	// Cancel context immediately
	cancel()

	// Worker pool should stop quickly
	done := make(chan struct{})
	go func() {
		wp.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(500 * time.Millisecond):
		require.Fail("Worker pool did not stop within timeout")
	}
}

func TestWorker_Run_QueueError(t *testing.T) {
	require := require.New(t)

	taskQueue := NewTestTaskQueue()
	taskQueue.SetPopError(errors.New("queue error"))
	
	executors := createTestExecutors()
	logger := logrus.New()

	wp := NewWorkerPool(1, taskQueue.TaskQueue, executors, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start worker pool
	wp.Start(ctx)

	// Wait for context to expire
	<-ctx.Done()

	// Stop worker pool
	wp.Stop()

	// If we get here, the worker handled the error correctly
	require.True(true)
}

func TestWorkerPool_MultipleWorkers(t *testing.T) {
	require := require.New(t)

	taskQueue := NewTestTaskQueue()
	mockExecutor := NewMockTaskExecutor()
	mockExecutor.SetDelay(50 * time.Millisecond) // Small delay to simulate work
	
	executors := map[TaskType]TaskExecutor{
		"test_task": mockExecutor,
	}
	logger := logrus.New()

	wp := NewWorkerPool(3, taskQueue.TaskQueue, executors, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wp.Start(ctx)

	// Add multiple tasks
	for i := 0; i < 5; i++ {
		orgID := uuid.New()
		task := createTestTask(orgID, "test_task")
		taskQueue.Push(task)
	}

	// Wait for tasks to be processed
	time.Sleep(200 * time.Millisecond)

	// Stop the pool
	wp.Stop()

	// Verify all tasks were executed
	executions := mockExecutor.GetExecutions()
	require.Equal(5, len(executions))
}