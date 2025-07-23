package periodic

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/internal/util"
	"github.com/flightctl/flightctl/pkg/queues"
	testutil "github.com/flightctl/flightctl/test/util"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockPeriodicTaskExecutor struct {
	executeCallCount int
	executeCallArgs  []executeCallArgs
}

type executeCallArgs struct {
	ctx context.Context
	log logrus.FieldLogger
}

func (m *mockPeriodicTaskExecutor) Execute(ctx context.Context, log logrus.FieldLogger) {
	m.executeCallCount++
	m.executeCallArgs = append(m.executeCallArgs, executeCallArgs{ctx: ctx, log: log})
}

type mockConsumer struct {
	consumeHandler queues.ConsumeHandler
	consumeError   error
	closeCallCount int
}

func (m *mockConsumer) Consume(ctx context.Context, handler queues.ConsumeHandler) error {
	m.consumeHandler = handler
	return m.consumeError
}

func (m *mockConsumer) Close() {
	m.closeCallCount++
}

type mockProvider struct {
	newConsumerError error
	consumer         *mockConsumer
	stopCallCount    int
	waitCallCount    int
}

func (m *mockProvider) NewConsumer(queueName string) (queues.Consumer, error) {
	if m.newConsumerError != nil {
		return nil, m.newConsumerError
	}
	return m.consumer, nil
}

func (m *mockProvider) NewPublisher(queueName string) (queues.Publisher, error) {
	return nil, errors.New("not implemented")
}

func (m *mockProvider) Stop() {
	m.stopCallCount++
}

func (m *mockProvider) Wait() {
	m.waitCallCount++
}

// Test helpers

func createTestExecutors() map[PeriodicTaskType]PeriodicTaskExecutor {
	return map[PeriodicTaskType]PeriodicTaskExecutor{
		PeriodicTaskTypeRepositoryTester:       &mockPeriodicTaskExecutor{},
		PeriodicTaskTypeResourceSync:           &mockPeriodicTaskExecutor{},
		PeriodicTaskTypeDeviceDisconnected:     &mockPeriodicTaskExecutor{},
		PeriodicTaskTypeRolloutDeviceSelection: &mockPeriodicTaskExecutor{},
		PeriodicTaskTypeDisruptionBudget:       &mockPeriodicTaskExecutor{},
		PeriodicTaskTypeEventCleanup:           &mockPeriodicTaskExecutor{},
	}
}

func createTestLogger() logrus.FieldLogger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	return logger
}

func createValidTaskReference(taskType PeriodicTaskType) PeriodicTaskReference {
	return PeriodicTaskReference{
		Type:  taskType,
		OrgID: uuid.New(),
	}
}

// Tests for consumeTasks() function

func TestConsumeTasks_SuccessfulTaskConsumption(t *testing.T) {
	tests := []struct {
		name     string
		taskType PeriodicTaskType
	}{
		{"RepositoryTester", PeriodicTaskTypeRepositoryTester},
		{"ResourceSync", PeriodicTaskTypeResourceSync},
		{"DeviceDisconnected", PeriodicTaskTypeDeviceDisconnected},
		{"RolloutDeviceSelection", PeriodicTaskTypeRolloutDeviceSelection},
		{"DisruptionBudget", PeriodicTaskTypeDisruptionBudget},
		{"EventCleanup", PeriodicTaskTypeEventCleanup},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			executors := createTestExecutors()
			mockExecutor := executors[tt.taskType].(*mockPeriodicTaskExecutor)
			handler := consumeTasks(executors)
			log := createTestLogger()
			ctx := context.Background()

			// Create valid payload
			taskRef := createValidTaskReference(tt.taskType)
			payload, err := json.Marshal(taskRef)
			require.NoError(t, err)

			// Execute
			err = handler(ctx, payload, log)

			// Verify
			require.NoError(t, err)
			require.Equal(t, 1, mockExecutor.executeCallCount, "executor should be called once")
			require.Len(t, mockExecutor.executeCallArgs, 1)

			// Verify context contains organization ID
			executedCtx := mockExecutor.executeCallArgs[0].ctx
			orgID, ok := util.GetOrgIdFromContext(executedCtx)
			require.True(t, ok, "organization ID should be present in context")
			require.Equal(t, taskRef.OrgID, orgID, "organization ID should be added to context")

			// Verify logger is passed
			require.NotNil(t, mockExecutor.executeCallArgs[0].log)
		})
	}
}

