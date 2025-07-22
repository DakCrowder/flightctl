package periodic

import (
	"time"

	"github.com/google/uuid"
)

type PeriodicTaskType string

const (
	PeriodicTaskTypeRepositoryTester PeriodicTaskType = "repository-tester"
)

type PeriodicTaskMetadata struct {
	TaskType PeriodicTaskType
	Interval time.Duration
}

type PeriodicTaskReference struct {
	Type  PeriodicTaskType
	OrgID uuid.UUID
}
