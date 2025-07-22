package periodic

import (
	"context"
	"encoding/json"

	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/pkg/queues"
	"github.com/sirupsen/logrus"
)

func consumeTasks() queues.ConsumeHandler {
	return func(ctx context.Context, payload []byte, log logrus.FieldLogger) error {
		var reference PeriodicTaskReference
		if err := json.Unmarshal(payload, &reference); err != nil {
			log.WithError(err).Error("failed to unmarshal consume payload")
			return err
		}

		switch reference.Type {
		case PeriodicTaskTypeRepositoryTester:
			log.Infof("received %s task for organization %s", reference.Type, reference.OrgID)
		case PeriodicTaskTypeResourceSync:
			log.Infof("received %s task for organization %s", reference.Type, reference.OrgID)
		case PeriodicTaskTypeDeviceDisconnected:
			log.Infof("received periodic task %s for organization %s", reference.Type, reference.OrgID)
		case PeriodicTaskTypeRolloutDeviceSelection:
			log.Infof("received periodic task %s for organization %s", reference.Type, reference.OrgID)
		case PeriodicTaskTypeDisruptionBudget:
			log.Infof("received periodic task %s for organization %s", reference.Type, reference.OrgID)
		case PeriodicTaskTypeEventCleanup:
			log.Infof("received periodic task %s for organization %s", reference.Type, reference.OrgID)
		}

		return nil
	}
}

type PeriodicTaskConsumer struct {
	queuesProvider queues.Provider
	log            logrus.FieldLogger
}

func NewPeriodicTaskConsumer(queuesProvider queues.Provider, log logrus.FieldLogger) *PeriodicTaskConsumer {
	return &PeriodicTaskConsumer{
		queuesProvider: queuesProvider,
		log:            log,
	}
}

func (c *PeriodicTaskConsumer) Start(ctx context.Context) error {
	consumer, err := c.queuesProvider.NewConsumer(consts.PeriodicTaskQueue)
	if err != nil {
		return err
	}

	if err = consumer.Consume(ctx, consumeTasks()); err != nil {
		return err
	}

	return nil
}