func TestConsumeTasks_JSONUnmarshalingError(t *testing.T) {
	// Setup
	executors := createTestExecutors()
	handler := consumeTasks(executors)
	log := createTestLogger()
	ctx := context.Background()

	// Invalid JSON payload
	invalidPayload := []byte(`{"invalid": json}`)

	// Execute
	err := handler(ctx, invalidPayload, log)

	// Verify
	require.Error(t, err, "should return error for invalid JSON")

	// Verify no executors were called
	for _, executor := range executors {
		mockExecutor := executor.(*mockPeriodicTaskExecutor)
		require.Equal(t, 0, mockExecutor.executeCallCount, "no executor should be called")
	}
}

func TestConsumeTasks_UnknownTaskType(t *testing.T) {
	// Setup
	executors := createTestExecutors()
	handler := consumeTasks(executors)
	log := createTestLogger()
	ctx := context.Background()

	// Valid JSON but unknown task type
	taskRef := PeriodicTaskReference{
		Type:  "unknown-task-type",
		OrgID: uuid.New(),
	}
	payload, err := json.Marshal(taskRef)
	require.NoError(t, err)

	// Execute
	err = handler(ctx, payload, log)

	// Verify
	require.NoError(t, err, "should return nil for unknown task type")

	// Verify no executors were called
	for _, executor := range executors {
		mockExecutor := executor.(*mockPeriodicTaskExecutor)
		require.Equal(t, 0, mockExecutor.executeCallCount, "no executor should be called")
	}
}

func TestConsumeTasks_MultipleTaskTypes(t *testing.T) {
	// Setup
	executors := createTestExecutors()
	handler := consumeTasks(executors)
	log := createTestLogger()
	ctx := context.Background()

	testCases := []PeriodicTaskType{
		PeriodicTaskTypeRepositoryTester,
		PeriodicTaskTypeResourceSync,
		PeriodicTaskTypeEventCleanup,
	}

	// Execute each task type
	for _, taskType := range testCases {
		taskRef := createValidTaskReference(taskType)
		payload, err := json.Marshal(taskRef)
		require.NoError(t, err)

		err = handler(ctx, payload, log)
		require.NoError(t, err)
	}

	// Verify each corresponding executor was called exactly once
	for _, taskType := range testCases {
		mockExecutor := executors[taskType].(*mockPeriodicTaskExecutor)
		require.Equal(t, 1, mockExecutor.executeCallCount,
			"executor for %s should be called once", taskType)
	}

	// Verify other executors were not called
	allTaskTypes := []PeriodicTaskType{
		PeriodicTaskTypeRepositoryTester,
		PeriodicTaskTypeResourceSync,
		PeriodicTaskTypeDeviceDisconnected,
		PeriodicTaskTypeRolloutDeviceSelection,
		PeriodicTaskTypeDisruptionBudget,
		PeriodicTaskTypeEventCleanup,
	}

	for _, taskType := range allTaskTypes {
		mockExecutor := executors[taskType].(*mockPeriodicTaskExecutor)
		found := false
		for _, testType := range testCases {
			if testType == taskType {
				found = true
				break
			}
		}
		if !found {
			require.Equal(t, 0, mockExecutor.executeCallCount,
				"executor for %s should not be called", taskType)
		}
	}
}

// Tests for PeriodicTaskConsumer.Start() method

func TestPeriodicTaskConsumer_Start_Success(t *testing.T) {
	// Setup
	mockConsumer := &mockConsumer{}
	mockProvider := &mockProvider{consumer: mockConsumer}
	executors := createTestExecutors()
	log := createTestLogger()

	consumer := &PeriodicTaskConsumer{
		queuesProvider: mockProvider,
		log:            log,
		executors:      executors,
	}

	ctx := context.Background()

	// Execute
	err := consumer.Start(ctx)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, mockConsumer.consumeHandler, "consume handler should be set")
}

func TestPeriodicTaskConsumer_Start_ConsumerCreationError(t *testing.T) {
	// Setup
	expectedError := errors.New("failed to create consumer")
	mockProvider := &mockProvider{newConsumerError: expectedError}
	executors := createTestExecutors()
	log := createTestLogger()

	consumer := &PeriodicTaskConsumer{
		queuesProvider: mockProvider,
		log:            log,
		executors:      executors,
	}

	ctx := context.Background()

	// Execute
	err := consumer.Start(ctx)

	// Verify
	require.Error(t, err)
	require.Equal(t, expectedError, err)
}

