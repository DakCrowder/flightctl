package tasks_runner

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewTaskQueue(t *testing.T) {
	require := require.New(t)

	tq := NewTaskQueue()

	require.NotNil(tq)
	require.NotNil(tq.tasks)
	require.Equal(0, len(tq.tasks))
	require.Equal(0, tq.heap.Len())
	require.NotNil(tq.cond)
	require.False(tq.closed)
}

func TestTaskQueue_Push_Basic(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	orgID := uuid.New()
	task := &Task{
		ID:          "test-task-1",
		OrgID:       orgID,
		Type:        TaskType("test"),
		NextRunTime: time.Now().Add(1 * time.Hour),
		Interval:    time.Hour,
	}

	tq.Push(task)

	require.Equal(1, len(tq.tasks))
	require.Equal(1, tq.heap.Len())
	require.Equal(task, tq.tasks["test-task-1"])
	require.Equal(task, tq.heap[0])
	require.Equal(0, task.index)
}

func TestTaskQueue_Push_DuplicateID(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	orgID := uuid.New()
	task1 := &Task{
		ID:          "test-task-1",
		OrgID:       orgID,
		Type:        TaskType("test"),
		NextRunTime: time.Now().Add(1 * time.Hour),
		Interval:    time.Hour,
	}

	task2 := &Task{
		ID:          "test-task-1",
		OrgID:       orgID,
		Type:        TaskType("test"),
		NextRunTime: time.Now().Add(2 * time.Hour),
		Interval:    time.Hour,
	}

	tq.Push(task1)
	tq.Push(task2)

	require.Equal(1, len(tq.tasks))
	require.Equal(1, tq.heap.Len())
	require.Equal(task2, tq.tasks["test-task-1"])
	require.Equal(task2, tq.heap[0])
	require.Equal(0, task2.index)
}

func TestTaskQueue_Push_ClosedQueue(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()
	tq.Close()

	orgID := uuid.New()
	task := &Task{
		ID:          "test-task-1",
		OrgID:       orgID,
		Type:        TaskType("test"),
		NextRunTime: time.Now().Add(1 * time.Hour),
		Interval:    time.Hour,
	}

	tq.Push(task)

	require.Equal(0, len(tq.tasks))
	require.Equal(0, tq.heap.Len())
}

func TestTaskQueue_Pop_ImmediateTask(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	orgID := uuid.New()
	task := &Task{
		ID:          "test-task-1",
		OrgID:       orgID,
		Type:        TaskType("test"),
		NextRunTime: time.Now().Add(-1 * time.Second), // Past time
		Interval:    time.Hour,
	}

	tq.Push(task)

	ctx := context.Background()
	poppedTask, err := tq.Pop(ctx)

	require.NoError(err)
	require.Equal(task, poppedTask)
	require.Equal(0, len(tq.tasks))
	require.Equal(0, tq.heap.Len())
}

func TestTaskQueue_Pop_ContextCancellation(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tq.Pop(ctx)

	require.Error(err)
	require.Equal(context.Canceled, err)
}

func TestTaskQueue_Pop_ClosedQueue(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()
	tq.Close()

	ctx := context.Background()
	_, err := tq.Pop(ctx)

	require.Error(err)
	require.Equal(context.Canceled, err)
}

func TestTaskQueue_Pop_FutureTask(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	orgID := uuid.New()
	task := &Task{
		ID:          "test-task-1",
		OrgID:       orgID,
		Type:        TaskType("test"),
		NextRunTime: time.Now().Add(100 * time.Millisecond),
		Interval:    time.Hour,
	}

	tq.Push(task)

	ctx := context.Background()
	start := time.Now()
	poppedTask, err := tq.Pop(ctx)
	elapsed := time.Since(start)

	require.NoError(err)
	require.Equal(task, poppedTask)
	require.GreaterOrEqual(elapsed, 100*time.Millisecond)
	require.Less(elapsed, 200*time.Millisecond)
	require.Equal(0, len(tq.tasks))
	require.Equal(0, tq.heap.Len())
}

func TestTaskQueue_Pop_TimeoutContext(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	orgID := uuid.New()
	task := &Task{
		ID:          "test-task-1",
		OrgID:       orgID,
		Type:        TaskType("test"),
		NextRunTime: time.Now().Add(1 * time.Hour),
		Interval:    time.Hour,
	}

	tq.Push(task)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := tq.Pop(ctx)

	require.Error(err)
	require.Equal(context.DeadlineExceeded, err)
}