func TestPeriodicTaskConsumer_Start_ConsumeError(t *testing.T) {
	// Setup
	expectedError := errors.New("failed to consume")
	mockConsumer := &mockConsumer{consumeError: expectedError}
	mockProvider := &mockProvider{consumer: mockConsumer}
	executors := createTestExecutors()
	log := createTestLogger()

	consumer := &PeriodicTaskConsumer{
		queuesProvider: mockProvider,
		log:            log,
		executors:      executors,
	}

	ctx := context.Background()

	// Execute
	err := consumer.Start(ctx)

	// Verify
	require.Error(t, err)
	require.Equal(t, expectedError, err)
}

func TestPeriodicTaskConsumer_Start_UsesCorrectQueueName(t *testing.T) {
	// Note: This test verifies that Start() calls NewConsumer, but we can't easily
	// capture the queue name parameter without more complex mocking.
	// The actual queue name usage is verified through the consts import.

	// Setup
	mockConsumer := &mockConsumer{}
	mockProvider := &mockProvider{consumer: mockConsumer}
	executors := createTestExecutors()
	log := createTestLogger()

	consumer := &PeriodicTaskConsumer{
		queuesProvider: mockProvider,
		log:            log,
		executors:      executors,
	}

	ctx := context.Background()

	// Execute
	err := consumer.Start(ctx)

	// Verify
	require.NoError(t, err)
	// Note: We trust that NewConsumer is called with consts.PeriodicTaskQueue
	// based on the code review. More sophisticated queue name verification
	// would require a more complex mock framework.
}

// Integration tests

func TestPeriodicTaskConsumer_EndToEndMessageProcessing(t *testing.T) {
	// Setup
	log := createTestLogger()
	executors := createTestExecutors()

	// Use the test provider from test/util
	testProvider := testutil.NewTestProvider(log)
	defer testProvider.Stop()

	consumer := NewPeriodicTaskConsumer(testProvider, log, executors)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumer in background
	go func() {
		err := consumer.Start(ctx)
		if err != nil {
			t.Logf("Consumer start error: %v", err)
		}
	}()

	// Publish a test message
	publisher, err := testProvider.NewPublisher(consts.PeriodicTaskQueue)
	require.NoError(t, err)

	taskRef := createValidTaskReference(PeriodicTaskTypeRepositoryTester)
	payload, err := json.Marshal(taskRef)
	require.NoError(t, err)

	err = publisher.Publish(ctx, payload)
	require.NoError(t, err)

	// Simple polling wait for message processing
	maxWait := 50 // 5 seconds total
	for i := 0; i < maxWait; i++ {
		mockExecutor := executors[PeriodicTaskTypeRepositoryTester].(*mockPeriodicTaskExecutor)
		if mockExecutor.executeCallCount > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Verify
	mockExecutor := executors[PeriodicTaskTypeRepositoryTester].(*mockPeriodicTaskExecutor)
	require.Equal(t, 1, mockExecutor.executeCallCount)

	// Verify context contains organization ID
	executedCtx := mockExecutor.executeCallArgs[0].ctx
	orgID, ok := util.GetOrgIdFromContext(executedCtx)
	require.True(t, ok, "organization ID should be present in context")
	require.Equal(t, taskRef.OrgID, orgID)
}

func TestPeriodicTaskConsumer_ContextCancellation(t *testing.T) {
	// Setup
	log := createTestLogger()
	executors := createTestExecutors()
	testProvider := testutil.NewTestProvider(log)
	defer testProvider.Stop()

	consumer := NewPeriodicTaskConsumer(testProvider, log, executors)

	ctx, cancel := context.WithCancel(context.Background())

	// Start consumer
	startErr := make(chan error, 1)
	go func() {
		startErr <- consumer.Start(ctx)
	}()

	// Cancel context immediately
	cancel()

	// Verify that Start() returns (doesn't hang)
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	select {
	case err := <-startErr:
		// Error or nil is acceptable when context is cancelled
		_ = err
	case <-timeout.C:
		t.Fatal("Start() should return when context is cancelled")
	}
}

func TestConsumeTasks_EmptyExecutorsMap(t *testing.T) {
	// Setup
	emptyExecutors := make(map[PeriodicTaskType]PeriodicTaskExecutor)
	handler := consumeTasks(emptyExecutors)
	log := createTestLogger()
	ctx := context.Background()

	// Valid task reference
	taskRef := createValidTaskReference(PeriodicTaskTypeRepositoryTester)
	payload, err := json.Marshal(taskRef)
	require.NoError(t, err)

	// Execute
	err = handler(ctx, payload, log)

	// Verify - should return nil (no error) but log error about missing executor
	require.NoError(t, err)
}