func TestTaskQueue_RemoveOrganization(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	orgID1 := uuid.New()
	orgID2 := uuid.New()

	tasks := []*Task{
		{
			ID:          "task-1-org1",
			OrgID:       orgID1,
			Type:        TaskType("test"),
			NextRunTime: time.Now().Add(1 * time.Hour),
			Interval:    time.Hour,
		},
		{
			ID:          "task-2-org1",
			OrgID:       orgID1,
			Type:        TaskType("test"),
			NextRunTime: time.Now().Add(2 * time.Hour),
			Interval:    time.Hour,
		},
		{
			ID:          "task-1-org2",
			OrgID:       orgID2,
			Type:        TaskType("test"),
			NextRunTime: time.Now().Add(3 * time.Hour),
			Interval:    time.Hour,
		},
	}

	for _, task := range tasks {
		tq.Push(task)
	}

	require.Equal(3, len(tq.tasks))
	require.Equal(3, tq.heap.Len())

	tq.RemoveOrganization(orgID1)

	require.Equal(1, len(tq.tasks))
	require.Equal(1, tq.heap.Len())
	require.Equal(tasks[2], tq.tasks["task-1-org2"])
	require.Equal(tasks[2], tq.heap[0])
}

func TestTaskQueue_Close(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	var wg sync.WaitGroup
	results := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ctx := context.Background()
			_, err := tq.Pop(ctx)
			results[index] = err
		}(i)
	}

	time.Sleep(10 * time.Millisecond)
	tq.Close()

	wg.Wait()

	require.True(tq.closed)
	for _, err := range results {
		require.Equal(context.Canceled, err)
	}
}

func TestTaskHeap_Ordering(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	now := time.Now()
	tasks := []*Task{
		{
			ID:          "task-3",
			NextRunTime: now.Add(3 * time.Hour),
		},
		{
			ID:          "task-1",
			NextRunTime: now.Add(1 * time.Hour),
		},
		{
			ID:          "task-2",
			NextRunTime: now.Add(2 * time.Hour),
		},
	}

	for _, task := range tasks {
		tq.Push(task)
	}

	require.Equal(3, tq.heap.Len())
	require.Equal("task-1", tq.heap[0].ID)
	require.Equal(now.Add(1*time.Hour), tq.heap[0].NextRunTime)
}

func TestTaskQueue_ConcurrentOperations(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	var wg sync.WaitGroup
	numGoroutines := 10
	tasksPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			orgID := uuid.New()
			
			for j := 0; j < tasksPerGoroutine; j++ {
				task := &Task{
					ID:          fmt.Sprintf("task-%d-%d", goroutineID, j),
					OrgID:       orgID,
					Type:        TaskType("test"),
					NextRunTime: time.Now().Add(-1 * time.Second),
					Interval:    time.Hour,
				}
				tq.Push(task)
			}
		}(i)
	}

	wg.Wait()

	require.Equal(numGoroutines*tasksPerGoroutine, len(tq.tasks))
	require.Equal(numGoroutines*tasksPerGoroutine, tq.heap.Len())
}

func TestTaskQueue_ConcurrentPopPush(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	var pushWg, popWg sync.WaitGroup
	numPushers := 5
	numPoppers := 3
	tasksPerPusher := 20
	
	poppedTasks := make(chan *Task, numPushers*tasksPerPusher)
	
	for i := 0; i < numPushers; i++ {
		pushWg.Add(1)
		go func(pusherID int) {
			defer pushWg.Done()
			orgID := uuid.New()
			
			for j := 0; j < tasksPerPusher; j++ {
				task := &Task{
					ID:          fmt.Sprintf("task-%d-%d", pusherID, j),
					OrgID:       orgID,
					Type:        TaskType("test"),
					NextRunTime: time.Now().Add(-1 * time.Second),
					Interval:    time.Hour,
				}
				tq.Push(task)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	for i := 0; i < numPoppers; i++ {
		popWg.Add(1)
		go func() {
			defer popWg.Done()
			ctx := context.Background()
			
			for {
				task, err := tq.Pop(ctx)
				if err != nil {
					return
				}
				poppedTasks <- task
			}
		}()
	}

	pushWg.Wait()
	time.Sleep(100 * time.Millisecond)
	tq.Close()
	popWg.Wait()
	close(poppedTasks)

	poppedCount := 0
	for range poppedTasks {
		poppedCount++
	}

	require.Equal(numPushers*tasksPerPusher, poppedCount)
}

func TestTaskQueue_MultipleTasksOrdering(t *testing.T) {
	require := require.New(t)
	tq := NewTaskQueue()

	now := time.Now()
	expectedOrder := []string{"task-1", "task-2", "task-3", "task-4", "task-5"}
	
	tasks := []*Task{
		{ID: "task-3", NextRunTime: now.Add(-3 * time.Second)},
		{ID: "task-1", NextRunTime: now.Add(-5 * time.Second)},
		{ID: "task-5", NextRunTime: now.Add(-1 * time.Second)},
		{ID: "task-2", NextRunTime: now.Add(-4 * time.Second)},
		{ID: "task-4", NextRunTime: now.Add(-2 * time.Second)},
	}

	for _, task := range tasks {
		tq.Push(task)
	}

	ctx := context.Background()
	var actualOrder []string
	
	for i := 0; i < len(tasks); i++ {
		task, err := tq.Pop(ctx)
		require.NoError(err)
		actualOrder = append(actualOrder, task.ID)
	}

	require.Equal(expectedOrder, actualOrder)
}